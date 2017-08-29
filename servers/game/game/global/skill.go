package global

//技能类型
const (
	Skill_Type_Passive = int8(1) + iota //被动技能
	Skill_Type_Active                   //主动技能
)

//buff类型
const (
	Buff_Type_Dizzy             = int16(1) + iota //1 眩晕
	Buff_Type_Invincible                          //2 无敌
	Buff_Type_SkillDisable                        //3 禁法
	Buff_Type_Sleep                               //4 睡觉
	Buff_Type_Prop                                //5 增加战斗属性值
	Buff_Type_Prop_Per                            //6 增加战斗属性值百分比
	Buff_Type_Prop_Continue                       //7 按照固定值增加战斗属性
	Buff_Type_Prop_Per_Continue                   //8 按照百分比增加战斗属性
)

//buff作用对象（位置）[敌方]
const (
	Skill_Target_Type_Pos1       = int16(1) + iota //1 对方单体，按照顺序攻击
	Skill_Target_Type_Pos2                         //2 对方2号位，如果没有找到就按照Pos3，Pos1的顺序找
	Skill_Target_Type_Pos3                         //3 对方3号位，如果没有找到就按照Pos2,Pos1的顺序找
	Skill_Target_Type_Rand1                        //4 对方列表中随机一个
	Skill_Target_Type_Rand2                        //5 对方列表中随机两个
	Skill_Target_Type_All                          //6 全部攻击
	Skill_Target_Type_BloodMin_1                   //7 对方生命最少的目标
	Skill_Target_Type_BloodMax_1                   //8 对方生命最多的目标
	Skill_Target_Type_Enemy_Max
)

//buff作用对象（位置）[友方]
const (
	Skill_Target_Type_Self_1          = int16(100) + iota //100 自己
	Skill_Target_Type_Self_Rand1                          //101 己方列表中随机一个
	Skill_Target_Type_Self_Rand2                          //102 己方列表中随机两个
	Skill_Target_Type_Self_All                            //103 己方全部
	Skill_Target_Type_Self_BloodMin_1                     //104 己方血量最少的
	Skill_Target_Type_Self_BloodMax_1                     //105 己方血量最多的
	Skill_Target_Type_Self_Dead_1                         //106 己方已死亡中随机一个
)

const Player_Max_Skill_Num = 5

type SkillDBItem struct {
	SkillId int32
	Pos int32
}
type SkillDBData struct {
	Equips map[int32]*SkillDBItem
	Bags map[int32]*SkillDBItem
}

type BuffLogic interface {
	GetType() int16
	CanAttack() bool
	CanUseSkill() bool
	CanBeAttacked() bool
	CanBeInterrupt() bool
	EffectPerRound(IFightObject, int32) IFightEventData //每回合执行
	EffectNow(IFightObject, int32) IFightEventData      //是否立即执行
	Reset(IFightObject, int32) IFightEventData          //恢复
}

//技能类型(注：buff都是概率触发 1000是必然触发)
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
	Skill_Type_Add_Prop            = int32(1009) //1009回血回蓝加永久加属性等
)

//技能激活条件
const (
	Skill_ActiveCond_Type_Level = 1 + iota
	Skill_ActiveCond_Type_Advance
)

type SkillLogic interface {
	GetType() int32
	Logic(sCfg *SkillCfg, skillId int32, attacker IFightObject, attackers []IFightObject, defenders []IFightObject) []IFightEventData
}

type SkillMgr interface {
	GetBuffLogic(int32) BuffLogic
	Logic(skillId int32, attacker IFightObject, attackers []IFightObject, defenders []IFightObject) []IFightEventData
}
