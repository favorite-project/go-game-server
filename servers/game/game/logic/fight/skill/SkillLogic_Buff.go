package skill

import (
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/game/global"
)

/*
	Skill_Type_Buff                = int32(101)  //101 buff技能
*/
type SkillLogic_Buff struct {
	SkillLogic_Base
}

func (sl *SkillLogic_Buff) getBuffs(sCfg *global.SkillCfg) []*SkillBuffItem {

	configMgr := global.ServerG.GetConfigMgr()
	sBArr := []*SkillBuffItem{}
	//生成buff
	if sCfg.Param1 != 0 {
		icfg := configMgr.GetCfg("Buff", sCfg.Param1)
		if icfg == nil {
			log.Error("SkillLogic_Buff::getBuffs buffId:%d cfg empty", sCfg.Param1)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Param1)
		if BuffLogic == nil {
			log.Error("SkillLogic_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Param1)
			return nil
		}

		sBArr = append(sBArr, &SkillBuffItem{
			BuffCfg:   buffCfg,
			RandValue: sCfg.Value1,
			BuffLogic: BuffLogic,
		})
	}

	if sCfg.Param2 != 0 {
		icfg := configMgr.GetCfg("Buff", sCfg.Param2)
		if icfg == nil {
			log.Error("SkillLogic_Buff::getBuffs buffId:%d cfg empty", sCfg.Param2)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Param2)
		if BuffLogic == nil {
			log.Error("SkillLogic_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Param2)
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
			log.Error("SkillLogic_Buff::getBuffs buffId:%d cfg empty", sCfg.Param3)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Param3)
		if BuffLogic == nil {
			log.Error("SkillLogic_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Param3)
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
			log.Error("SkillLogic_Buff::getBuffs buffId:%d cfg empty", sCfg.Param4)
			return nil
		}

		buffCfg := icfg.(*global.BuffCfg)
		BuffLogic := SkillMgr.GetBuffLogic(sCfg.Param4)
		if BuffLogic == nil {
			log.Error("SkillLogic_Buff::getBuffs buffId:%d GetBuffLogic nil", sCfg.Param4)
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

func (sl *SkillLogic_Buff) Logic(sCfg *global.SkillCfg, skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
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

	//计算buff
	sbArr := sl.getBuffs(sCfg)
	if sbArr == nil {
		return nil
	}

	rItemArr := []global.IFightEventData{}
	//添加buff
	for _, targeter := range targeters {
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

	return rItemArr
}
