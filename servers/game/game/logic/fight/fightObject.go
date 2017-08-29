package fight

import (
	"bytes"
	"encoding/binary"
	_ "fmt"
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/conf"
	"xianxia/servers/game/game/global"
)

//战斗角色信息 需要具体实现
type fightObject struct {
	object     global.Creature
	pos        int8
	curBlood   int32
	fightProps [global.Creature_Prop_Two_Max - global.Creature_Prop_Two_Attack]int32
	SkillDatas []*struct {
		SkillId   int32
		LeftRound int8
	}

	BuffDatas []*struct {
		BuffId    int32
		LeftRound int8
	}

	master global.Player
}

func (fo *fightObject) GetType() byte {
	return fo.object.GetType()
}

//字节化
func (fo *fightObject) ToBytes() []byte {
	buf := new(bytes.Buffer)

	ctype := fo.object.GetType()
	binary.Write(buf, conf.RdWrEndian, ctype)

	if ctype == global.Creature_Type_Monster {
		binary.Write(buf, conf.RdWrEndian, fo.object.GetCfgId())
	} else if ctype == global.Creature_Type_Player {
		//先默认是自己
		if fo.master != nil {
			if p, ok :=fo.object.(global.Player); ok {
				binary.Write(buf, conf.RdWrEndian,p.GetDBId())
				if p.GetDBId() != fo.master.GetDBId() {
					//名字
					name := [global.NAME_MAX_LEN]byte{}
					copy(name[:cap(name)], p.GetName())
					binary.Write(buf, conf.RdWrEndian, name)

					//头像
					pic := [global.PIC_MAX_LEN]byte{}
					copy(pic[:cap(pic)], p.GetPic())
					binary.Write(buf, conf.RdWrEndian, pic)

					//等级
					lv, _ := p.GetProp(global.Player_Prop_Level)
					binary.Write(buf, conf.RdWrEndian, lv)

					//血量
					blood, _ := p.GetProp(global.Creature_Prop_Two_Blood)
					binary.Write(buf, conf.RdWrEndian, blood)
				}
			} else if op, ok :=fo.object.(global.OffLinePlayer); ok {
				binary.Write(buf, conf.RdWrEndian, op.GetDBId())
				if op.GetDBId() != fo.master.GetDBId() {
					//名字
					name := [global.NAME_MAX_LEN]byte{}
					copy(name[:cap(name)], op.GetName())
					binary.Write(buf, conf.RdWrEndian, name)

					//头像
					pic := [global.PIC_MAX_LEN]byte{}
					copy(pic[:cap(pic)], op.GetPic())
					binary.Write(buf, conf.RdWrEndian, pic)

					//等级
					lv, _ := op.GetProp(global.Player_Prop_Level)
					binary.Write(buf, conf.RdWrEndian, lv)

					//血量
					blood, _ := op.GetProp(global.Creature_Prop_Two_Blood)
					binary.Write(buf, conf.RdWrEndian, blood)
				}
			} else {
				log.Error("fightObject ToBytes fo.object type error")
			}
		}
	}

	return buf.Bytes()
}

func (fo *fightObject) GetPos() int8 {
	return fo.pos
}

func (fo *fightObject) IsDead() bool {
	return fo.curBlood <= 0
}

func (fo *fightObject) GetBlood() int32 {
	return fo.curBlood
}

func (fo *fightObject) SetBlood(value int32) {
	fo.curBlood = value
	if fo.curBlood < 0 {
		fo.curBlood = 0
	} else {
		maxBlood, _ := fo.object.GetProp(global.Creature_Prop_Two_Blood)
		if fo.curBlood > maxBlood {
			fo.curBlood = maxBlood
		}
	}

}

func (fo *fightObject) GetFighterSrc() global.Creature {
	return fo.object
}

func (fo *fightObject) AddBuff(cfg *global.BuffCfg) {
	if fo.IsDead() {
		return
	}

	if cfg == nil {
		return
	}

	if fo.BuffDatas == nil {
		fo.BuffDatas = []*struct {
			BuffId    int32
			LeftRound int8
		}{}
	}

	found := false
	for _, item := range fo.BuffDatas {
		if item.BuffId == cfg.Id { //叠加规则是覆盖 不是简单的相加
			item.LeftRound = cfg.Round
			found = true
		}
	}

	if !found {
		fo.BuffDatas = append(fo.BuffDatas, &struct {
			BuffId    int32
			LeftRound int8
		}{cfg.Id, cfg.Round})
	}
}

func (fo *fightObject) ClearBuff(buffId int32) {
	if fo.BuffDatas != nil {
		for _, item := range fo.BuffDatas {
			if item.BuffId == buffId { //叠加规则是覆盖 不是简单的相加
				item.LeftRound = 0
				break
			}
		}
	}
}

func (fo *fightObject) GetFightProp(propId int) int32 {
	return fo.fightProps[propId-global.Creature_Prop_Two_Attack]
}

func (fo *fightObject) SetFightProp(propId int, value int32) {
	fo.fightProps[propId-global.Creature_Prop_Two_Attack] = value
}

func (fo *fightObject) CanAttack() bool {
	if fo.IsDead() {
		return false
	}

	if fo.BuffDatas != nil {
		//检查是否有被定住之类的buff
		for _, item := range fo.BuffDatas {
			if item.LeftRound > 0 {
				bl := global.ServerG.GetSkillMgr().GetBuffLogic(item.BuffId)
				if bl != nil && !bl.CanAttack() {
					return false
				}
			}
		}
	}

	return true
}

