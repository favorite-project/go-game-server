package object

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/name5566/leaf/log"
	"xianxia/common/dbengine"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/conf"
	"strconv"
	"strings"
	"xianxia/servers/game/msg"
)

func (p *player) GetPlayerSkill() {
	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_GET_SKILL, int64(0),"GET", fmt.Sprintf("skill:%d", p.dbId))
}

func (p *player) ReadSkillFromDB(ret *dbengine.CDBRet) (err error) {
	if ret.Err != nil {
		// TODO:打日志
		log.Error("player:%d ReadSkillFromDB error:%p", p.GetDBId(), ret.Err)
		return nil
	}

	values, err := redis.String(ret.Content, nil)
	if err == redis.ErrNil {
		p.initSkillData()
		return nil
	}

	p.initSkillData()
	err = json.Unmarshal([]byte(values), p.SkillDBData)
	if err != nil {
		log.Error("player:%d ReadSkillFromDB error:%p", p.GetDBId(), err)
		return nil
	}

	return nil
}

func (p *player) initSkillData() {
	p.SkillDBData = &global.SkillDBData{}
	p.SkillDBData.Equips = make(map[int32]*global.SkillDBItem)
	p.SkillDBData.Bags = make(map[int32]*global.SkillDBItem)
}

func (p *player) SaveSkill() {

	j, err := json.Marshal(p.SkillDBData)
	if err != nil {
		log.Error("player:%d SaveSkill Marshal Error:%p", p.GetDBId(), err)
		return
	}

	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SKILL, int64(0),"SET", fmt.Sprintf("skill:%d", p.dbId), j)
}

func (p *player) skillCheckCond(condStr string) bool {
	//檢查激活條件
	if len(condStr) > 0 {
		condStrArr := strings.Split(condStr, ";")
		for _, condItem := range condStrArr {
			condItemArr := strings.Split(condItem, "#")
			if len(condItemArr) != 2 {
				log.Error("skill cfg %s condtion error", condStr)
				return false
			}

			ctype, err := strconv.Atoi(condItemArr[0])
			if err != nil {
				log.Error("skill cfg %s strconv.Atoi(condItemArr[0]) error:%s", condStr)
				return false
			}

			value, err := strconv.Atoi(condItemArr[1])
			if err != nil {
				log.Error("skill cfg %sd strconv.Atoi(condItemArr[1]) error:%s", condStr)
				return false
			}

			if ctype == global.Skill_ActiveCond_Type_Level {
				playerLv, _ := p.GetProp(global.Player_Prop_Level)
				if value > int(playerLv) {
					return false
				}
			} else if ctype == global.Skill_ActiveCond_Type_Advance {
				adLv, _ := p.GetProp(global.Player_Prop_Advance_Level)
				if value > int(adLv) {
					return false
				}
			}
		}
	}

	return true
}

func (p *player)skill_study(recvData []byte) {
	if len(recvData) < 4 {
		return
	}

	skillId := int32(conf.RdWrEndian.Uint32(recvData))
	if _, ok := p.SkillDBData.Equips[skillId]; ok {
		return
	}

	if _, ok := p.SkillDBData.Bags[skillId]; ok {
		return
	}

	icfg := global.ServerG.GetConfigMgr().GetCfg("Skill", skillId)
	if icfg == nil {
		return
	}

	scfg := icfg.(*global.SkillCfg)
	if !scfg.CanPlayerStrud || scfg.Id != scfg.SrcSkillId {
		return
	}

	if !p.skillCheckCond(scfg.ActiveCond) {
		return
	}

	sItem := &global.SkillDBItem{
		SkillId: skillId,
		Pos: int32(len(p.SkillDBData.Bags)),
	}

	p.SkillDBData.Bags[skillId] = sItem

	m := &msg.GSCL_PlayerSkilLStudy{
		SkillItem:sItem,
	}

	p.conn.Send(m)
}

