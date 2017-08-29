package skill

import (
	_ "fmt"
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/game/global"
)

type SkillDataItem struct {
	Type  int16
	Value string
}
type CSkillMgr struct {
	BuffLogics  map[int16]global.BuffLogic
	SkillLogics map[int32]global.SkillLogic
}

/*
	Skill_Type_Buff                = int32(101)  //101 buff技能
	Skill_Type_Attack_Continue     = int32(1001) //1001 连续攻击 单次伤害百分比+固定额外伤害
	Skill_Type_Relive              = int32(1002) //1002 复活 回复血量百分比
	Skill_Type_Attack_Add          = int32(1003) //1003 攻击加固定额外伤害
	Skill_Type_Attack_Add_Buff     = int32(1004) //1004 攻击加固定额外伤害+buff
	Skill_Type_Attack_Per          = int32(1005) //1005 百分多少的伤害
	Skill_Type_Attack_Per_Buff     = int32(1006) //1006 百分多少的伤害+buff
	Skill_Type_Attack_Per_Add      = int32(1007) //1007 百分多少的伤害+额外伤害
	Skill_Type_Attack_Per_Add_Buff = int32(1008) //1008 百分多少的伤害+额外伤害+buff
	Skill_Type_Add_Prop            = int32(1009) //1009回血回蓝加永久加属性等
*/
var SkillMgr *CSkillMgr

func init() {
	SkillMgr = &CSkillMgr{
		BuffLogics:  make(map[int16]global.BuffLogic),
		SkillLogics: make(map[int32]global.SkillLogic),
	}

	//注册所有buff逻辑
	SkillMgr.BuffLogics[global.Buff_Type_Dizzy] = &BuffLogic_Dizzy{BuffLogic_Base: BuffLogic_Base{global.Buff_Type_Dizzy}}
	SkillMgr.BuffLogics[global.Buff_Type_Sleep] = &BuffLogic_Sleep{BuffLogic_Base: BuffLogic_Base{global.Buff_Type_Sleep}}
	SkillMgr.BuffLogics[global.Buff_Type_SkillDisable] = &BuffLogic_SkillDisable{BuffLogic_Base: BuffLogic_Base{global.Buff_Type_SkillDisable}}
	SkillMgr.BuffLogics[global.Buff_Type_Prop] = &BuffLogic_Prop{BuffLogic_Base: BuffLogic_Base{global.Buff_Type_Prop}}
	SkillMgr.BuffLogics[global.Buff_Type_Prop_Per] = &BuffLogic_Prop_Per{BuffLogic_Base: BuffLogic_Base{global.Buff_Type_Prop_Per}}
	SkillMgr.BuffLogics[global.Buff_Type_Invincible] = &BuffLogic_Invincible{BuffLogic_Base: BuffLogic_Base{global.Buff_Type_Invincible}}
	SkillMgr.BuffLogics[global.Buff_Type_Prop_Continue] = &BuffLogic_Prop_Continue{BuffLogic_Base: BuffLogic_Base{global.Buff_Type_Prop_Continue}}
	SkillMgr.BuffLogics[global.Buff_Type_Prop_Per_Continue] = &BuffLogic_Prop_Per_Continue{BuffLogic_Base: BuffLogic_Base{global.Buff_Type_Prop_Per_Continue}}

	//注册所有skill逻辑
	SkillMgr.SkillLogics[global.Skill_Type_Buff] = &SkillLogic_Buff{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Buff}}
	SkillMgr.SkillLogics[global.Skill_Type_Attack_Continue] = &SkillLogic_Attack_Continue{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Attack_Continue}}
	SkillMgr.SkillLogics[global.Skill_Type_Relive] = &SkillLogic_Relive{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Relive}}
	SkillMgr.SkillLogics[global.Skill_Type_Attack_Add] = &SkillLogic_Attack_Add{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Attack_Add}}
	SkillMgr.SkillLogics[global.Skill_Type_Attack_Add_Buff] = &SkillLogic_Attack_Add_Buff{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Attack_Add_Buff}}
	SkillMgr.SkillLogics[global.Skill_Type_Attack_Per] = &SkillLogic_Attack_Per{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Attack_Per}}
	SkillMgr.SkillLogics[global.Skill_Type_Attack_Per_Buff] = &SkillLogic_Attack_Per_Buff{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Attack_Per_Buff}}
	SkillMgr.SkillLogics[global.Skill_Type_Attack_Per_Add] = &SkillLogic_Attack_Per_Add{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Attack_Per_Add}}
	SkillMgr.SkillLogics[global.Skill_Type_Attack_Per_Add_Buff] = &SkillLogic_Attack_Per_Add_Buff{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Attack_Per_Add_Buff}}
	SkillMgr.SkillLogics[global.Skill_Type_Add_Prop] = &SkillLogic_Add_Prop{SkillLogic_Base: SkillLogic_Base{global.Skill_Type_Add_Prop}}
}

