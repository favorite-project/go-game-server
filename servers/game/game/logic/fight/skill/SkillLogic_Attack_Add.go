package skill

import (
	"xianxia/servers/game/game/global"
)

/*
	Skill_Type_Attack_Add          = int32(1003) //1003 攻击加固定额外伤害
*/
type SkillLogic_Attack_Add struct {
	SkillLogic_Base
}

func (sl *SkillLogic_Attack_Add) getAttackAddValue(sCfg *global.SkillCfg) int32 {
	return sCfg.Param1
}

func (sl *SkillLogic_Attack_Add) Logic(sCfg *global.SkillCfg, skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
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

	//给角色加上额外攻击
	addV := sl.getAttackAddValue(sCfg)
	attacker.SetFightProp(global.Creature_Prop_Two_FAAdd, attacker.GetFightProp(global.Creature_Prop_Two_FAAdd)+addV)
	rItemArr := []global.IFightEventData{}
	for _, targeter := range targeters {
		rItem := global.ServerG.GetFightMgr().DoRoundAttack(attacker, targeter)
		if rItem != nil {
			rItemArr = append(rItemArr, rItem)
			rItems := targeter.BeAttacked()
			if rItems != nil {
				rItemArr = append(rItemArr, rItems...)
			}
		}
	}

	//给角色去掉额外攻击
	attacker.SetFightProp(global.Creature_Prop_Two_FAAdd, attacker.GetFightProp(global.Creature_Prop_Two_FAAdd)-addV)

	return rItemArr
}
