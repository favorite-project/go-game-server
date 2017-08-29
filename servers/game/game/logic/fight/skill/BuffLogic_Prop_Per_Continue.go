package skill

import (
	"github.com/name5566/leaf/log"
	"strconv"
	"strings"
	"xianxia/servers/game/game/global"
)

/*
按照固定百分比持续改变战斗属性
*/
type BuffLogic_Prop_Per_Continue struct {
	BuffLogic_Base
}

func (bl *BuffLogic_Prop_Per_Continue) EffectPerRound(defender global.IFightObject, buffId int32) global.IFightEventData {
	if defender.IsDead() {
		return nil
	}

	cfg := bl.getBuffCfgById(buffId)
	if cfg == nil {
		return nil
	}

	propArr := strings.Split(cfg.Value, "#")
	if len(propArr) != 2 {
		log.Error("BuffLogic_Prop_Per_Continue::EffectNow buffId:%d Value Split error", buffId)
		return nil
	}

	propId, err := strconv.Atoi(propArr[0])
	if err != nil {
		log.Error("BuffLogic_Prop_Per_Continue::EffectNow buffId:%d Value Atoi(propArr[0]) error", buffId)
		return nil
	}

	pv, err := strconv.Atoi(propArr[1])
	if err != nil {
		log.Error("BuffLogic_Prop_Per_Continue::EffectNow buffId:%d Value Atoi(propArr[1]) error", buffId)
		return nil
	}

	baseV, _ := defender.GetFighterSrc().GetProp(propId)
	addV := int32(baseV * int32(pv) / 1000)
	if propId == global.Creature_Prop_Two_Blood {
		defender.SetBlood(defender.GetBlood() + addV)

		eDataItem := &global.FightEventData_BuffEffect{
			FightEventData_Base: global.FightEventData_Base{
				EType: global.FIGHT_EVENT_BUFFEFFECT,
				Pos:   defender.GetPos(),
			},
			BuffId:      buffId,
			ChangeProps: make(map[int32]int32),
		}

		eDataItem.ChangeProps[int32(propId)] = addV
		return eDataItem
	} else {
		defender.SetFightProp(propId, defender.GetFightProp(propId)+addV)

		return nil
	}
}

func (bl *BuffLogic_Prop_Per_Continue) Reset(defender global.IFightObject, buffId int32) global.IFightEventData {
	cfg := bl.getBuffCfgById(buffId)
	if cfg == nil {
		return nil
	}

	propArr := strings.Split(cfg.Value, "#")
	if len(propArr) != 2 {
		log.Error("BuffLogic_Prop_Per_Continue::Reset buffId:%d Value Split error", buffId)
		return nil
	}

	propId, err := strconv.Atoi(propArr[0])
	if err != nil {
		log.Error("BuffLogic_Prop_Per_Continue::Reset buffId:%d Value Atoi(propArr[0]) error", buffId)
		return nil
	}

	pv, err := strconv.Atoi(propArr[1])
	if err != nil {
		log.Error("BuffLogic_Prop_Per_Continue::Reset buffId:%d Value Atoi(propArr[1]) error", buffId)
		return nil
	}

	if propId != global.Creature_Prop_Two_Blood {
		baseV, _ := defender.GetFighterSrc().GetProp(propId)
		addV := int32(baseV * int32(pv) / 1000)
		defender.SetFightProp(propId, defender.GetFightProp(propId) - addV)
	}

	return nil
}
