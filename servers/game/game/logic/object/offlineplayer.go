package object

import (
	"xianxia/servers/game/game/global/db"
	"xianxia/servers/game/game/global"
	"encoding/json"
	"github.com/name5566/leaf/log"
	"strings"
	"strconv"
)

/*
	离线数据主要是拉取玩家战斗相关数据：
		1、属性
		2、装备
		3、技能
*/

type offlineplayer struct {
	creature
	dbId    int64
	props   PlayerProps

	*global.EquipDBData

	loadTime int32
}

func (p *offlineplayer) GetDBId() int64 {
	return p.dbId
}

func (p *offlineplayer) GetProp(index int) (int32, bool) {
	value, suc := p.creature.GetProp(index)
	if suc {
		return value, suc
	}

	if index >= global.Player_Prop_Max || index < global.Player_Prop_Level {
		return 0, false
	}

	return p.props[index-global.Player_Prop_Level], true
}

func (p *offlineplayer) SetProp(index int, value int32, bAdd bool) (int32, bool) {
	v, suc := p.creature.SetProp(index, value, bAdd)
	if suc {
		return v, suc
	}

	if index >= global.Player_Prop_Max || index < global.Player_Prop_Level {
		return 0, false
	}

	ov := p.props[index-global.Player_Prop_Level]

	if bAdd {
		p.props[index-global.Player_Prop_Level] += value
	} else {
		p.props[index-global.Player_Prop_Level] = value
	}

	//等级和经验药特殊判断
	if index == global.Player_Prop_Exp {
		for {
			level, _ := p.GetProp(global.Player_Prop_Level)
			bcfg := global.ServerG.GetConfigMgr().GetCfg("PlayerLevel", level)
			if bcfg == nil {
				return 0, false
			}

			cfg := bcfg.(*global.PlayerLevelCfg)
			if p.props[index-global.Player_Prop_Level] >= cfg.NeedExp { //满足升级经验
				bcfg = global.ServerG.GetConfigMgr().GetCfg("PlayerLevel", level+1)
				if bcfg == nil { //满级了
					p.props[index-global.Player_Prop_Level] = cfg.NeedExp
					break
				} else {
					p.SetProp(global.Player_Prop_Level, 1, true)
					p.SetProp(global.Player_Prop_SkillPoint, cfg.AddSkillPoint, true)
					p.props[index-global.Player_Prop_Level] -= cfg.NeedExp
				}
			} else {
				break
			}
		}
	} else if index == global.Player_Prop_Level {
		bcfg := global.ServerG.GetConfigMgr().GetCfg("PlayerLevel", p.props[index-global.Player_Prop_Level])
		if bcfg == nil { //等级不存在或者已经满级了
			p.props[index-global.Player_Prop_Level] = ov
			return 0, false
		}

		if p.props[index-global.Player_Prop_Level] != ov { //重新计算等级相关的属性
			bCsv := global.ServerG.GetConfigMgr().GetCsv("PlayerLevel")
			if bCsv != nil {
				ld := int32(1)
				begin := ov
				end := p.props[index-global.Player_Prop_Level]
				if p.props[index-global.Player_Prop_Level] < ov {
					ld = -ld
					begin = p.props[index-global.Player_Prop_Level]
					end = ov
				}

				for i := begin; i < end; i++ {
					bcfg := bCsv.Record(int(i))
					if bcfg == nil {
						p.props[index-global.Player_Prop_Level] = ov
						return 0, false
					}

					cfg := bcfg.(*global.PlayerLevelCfg)
					propArr := strings.Split(cfg.AddProps, ";")
					for _, propItemStr := range propArr {
						propItemArr := strings.Split(propItemStr, "+")
						if len(propItemArr) != 2 {
							p.props[index-global.Player_Prop_Level] = ov
							return 0, false
						}

						k, err := strconv.Atoi(propItemArr[0])
						if err != nil {
							p.props[index-global.Player_Prop_Level] = ov
							return 0, false
						}

						v, err := strconv.Atoi(propItemArr[1])
						if err != nil {
							p.props[index-global.Player_Prop_Level] = ov
							return 0, false
						}

						p.SetProp(k, int32(v)*ld, true)
					}
				}
			}
		}
	} else if index == global.Player_Prop_VipExp {
		//检查是否达成vip升级条件
		for {
			vipLevel, _ := p.GetProp(global.Player_Prop_VipLevel)
			cfg := global.ServerG.GetConfigMgr().GetCfg("Vip", vipLevel + 1)
			if cfg == nil {
				break
			}

			icfg := cfg.(*global.VipCfg)
			if p.props[index - global.Player_Prop_Level] >= icfg.Recharge {
				p.SetProp(global.Player_Prop_VipLevel, vipLevel + 1, false)
			} else {
				break
			}
		}
	} else if index == global.Player_Prop_Advance_Level {
		oicfg := global.ServerG.GetConfigMgr().GetCfg("Advance", ov)
		oparr := make(map[int]int32)
		if oicfg != nil {
			oparr = getPropsArr(oicfg.(*global.AdvanceCfg).AddProps)
		}

		icfg := global.ServerG.GetConfigMgr().GetCfg("Advance", p.props[index-global.Player_Prop_Level])
		if icfg != nil {
			parr := getPropsArr(icfg.(*global.AdvanceCfg).AddProps)
			for ppid, ppv := range parr {
				if oppv, ok := oparr[ppid]; ok {
					ppv -= oppv
				}

				p.SetProp(ppid, ppv, true)
			}
		}
	}

	return p.props[index-global.Player_Prop_Level], true
}