func (p *player) skill_levelUp(recvData []byte) {
	if len(recvData) < 4 {
		return
	}

	skillId := int32(conf.RdWrEndian.Uint32(recvData))
	bEquip := true
	var sItem *global.SkillDBItem
	var ok bool
	if sItem, ok = p.SkillDBData.Equips[skillId]; !ok {
		if sItem, ok = p.SkillDBData.Bags[skillId]; !ok {
			return
		}

		bEquip = false
	}

	icfg := global.ServerG.GetConfigMgr().GetCfg("Skill", skillId)
	if icfg == nil {
		return
	}

	scfg := icfg.(*global.SkillCfg)
	if scfg.Level == scfg.MaxLevel || scfg.NextLevel == 0 {
		return
	}

	nicfg := global.ServerG.GetConfigMgr().GetCfg("Skill", scfg.NextLevel)
	if nicfg == nil {
		return
	}

	nscfg := nicfg.(*global.SkillCfg)

	if !p.skillCheckCond(nscfg.ActiveCond) {
		return
	}

	//檢查升級條件
	pspoint, _ := p.GetProp(global.Player_Prop_SkillPoint)
	if scfg.UpLvCostPoint > pspoint {
		return
	}

	upItems := []struct{
		CfgId int32
		Num int32
	}{}

	if len(scfg.UpLvCostItems) > 0 {
		upItemStrArr := strings.Split(scfg.UpLvCostItems, ";")
		for _, upItem := range upItemStrArr {
			upItemArr := strings.Split(upItem, "#")
			if len(upItemArr) != 2 {
				log.Error("skill cfg id:%d upItemArr error", scfg.Id)
				return
			}

			cfgId, err := strconv.Atoi(upItemArr[0])
			if err != nil {
				log.Error("skill cfg id:%d strconv.Atoi(upItemArr[0]) error:%s", scfg.Id, err)
				return
			}

			itemNum, err := strconv.Atoi(upItemArr[1])
			if err != nil {
				log.Error("skill cfg id:%d strconv.Atoi(upItemArr[1]) error:%s", scfg.Id, err)
				return
			}

			//數量不足
			if p.GetBagItemNum(int32(cfgId)) < int32(itemNum) {
				return
			}

			upItems = append(upItems, struct{
				CfgId int32
				Num int32
			} {int32(cfgId), int32(itemNum)})
		}
	}

	//扣除技能點
	p.SetProp(global.Player_Prop_SkillPoint, -scfg.UpLvCostPoint, true)

	//刪除道具
	for _, upItem := range upItems {
		p.AddItem(upItem.CfgId, -upItem.Num, true, true)
	}

	//更新技能
	sItem.SkillId = nscfg.Id
	if bEquip {
		delete(p.SkillDBData.Equips, skillId)
		p.SkillDBData.Equips[sItem.SkillId] = sItem
	} else {
		delete(p.SkillDBData.Bags, skillId)
		p.SkillDBData.Bags[sItem.SkillId] = sItem
	}

	m := &msg.GSCL_PlayerSkillLevelUp{
		DelSkillId:scfg.Id,
		AddSItem:sItem,
	}

	p.conn.Send(m)
}

func (p *player) skill_changePos(recvData []byte) {
	if len(recvData) < 8 {
		return
	}

	skillId1 := int32(conf.RdWrEndian.Uint32(recvData))
	targetPos := int32(conf.RdWrEndian.Uint32(recvData[4:]))

	var sItem1 *global.SkillDBItem
	var ok bool
	if sItem1, ok = p.Equips[skillId1]; !ok {
		return
	}

	if sItem1.Pos == targetPos {
		return
	}

	var sItem2 *global.SkillDBItem
	for _, sItem := range p.SkillDBData.Equips {
		if sItem.Pos == targetPos {
			sItem2 = sItem
			break
		}
	}

	tmpPos := sItem1.Pos
	sItem1.Pos = targetPos
	if sItem2 != nil {
		sItem2.Pos = tmpPos
	}

	m := &msg.GSCL_PlayerSkilLChangePos{
		OldSItem:sItem2,
		NewSItem:sItem1,
	}

	p.conn.Send(m)
}

func (p *player) skill_equip(recvData []byte) {
	if len(recvData) < 8 {
		return
	}

	skillId := int32(conf.RdWrEndian.Uint32(recvData))
	skillPos := int32(conf.RdWrEndian.Uint32(recvData[4:]))
	if skillPos < 0 || skillPos >= global.Player_Max_Skill_Num {
		return
	}

	var sItem *global.SkillDBItem
	var ok bool
	if sItem, ok = p.SkillDBData.Bags[skillId]; !ok {
		return
	}

	var sReplaceItem *global.SkillDBItem
	for _, sBItem := range p.SkillDBData.Equips {
		if sBItem.Pos == skillPos {
			sReplaceItem = sBItem
			break
		}
	}

	tmpPos := sItem.Pos
	sItem.Pos = skillPos
	p.SkillDBData.Equips[sItem.SkillId] = sItem
	delete(p.SkillDBData.Bags, sItem.SkillId)
	if sReplaceItem != nil {
		delete(p.SkillDBData.Equips, sReplaceItem.SkillId)
		sReplaceItem.Pos = tmpPos
		p.SkillDBData.Bags[sReplaceItem.SkillId] = sReplaceItem
	} else {
		//如果skillPos 前面有空的地方 会默认修改skillPos至此
		for i:= int32(0); i< skillPos;i++ {
			found := false
			for _, sbItem := range p.SkillDBData.Equips {
				if sbItem.Pos == i {
					found = true
					break
				}
			}

			if !found {
				sItem.Pos = i
				break
			}
		}

		//重新整理背包,这个算法一定要和客户端对齐，不然会出现技能背包位置bug
		for _, sbItem := range p.SkillDBData.Bags {
			if sbItem.Pos >  tmpPos {
				sbItem.Pos -= 1
			}
		}
	}

	m := &msg.GSCL_PlayerSkillEquip{
		EquipSItem:sItem,
		BagSItem:sReplaceItem,
	}

	p.conn.Send(m)
}

func (p *player) skill_unequip(recvData []byte) {
	if len(recvData) < 4 {
		return
	}

	skillId := int32(conf.RdWrEndian.Uint32(recvData))

	var sItem *global.SkillDBItem
	if _, ok := p.SkillDBData.Equips[skillId]; !ok {
		return
	}

	sItem = p.SkillDBData.Equips[skillId]
	delete(p.SkillDBData.Equips, skillId)

	for _, sBItem := range p.SkillDBData.Equips {
		if sBItem.Pos > sItem.Pos {
			sBItem.Pos -= 1
			break
		}
	}

	sItem.Pos = int32(len(p.SkillDBData.Bags))
	p.SkillDBData.Bags[sItem.SkillId] = sItem
}