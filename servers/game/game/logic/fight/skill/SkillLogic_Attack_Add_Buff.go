package skill

import (
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/game/global"
)

/*
	Skill_Type_Attack_Add_Buff     = int32(1004) //1004 攻击加固定额外伤害+buff
*/
type SkillLogic_Attack_Add_Buff struct {
	SkillLogic_Base
}

func (sl *SkillLogic_Attack_Add_Buff) getAttackAddValue(sCfg *global.SkillCfg) int32 {
	return sCfg.Param1
}

func (sl *SkillLogic_Attack_Add_Buff) getBuffs(sCfg *global.SkillCfg) []*SkillBuffItem {

	configMgr := global.ServerG.GetConfigMgr()
	sBArr := []*SkillBuffItem{}
	//生成buff
	if sCfg.Value1 != 0 {
		icfg := configMgr.GetCfg("Buff", sCfg.Value1)
		if icfg == nil {
			log.Error("SkillLogic_Attack_Add_Buff::getBuffs buffId:%d cfg empty", sCfg.Value1)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Value1)
		if BuffLogic == nil {
			log.Error("SkillLogic_Attack_Add_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Value1)
			return nil
		}

		sBArr = append(sBArr, &SkillBuffItem{
			BuffCfg:   buffCfg,
			RandValue: sCfg.Param2,
			BuffLogic: BuffLogic,
		})
	}

	if sCfg.Value2 != 0 {
		icfg := configMgr.GetCfg("Buff", sCfg.Value2)
		if icfg == nil {
			log.Error("SkillLogic_Attack_Add_Buff::getBuffs buffId:%d cfg empty", sCfg.Value2)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Value2)
		if BuffLogic == nil {
			log.Error("SkillLogic_Attack_Add_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Value2)
			return nil
		}

		sBArr = append(sBArr, &SkillBuffItem{
			BuffCfg:   buffCfg,
			RandValue: sCfg.Param3,
			BuffLogic: BuffLogic,
		})
	}

	if sCfg.Value3 != 0 {
		icfg := configMgr.GetCfg("Buff", sCfg.Value3)
		if icfg == nil {
			log.Error("SkillLogic_Attack_Add_Buff::getBuffs buffId:%d cfg empty", sCfg.Value3)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Value3)
		if BuffLogic == nil {
			log.Error("SkillLogic_Attack_Add_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Value3)
			return nil
		}

		sBArr = append(sBArr, &SkillBuffItem{
			BuffCfg:   buffCfg,
			RandValue: sCfg.Param4,
			BuffLogic: BuffLogic,
		})
	}

	return sBArr
}

func (sl *SkillLogic_Attack_Add_Buff) Logic(sCfg *global.SkillCfg, skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
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

			//buff效果
			for _, sbItem := range sbArr {
				if sbItem.RandValue <= 0 || global.ServerG.GetRandSrc().Int31n(1000)+1 > sbItem.RandValue {
					continue
				}

				targeter.AddBuff(sbItem.BuffCfg)

				//addBuff
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
	attacker.SetFightProp(global.Creature_Prop_Two_FAAdd, attacker.GetFightProp(global.Creature_Prop_Two_FAAdd)-addV)

	return rItemArr
}
