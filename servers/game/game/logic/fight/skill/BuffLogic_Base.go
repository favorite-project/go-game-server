package skill

import (
	"xianxia/servers/game/game/global"
)

type BuffLogic_Base struct {
	Type int16
}

func (bl *BuffLogic_Base) GetType() int16 {
	return bl.Type
}

func (bl *BuffLogic_Base) CanAttack() bool {
	return true
}

func (bl *BuffLogic_Base) CanUseSkill() bool {
	return true
}

func (bl *BuffLogic_Base) CanBeAttacked() bool {
	return true
}

func (bl *BuffLogic_Base) EffectPerRound(defender global.IFightObject, buffId int32) global.IFightEventData {
	return nil
}

func (bl *BuffLogic_Base) EffectNow(defender global.IFightObject, buffId int32) global.IFightEventData {
	return nil
}

func (bl *BuffLogic_Base) Reset(defender global.IFightObject, buffId int32) global.IFightEventData {
	return nil
}

func (bl *BuffLogic_Base) CanBeInterrupt() bool {
	return false
}

func (bl *BuffLogic_Base) getBuffCfgById(buffId int32) *global.BuffCfg {
	icfg := global.ServerG.GetConfigMgr().GetCfg("Buff", buffId)
	if icfg == nil {
		return nil
	}

	cfg := icfg.(*global.BuffCfg)
	return cfg
}