func (sm *CSkillMgr) GetBuffLogic(buffId int32) global.BuffLogic {
	icfg := global.ServerG.GetConfigMgr().GetCfg("Buff", buffId)
	if icfg == nil {
		return nil
	}

	if bl, ok := sm.BuffLogics[icfg.(*global.BuffCfg).Type]; ok {
		return bl
	}

	return nil
}

func (sm *CSkillMgr) GetSkillLogic(skillId int32) global.SkillLogic {
	icfg := global.ServerG.GetConfigMgr().GetCfg("Skill", skillId)
	if icfg == nil {
		return nil
	}

	if bl, ok := sm.SkillLogics[icfg.(*global.SkillCfg).RType]; ok {
		return bl
	}

	return nil
}

func (sm *CSkillMgr) Logic(skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
	if attacker == nil || attackers == nil || len(attackers) == 0 || defenders == nil || len(defenders) == 0 {
		return nil
	}

	configMgr := global.ServerG.GetConfigMgr()
	var sCfg *global.SkillCfg
	icfg := configMgr.GetCfg("Skill", skillId)
	if icfg == nil {
		sCfg = nil
	} else {
		sCfg = icfg.(*global.SkillCfg)
	}

	if sCfg == nil { //未使用技能，就是普攻
		targeters := sm.getSkillTargets(global.Skill_Target_Type_Pos1, attacker, attackers, defenders)
		if targeters == nil {
			//log.Error("CSkillMgr::Logic getSkillTargets nil")
			return nil
		}

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

		return rItemArr
	} else {
		SkillLogic, ok := sm.SkillLogics[sCfg.RType]
		if !ok {
			log.Error("CSkillMgr::Logic SkillId:%d GetSkillLogic nil", skillId)
			return nil
		}

		return SkillLogic.Logic(sCfg, skillId, attacker, attackers, defenders)
	}
}

