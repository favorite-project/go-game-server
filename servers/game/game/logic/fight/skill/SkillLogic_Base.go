package skill

import (
	"xianxia/servers/game/game/global"
)

/*
const (
	Skill_Type_Buff                = int32(101)  //101 buff技能
	Skill_Type_Attack_Continue     = int32(1001) //1001 连续攻击 单次伤害百分比+固定额外伤害
	Skill_Type_Relive              = int32(1002) //1002 复活 回复血量百分比
	Skill_Type_Attack_Add          = int32(1003) //1003 攻击加固定额外伤害
	Skill_Type_Attack_Add_Buff     = int32(1004) //1004 攻击加固定额外伤害+buff
	Skill_Type_Attack_Per          = int32(1005) //1005 百分多少的伤害
	Skill_Type_Attack_Per_Buff     = int32(1006) //1006 百分多少的伤害+buff
	Skill_Type_Attack_Per_Add      = int32(1007) //1007 百分多少的伤害+额外伤害
	Skill_Type_Attack_Per_Add_Buff = int32(1008) //1008 百分多少的伤害+额外伤害+buff
)
*/

type SkillBuffItem struct {
	BuffCfg   *global.BuffCfg
	RandValue int32
	BuffLogic global.BuffLogic
}

type SkillLogic_Base struct {
	Type int32
}

func (sl *SkillLogic_Base) GetType() int32 {
	return sl.Type
}

func (sl *SkillLogic_Base) Logic(sCfg *global.SkillCfg, skillId int32, attacker global.IFightObject, attackers []global.IFightObject, defenders []global.IFightObject) []global.IFightEventData {
	return nil
}

func (sl *SkillLogic_Base) getSkillCfg(skillId int32) *global.SkillCfg {
	icfg := global.ServerG.GetConfigMgr().GetCfg("Skill", skillId)
	if icfg == nil {
		return nil
	}

	return icfg.(*global.SkillCfg)
}
