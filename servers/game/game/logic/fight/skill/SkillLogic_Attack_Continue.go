package skill

import (
	"xianxia/servers/game/game/global"
)

/*
	Skill_Type_Attack_Continue     = int32(1001) //1001 连续攻击 单次伤害百分比+固定额外伤害
*/
type SkillLogic_Attack_Continue struct {
	SkillLogic_Base
}

func (sl *SkillLogic_Attack_Continue) getAttackCount(sCfg *global.SkillCfg) int32 {
	return sCfg.Param1
}

func (sl *SkillLogic_Attack_Continue) getAttackPer(sCfg *global.SkillCfg) int32 {
	return sCfg.Value1
}

func (sl *SkillLogic_Attack_Continue) getAttackAddValue(sCfg *global.SkillCfg) int32 {
	return sCfg.Param2
}

func (sl *SkillLogic_Attack_Continue) Logic(sCfg *global.SkillCfg, skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
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

	//设置单次伤害百分比
	attacker.SetFightProp(global.Creature_Prop_Two_BaseGain, sl.getAttackPer(sCfg))

	//设置额外伤害
	attacker.SetFightProp(global.Creature_Prop_Two_FAAdd, attacker.GetFightProp(global.Creature_Prop_Two_FAAdd)+sl.getAttackAddValue(sCfg))

	//获取攻击次数
	attackCount := sl.getAttackCount(sCfg)

	rItemArr := []global.IFightEventData{}
	for _, targeter := range targeters {
		for i := 1; i <= int(attackCount); i++ {
			rItem := global.ServerG.GetFightMgr().DoRoundAttack(attacker, targeter)
			if rItem != nil {
				rItemArr = append(rItemArr, rItem)
				rItems := targeter.BeAttacked()
				if rItems != nil {
					rItemArr = append(rItemArr, rItems...)
				}
			}
		}
	}

	resetV, _ := attacker.GetFighterSrc().GetProp(global.Creature_Prop_Two_BaseGain)
	attacker.SetFightProp(global.Creature_Prop_Two_BaseGain, resetV)

	attacker.SetFightProp(global.Creature_Prop_Two_FAAdd, attacker.GetFightProp(global.Creature_Prop_Two_FAAdd)-sl.getAttackAddValue(sCfg))

	return rItemArr
}
