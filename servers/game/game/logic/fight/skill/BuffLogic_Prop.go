package skill

import (
	"github.com/name5566/leaf/log"
	"strconv"
	"strings"
	"xianxia/servers/game/game/global"
)

/*
增加属性值
*/
type BuffLogic_Prop struct {
	BuffLogic_Base
}

func (bl *BuffLogic_Prop) EffectNow(defender global.IFightObject, buffId int32) global.IFightEventData {
	if defender.IsDead() {
		return nil
	}

	cfg := bl.getBuffCfgById(buffId)
	if cfg == nil {
		return nil
	}

	propArr := strings.Split(cfg.Value, "#")
	if len(propArr) != 2 {
		log.Error("BuffLogic_Prop::EffectNow buffId:%d Value Split error", buffId)
		return nil
	}

	propId, err := strconv.Atoi(propArr[0])
	if err != nil {
		log.Error("BuffLogic_Prop::EffectNow buffId:%d Value Atoi(propArr[0]) error", buffId)
		return nil
	}

	pv, err := strconv.Atoi(propArr[1])
	if err != nil {
		log.Error("BuffLogic_Prop::EffectNow buffId:%d Value Atoi(propArr[1]) error", buffId)
		return nil
	}

	defender.SetFightProp(propId, defender.GetFightProp(propId)+int32(pv))
	return nil
}

func (bl *BuffLogic_Prop) Reset(defender global.IFightObject, buffId int32) global.IFightEventData {
	cfg := bl.getBuffCfgById(buffId)
	if cfg == nil {
		return nil
	}

	propArr := strings.Split(cfg.Value, "#")
	if len(propArr) != 2 {
		log.Error("BuffLogic_Prop::Reset buffId:%d Value Split error", buffId)
		return nil
	}

	propId, err := strconv.Atoi(propArr[0])
	if err != nil {
		log.Error("BuffLogic_Prop::Reset buffId:%d Value Atoi(propArr[0]) error", buffId)
		return nil
	}

	pv, err := strconv.Atoi(propArr[1])
	if err != nil {
		log.Error("BuffLogic_Prop::Reset buffId:%d Value Atoi(propArr[1]) error", buffId)
		return nil
	}

	defender.SetFightProp(propId, defender.GetFightProp(propId) - int32(pv))

	return nil
}
