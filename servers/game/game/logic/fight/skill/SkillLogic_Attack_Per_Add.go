package skill

import (
	"xianxia/servers/game/game/global"
)

/*
	Skill_Type_Attack_Per_Add      = int32(1007) //1007 百分多少的伤害+额外伤害
*/
type SkillLogic_Attack_Per_Add struct {
	SkillLogic_Base
}

func (sl *SkillLogic_Attack_Per_Add) getAttackPer(sCfg *global.SkillCfg) int32 {
	return sCfg.Param1
}

func (sl *SkillLogic_Attack_Per_Add) getAttackAddValue(sCfg *global.SkillCfg) int32 {
	return sCfg.Value1
}

func (sl *SkillLogic_Attack_Per_Add) Logic(sCfg *global.SkillCfg, skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
	if sCfg == nil {
		sCfg = sl.getSkillCfg(skillId)
		if sCfg == nil {
			return nil
		}
	}

	//获取目标
	targeters := SkillMgr.getSkillTargets(sCfg.TargetType, attacker, attackers, defenders)
	if targeters == nil {
		return nil
	}

	//获取攻击百分比
	attackPer := sl.getAttackPer(sCfg)
	attacker.SetFightProp(global.Creature_Prop_Two_BaseGain, attackPer)

	//获取额外伤害
	attackAddV := sl.getAttackAddValue(sCfg)
	attacker.SetFightProp(global.Creature_Prop_Two_FAAdd, attacker.GetFightProp(global.Creature_Prop_Two_FAAdd)+attackAddV)

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
	resetV, _ := attacker.GetFighterSrc().GetProp(global.Creature_Prop_Two_BaseGain)
	attacker.SetFightProp(global.Creature_Prop_Two_BaseGain, resetV)
	attacker.SetFightProp(global.Creature_Prop_Two_FAAdd, attacker.GetFightProp(global.Creature_Prop_Two_FAAdd)-attackAddV)

	return rItemArr
}