func (sm *CSkillMgr) getSkillTargets(sTargetType int16, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightObject {
	/*
		//buff作用对象（位置）[敌方]
		Skill_Target_Type_Pos1       = int16(1) + iota //1 对方单体，按照顺序攻击
		Skill_Target_Type_Pos2                         //2 对方2号位，如果没有找到就按照Pos3，Pos1的顺序找
		Skill_Target_Type_Pos3                         //3 对方3号位，如果没有找到就按照Pos2,Pos1的顺序找
		Skill_Target_Type_Rand1                        //4 对方列表中随机一个
		Skill_Target_Type_Rand2                        //5 对方列表中随机两个
		Skill_Target_Type_All                          //6 全部攻击
		Skill_Target_Type_BloodMin_1                   //7 对方生命最少的目标
		Skill_Target_Type_BloodMax_1                   //8 对方生命最多的目标
		Skill_Target_Type_Enemy_Max


		//buff作用对象（位置）[友方]
		Skill_Target_Type_Self_1          = int16(100) + iota //100 自己
		Skill_Target_Type_Self_Rand1                          //101 己方列表中随机一个
		Skill_Target_Type_Self_Rand2                          //102 己方列表中随机两个
		Skill_Target_Type_Self_All                            //103 己方全部
		Skill_Target_Type_Self_BloodMin_1                     //104 己方血量最少的
		Skill_Target_Type_Self_BloodMax_1                     //105 己方血量最多的
		Skill_Target_Type_Self_Dead_1                         //106 己方已死亡中随机一个
	*/

	bInAttacker := false
	for _, targeter := range attackers {
		if attacker == targeter {
			bInAttacker = true
			break
		}
	}

	targeters := attackers
	if (bInAttacker && sTargetType < global.Skill_Target_Type_Enemy_Max) || (!bInAttacker && sTargetType >= global.Skill_Target_Type_Enemy_Max) {
		targeters = defenders
	}

	targetersNum := int16(len(targeters))
	switch sTargetType {
	case global.Skill_Target_Type_Pos1,
		global.Skill_Target_Type_Pos2,
		global.Skill_Target_Type_Pos3:
		beginIndex := sTargetType - global.Skill_Target_Type_Pos1
		if beginIndex >= targetersNum {
			beginIndex = 0
		}

		endIndex := beginIndex - 1
		if beginIndex == 0 {
			endIndex = targetersNum - 1
		}

		for i := beginIndex; i < targetersNum+beginIndex; i++ {
			index := i
			if index >= targetersNum {
				index -= targetersNum
			}

			if targeters[index].CanBeAttacked() {
				return []global.IFightObject{targeters[index]}
			}

			if index == endIndex {
				break
			}
		}

		return nil
	case global.Skill_Target_Type_Rand1,
		global.Skill_Target_Type_Rand2,
		global.Skill_Target_Type_All:
		needNum := sTargetType - global.Skill_Target_Type_Rand1 + 1
		needTargeters := []global.IFightObject{}
		srcTargets := append([]global.IFightObject{}, targeters...)
		for {
			if needNum == 0 || len(srcTargets) == 0 {
				break
			}

			index := global.ServerG.GetRandSrc().Intn(len(srcTargets))
			if srcTargets[index].CanBeAttacked() {
				needTargeters = append(needTargeters, srcTargets[index])
				needNum--
			}

			srcTargets = append(srcTargets[:index], srcTargets[index+1:]...)
		}

		return needTargeters
	case global.Skill_Target_Type_BloodMin_1:
		needTargeters := []global.IFightObject{}
		for _, targeter := range targeters {
			if targeter.CanBeAttacked() {
				if len(needTargeters) == 0 {
					needTargeters = append(needTargeters, targeter)
				} else {
					if needTargeters[0].GetBlood() > targeter.GetBlood() {
						needTargeters[0] = targeter
					}
				}
			}
		}

		return needTargeters
	case global.Skill_Target_Type_BloodMax_1:
		needTargeters := []global.IFightObject{}
		for _, targeter := range targeters {
			if targeter.CanBeAttacked() {
				if len(needTargeters) == 0 {
					needTargeters = append(needTargeters, targeter)
				} else {
					if needTargeters[0].GetBlood() < targeter.GetBlood() {
						needTargeters[0] = targeter
					}
				}
			}
		}

		return needTargeters

	case global.Skill_Target_Type_Self_1:
		return []global.IFightObject{attacker}
	case global.Skill_Target_Type_Self_Rand1,
		global.Skill_Target_Type_Self_Rand2,
		global.Skill_Target_Type_Self_All:
		needNum := sTargetType - global.Skill_Target_Type_Self_Rand1 + 1
		needTargeters := []global.IFightObject{}
		srcTargets := append( []global.IFightObject{}, targeters...)
		for {
			if needNum == 0 || len(srcTargets) == 0 {
				break
			}

			index := global.ServerG.GetRandSrc().Intn(len(srcTargets))
			if !srcTargets[index].IsDead() {
				needTargeters = append(needTargeters, srcTargets[index])
				needNum--
			}

			srcTargets = append(srcTargets[:index], srcTargets[index+1:]...)
		}

		return needTargeters
	case global.Skill_Target_Type_Self_BloodMin_1:
		needTargeters := []global.IFightObject{}
		for _, targeter := range targeters {
			if !targeter.IsDead() {
				if len(needTargeters) == 0 {
					needTargeters = append(needTargeters, targeter)
				} else {
					if needTargeters[0].GetBlood() > targeter.GetBlood() {
						needTargeters[0] = targeter
					}
				}
			}
		}
		return needTargeters
	case global.Skill_Target_Type_Self_BloodMax_1:
		needTargeters := []global.IFightObject{}
		for _, targeter := range targeters {
			if !targeter.IsDead() {
				if len(needTargeters) == 0 {
					needTargeters = append(needTargeters, targeter)
				} else {
					if needTargeters[0].GetBlood() < targeter.GetBlood() {
						needTargeters[0] = targeter
					}
				}
			}
		}
		return needTargeters
	case global.Skill_Target_Type_Self_Dead_1:
		needTargeters := []global.IFightObject{}
		srcTargets := []global.IFightObject{}
		for _, targeter := range targeters {
			if targeter.IsDead() {
				srcTargets = append(srcTargets, targeter)
			}
		}

		if len(srcTargets) > 0 {
			index := global.ServerG.GetRandSrc().Intn(len(srcTargets))
			needTargeters = append(needTargeters, srcTargets[index])
		}

		return needTargeters
	}

	return nil
}
