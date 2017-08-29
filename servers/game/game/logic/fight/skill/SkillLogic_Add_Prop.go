package skill

import (
	"xianxia/servers/game/game/global"
)

/*
	Skill_Type_Add_Prop          = int32(1009) //1005 加属性，一般都是加血
*/
const (
	Skill_Add_Prop_Point = int32(1) + iota
	Skill_Add_Prop_Per
)

type SkillLogic_Add_Prop struct {
	SkillLogic_Base
}

func (sl *SkillLogic_Add_Prop) getAddPropId(sCfg *global.SkillCfg) int32 {
	return sCfg.Param1
}

func (sl *SkillLogic_Add_Prop) getAddPropType(sCfg *global.SkillCfg) int32 {
	return sCfg.Value1
}

func (sl *SkillLogic_Add_Prop) getAddValue(sCfg *global.SkillCfg) int32 {
	return sCfg.Param2
}

func (sl *SkillLogic_Add_Prop) Logic(sCfg *global.SkillCfg, skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
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

	//获取战斗属性
	propId := sl.getAddPropId(sCfg)
	addType := sl.getAddPropType(sCfg)
	addValue := sl.getAddValue(sCfg)

	rItemArr := []global.IFightEventData{}
	addedValue := addValue
	for _, targeter := range targeters {
		if addType != Skill_Add_Prop_Point {
			baseV, _ := targeter.GetFighterSrc().GetProp(int(propId))
			addedValue = int32(baseV * addValue / 1000)
		}

		if propId == int32(global.Creature_Prop_Two_Blood) {
			targeter.SetBlood(targeter.GetBlood() + addedValue)
		} else {
			targeter.SetFightProp(int(propId), targeter.GetFightProp(int(propId))+addedValue)
		}

		sItem := &global.FightEventData_Skill{
			FightEventData_Base: global.FightEventData_Base{
				EType: global.FIGHT_EVENT_SKILL_FRIEND,
				Pos:   targeter.GetPos(),
			},
			ChangeProps: make(map[int32]int32),
		}

		sItem.ChangeProps[propId] = addedValue
		rItemArr = append(rItemArr, sItem)
	}

	return rItemArr
}
