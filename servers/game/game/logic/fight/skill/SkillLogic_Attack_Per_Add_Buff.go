package skill

import (
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/game/global"
)

/*
	Skill_Type_Attack_Per_Add_Buff = int32(1008) //1008 百分多少的伤害+额外伤害+buff
*/
type SkillLogic_Attack_Per_Add_Buff struct {
	SkillLogic_Base
}

func (sl *SkillLogic_Attack_Per_Add_Buff) getAttackPer(sCfg *global.SkillCfg) int32 {
	return sCfg.Param1
}

func (sl *SkillLogic_Attack_Per_Add_Buff) getAttackAddValue(sCfg *global.SkillCfg) int32 {
	return sCfg.Value1
}

func (sl *SkillLogic_Attack_Per_Add_Buff) getBuffs(sCfg *global.SkillCfg) []*SkillBuffItem {

	configMgr := global.ServerG.GetConfigMgr()
	sBArr := []*SkillBuffItem{}
	//生成buff
	if sCfg.Param2 != 0 {
		icfg := configMgr.GetCfg("Buff", sCfg.Param2)
		if icfg == nil {
			log.Error("SkillLogic_Attack_Per_Add_Buff::getBuffs buffId:%d cfg empty", sCfg.Param2)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Param2)
		if BuffLogic == nil {
			log.Error("SkillLogic_Attack_Per_Add_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Param2)
			return nil
		}

		sBArr = append(sBArr, &SkillBuffItem{
			BuffCfg:   buffCfg,
			RandValue: sCfg.Value2,
			BuffLogic: BuffLogic,
		})
	}

	if sCfg.Param3 != 0 {
		icfg := configMgr.GetCfg("Buff", sCfg.Param3)
		if icfg == nil {
			log.Error("SkillLogic_Attack_Per_Add_Buff::getBuffs buffId:%d cfg empty", sCfg.Param3)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Param3)
		if BuffLogic == nil {
			log.Error("SkillLogic_Attack_Per_Add_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Param3)
			return nil
		}

		sBArr = append(sBArr, &SkillBuffItem{
			BuffCfg:   buffCfg,
			RandValue: sCfg.Value3,
			BuffLogic: BuffLogic,
		})
	}

	if sCfg.Param4 != 0 {
		icfg := configMgr.GetCfg("Buff", sCfg.Param4)
		if icfg == nil {
			log.Error("SkillLogic_Attack_Per_Add_Buff::getBuffs buffId:%d cfg empty", sCfg.Param4)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Param4)
		if BuffLogic == nil {
			log.Error("SkillLogic_Attack_Per_Add_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Param4)
			return nil
		}

		sBArr = append(sBArr, &SkillBuffItem{
			BuffCfg:   buffCfg,
			RandValue: sCfg.Value4,
			BuffLogic: BuffLogic,
		})
	}

	return sBArr
}

func (sl *SkillLogic_Attack_Per_Add_Buff) Logic(sCfg *global.SkillCfg, skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
	if sCfg == nil {
		sCfg = sl.getSkillCfg(skillId)
		if sCfg == nil {
			return nil
		}
	}

	//查找buff
	sbArr := sl.getBuffs(sCfg)
	if sbArr == nil {
		return nil
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
			//buff效果
			for _, sbItem := range sbArr {
				if sbItem.RandValue <= 0 || global.ServerG.GetRandSrc().Int31n(1000)+1 > sbItem.RandValue {
					continue
				}
				targeter.AddBuff(sbItem.BuffCfg)
				//增加buff
				addFeBuffItem := &global.FightEventData_Buff{
					FightEventData_Base: global.FightEventData_Base{
						EType: global.FIGHT_EVENT_BUFF_ADD,
						Pos:   targeter.GetPos(),
					},
					BuffId: sbItem.BuffCfg.Id,
				}
				rItemArr = append(rItemArr, addFeBuffItem)

				enFeItem := sbItem.BuffLogic.EffectNow(targeter, sbItem.BuffCfg.Id)
				if enFeItem != nil {
					rItemArr = append(rItemArr, enFeItem)
				}
			}
		}
	}

	//给角色去掉额外攻击
	resetV, _ := attacker.GetFighterSrc().GetProp(global.Creature_Prop_Two_BaseGain)
	attacker.SetFightProp(global.Creature_Prop_Two_BaseGain, resetV)
	attacker.SetFightProp(global.Creature_Prop_Two_FAAdd, attacker.GetFightProp(global.Creature_Prop_Two_FAAdd)-attackAddV)

	return rItemArr
}