func (fo *fightObject) CanBeAttacked() bool {
	if fo.IsDead() {
		return false
	}

	if fo.BuffDatas != nil {
		for _, item := range fo.BuffDatas {
			if item.LeftRound > 0 {
				bl := global.ServerG.GetSkillMgr().GetBuffLogic(item.BuffId)
				if bl != nil && !bl.CanBeAttacked() {
					return false
				}
			}
		}
	}

	return true
}

func (fo *fightObject) CanUseSkill() bool {
	if fo.IsDead() {
		return false
	}

	if fo.BuffDatas != nil {
		for _, item := range fo.BuffDatas {
			if item.LeftRound > 0 {
				bl := global.ServerG.GetSkillMgr().GetBuffLogic(item.BuffId)
				if bl != nil && !bl.CanUseSkill() {
					return false
				}
			}
		}
	}

	return true
}

func (fo fightObject) BeAttacked() []global.IFightEventData {
	feItemArr := []global.IFightEventData{}
	//检查遭受攻击时候打断的buff
	if fo.BuffDatas != nil {
		for _, item := range fo.BuffDatas {
			if item.LeftRound > 0 {
				bl := global.ServerG.GetSkillMgr().GetBuffLogic(item.BuffId)
				if bl != nil && bl.CanBeInterrupt() {
					item.LeftRound = 0

					//删除buff
					delFeBuffItem := &global.FightEventData_Buff{
						FightEventData_Base: global.FightEventData_Base{
							EType: global.FIGHT_EVENT_BUFF_DEL,
							Pos:   fo.GetPos(),
						},
						BuffId: item.BuffId,
					}

					feItemArr = append(feItemArr, delFeBuffItem)
				}
			}
		}
	}

	return feItemArr
}

func (fo *fightObject) Update(curRound int8) []global.IFightEventData {
	if fo.IsDead() {
		return nil
	}

	if fo.SkillDatas != nil {
		for _, item := range fo.SkillDatas {
			if item.LeftRound > int8(0) {
				item.LeftRound--
			}
		}
	}

	feItemArr := []global.IFightEventData{}
	if fo.BuffDatas != nil {
		for _, item := range fo.BuffDatas {
			if item.LeftRound > int8(0) {
				item.LeftRound--
				BuffLogic := global.ServerG.GetSkillMgr().GetBuffLogic(item.BuffId)
				if BuffLogic != nil {
					//计算buff的收益
					feItem := BuffLogic.EffectPerRound(fo, item.BuffId)
					if feItem != nil {
						feItemArr = append(feItemArr, feItem)
					}

					if item.LeftRound == 0 {
						//buff过期
						feItem = BuffLogic.Reset(fo, item.BuffId)
						if feItem != nil {
							feItemArr = append(feItemArr, feItem)
						}

						//删除buff
						delFeBuffItem := &global.FightEventData_Buff{
							FightEventData_Base: global.FightEventData_Base{
								EType: global.FIGHT_EVENT_BUFF_DEL,
								Pos:   fo.GetPos(),
							},
							BuffId: item.BuffId,
						}

						feItemArr = append(feItemArr, delFeBuffItem)
					}
				}
			}
		}
	}

	return feItemArr
}

func (fo *fightObject) GetNowSkill() int32 {
	skillId := int32(0)
	if fo.SkillDatas == nil {
		return skillId
	}

	for _, item := range fo.SkillDatas {
		if item.LeftRound == int8(0) {
			skillId = item.SkillId
			break
		}
	}

	return skillId
}

func (fo *fightObject) ResetNowSkill(skillId int32) {
	if fo.IsDead() {
		return
	}

	if skillId != 0 && fo.SkillDatas != nil {
		for _, item := range fo.SkillDatas {
			if item.SkillId == skillId {
				icfg := global.ServerG.GetConfigMgr().GetCfg("Skill", item.SkillId)
				if icfg == nil {
					log.Error("fightObject::initNowSkill get skillId:%d cfg empty", skillId)
					break
				}

				item.LeftRound = icfg.(*global.SkillCfg).CDRound
				break
			}
		}
	}
}

func (fo *fightObject) initFightProps() {
	for i := global.Creature_Prop_Two_Attack; i < global.Creature_Prop_Two_Max; i++ {
		fo.fightProps[i-global.Creature_Prop_Two_Attack], _ = fo.object.GetProp(i)
	}

	fo.curBlood, _ = fo.object.GetProp(global.Creature_Prop_Two_Blood)
}

func (fo *fightObject) initSkillData() {
	skillData := fo.object.GetSkillData()

	//测试

	if skillData != nil && skillData.Equips != nil && len(skillData.Equips) != 0 {
		//拉取csv
		confgMgr := global.ServerG.GetConfigMgr()

		SkillDatas := make([]*struct {
			SkillId   int32
			LeftRound int8
		}, global.Player_Max_Skill_Num, global.Player_Max_Skill_Num)

		for _, item := range skillData.Equips {
			icfg := confgMgr.GetCfg("Skill", item.SkillId)
			if icfg == nil {
				log.Error("fightObject::initSkillData get skillId:%d cfg empty", item.SkillId)
				continue
			}

			SkillDatas[item.Pos] = &struct {
				SkillId   int32
				LeftRound int8
			}{SkillId: item.SkillId, LeftRound: int8(0)}
		}

		fo.SkillDatas = []*struct {
			SkillId   int32
			LeftRound int8
		}{}
		for _, sItem := range SkillDatas {
			if sItem != nil {
				fo.SkillDatas = append(fo.SkillDatas, sItem)
			}
		}
	}

}