func (p *offlineplayer) setProps(dbData *db.DB_PLayer_Props) {
	if dbData == nil {
		return
	}

	p.dbId = dbData.DBId

	p.initBase()

	//通过等级计算一级和二级属性
	p.SetProp(global.Player_Prop_Level, dbData.Level, false)

	//设置vip经验，并计算出等级
	p.SetProp(global.Player_Prop_VipExp, dbData.VipExp, false)

	p.SetProp(global.Player_Prop_Advance_Level, dbData.AdvanceLevel, false)

	//基本数据
	p.creature.cType = global.Creature_Type_Player


	//角色属性
	p.props[global.Player_Prop_FightVal-global.Player_Prop_Level] = dbData.FightVal

	copy(p.creature.pic[:len(p.creature.pic)], dbData.Pic)
	copy(p.creature.name[:len(p.creature.name)], dbData.Name)

	return
}

func (p *offlineplayer) updateEquipmentProps(eItem *global.ItemDBData, unEquip bool) {
	if eItem == nil {
		return
	}

	propsMap := &global.EquipDBItemData{}
	err := json.Unmarshal([]byte(eItem.Data), propsMap)
	if err != nil {
		log.Error("offlineplayer:%d updateEquipmentProps error:%p", p.GetDBId(), err)
		return
	}

	per := int32(1)
	if unEquip {
		per = -per
	}

	for pid, pv := range propsMap.BData {
		p.SetProp(int(pid), pv*per, true)
	}

	for pid, pv := range propsMap.OData {
		p.SetProp(int(pid), pv*per, true)
	}
}


func (p *offlineplayer) setEquips(dbData *global.EquipDBData) {
	if dbData == nil {
		return
	}

	p.EquipDBData = dbData
	// 解析装备,筛出穿在身上的装备
	for _, eItem := range p.EquipDBData.EquipData {
		p.updateEquipmentProps(eItem, false)
	}

	p.CaculateFightVal()
}

func (p *offlineplayer) setSkills(dbData *global.SkillDBData) {
	p.SkillDBData = dbData
}

func (p *offlineplayer) CaculateFightVal()  {
	fightVal, _ := p.GetProp(global.Player_Prop_FightVal)

	var newFightVal int32
	for i:=global.Creature_Prop_Two_Attack; i <= global.Creature_Prop_Two_Tenacity;i++ {
		v, _ := p.GetProp(i)
		newFightVal += v * fightValTable(int32(i))
	}

	if newFightVal != fightVal {
		p.SetProp(global.Player_Prop_FightVal, newFightVal, false)
	}
}
