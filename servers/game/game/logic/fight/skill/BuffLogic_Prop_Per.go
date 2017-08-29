package skill

import (
	"github.com/name5566/leaf/log"
	"strconv"
	"strings"
	"xianxia/servers/game/game/global"
)

/*
增加属性百分比
*/
type BuffLogic_Prop_Per struct {
	BuffLogic_Base
}

func (bl *BuffLogic_Prop_Per) EffectNow(defender global.IFightObject, buffId int32) global.IFightEventData {
	if defender.IsDead() {
		return nil
	}

	cfg := bl.getBuffCfgById(buffId)
	if cfg == nil {
		return nil
	}

	propArr := strings.Split(cfg.Value, "#")
	if len(propArr) != 2 {
		log.Error("BuffLogic_Prop_Per::EffectNow buffId:%d Value Split error", buffId)
		return nil
	}

	propId, err := strconv.Atoi(propArr[0])
	if err != nil {
		log.Error("BuffLogic_Prop_Per::EffectNow buffId:%d Value Atoi(propArr[0]) error", buffId)
		return nil
	}

	pv, err := strconv.Atoi(propArr[1])
	if err != nil {
		log.Error("BuffLogic_Prop_Per::EffectNow buffId:%d Value Atoi(propArr[1]) error", buffId)
		return nil
	}

	//基于基础属性加成
	baseV,_ := defender.GetFighterSrc().GetProp(propId)
	addV := int32(baseV * int32(pv)/1000)
	defender.SetFightProp(propId, defender.GetFightProp(propId) + addV)
	return nil
}

func (bl *BuffLogic_Prop_Per) Reset(defender global.IFightObject, buffId int32) global.IFightEventData {
	cfg := bl.getBuffCfgById(buffId)
	if cfg == nil {
		return nil
	}

	propArr := strings.Split(cfg.Value, "#")
	if len(propArr) != 2 {
		log.Error("BuffLogic_Prop_Per::Reset buffId:%d Value Split error", buffId)
		return nil
	}

	propId, err := strconv.Atoi(propArr[0])
	if err != nil {
		log.Error("BuffLogic_Prop_Per::Reset buffId:%d Value Atoi(propArr[0]) error", buffId)
		return nil
	}

	pv, err := strconv.Atoi(propArr[1])
	if err != nil {
		log.Error("BuffLogic_Prop_Per::Reset buffId:%d Value Atoi(propArr[1]) error", buffId)
		return nil
	}

	baseV,_ := defender.GetFighterSrc().GetProp(propId)
	addV := int32(baseV * int32(pv)/1000)
	defender.SetFightProp(propId, defender.GetFightProp(propId) - addV)
	return nil
}
