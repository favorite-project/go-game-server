package skill

import (
	"xianxia/servers/game/game/global"
)

/*
	Skill_Type_Relive              = int32(1002) //1002 复活 回复血量百分比
*/
type SkillLogic_Relive struct {
	SkillLogic_Base
}

func (sl *SkillLogic_Relive) getReliveBloodPer(sCfg *global.SkillCfg) int32 {
	return sCfg.Param1
}

func (sl *SkillLogic_Relive) Logic(sCfg *global.SkillCfg, skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
	if sCfg == nil {
		sCfg = sl.getSkillCfg(skillId)
		if sCfg == nil {
			return nil
		}
	}

	//计算技能
	targeters := SkillMgr.getSkillTargets(sCfg.TargetType, attacker, attackers, defenders)
	if targeters == nil {
		return nil
	}

	bloodPer := sl.getReliveBloodPer(sCfg)
	rItemArr := []global.IFightEventData{}
	for _, targeter := range targeters {
		addBlood := int32(targeter.GetFightProp(global.Creature_Prop_Two_Blood) * bloodPer / 1000)
		targeter.SetBlood(addBlood)
		feItem := &global.FightEventData_Skill{
			FightEventData_Base: global.FightEventData_Base{
				EType: global.FIGHT_EVENT_SKILL_FRIEND,
				Pos:   targeter.GetPos(),
			},
			ChangeProps: make(map[int32]int32),
		}

		feItem.ChangeProps[int32(global.Creature_Prop_Two_Blood)] = addBlood

		rItemArr = append(rItemArr, feItem)
	}

	return rItemArr
}
