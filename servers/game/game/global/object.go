package global

import (
	"time"
	"xianxia/servers/game/game/global/db"
)

/*
1> 一级属性
	力量 【攻击力 暴击】
	智力 【魔法值】
	敏捷 【速度 闪避】
	体力 【生命值】
	幸运 【爆率】
	耐力 【防御】

2> 二级属性
	攻击力 力量*系数(10)
	暴击 力量*系数(0.01)
	速度
	闪避
	生命值
	爆率
	防御
	魔法值
*/

//生物属性
const (
	//一级属性
	Creature_Prop_One_Power        = 0  //力量
	Creature_Prop_One_Agile        = 1  //敏捷
	Creature_Prop_One_Intelligence = 2  //智力
	Creature_Prop_One_Strength     = 3  //体力
	Creature_Prop_One_Lucky        = 4  //幸运
	Creature_Prop_One_Endurance    = 5  //耐力
	Creature_Prop_One_Max          = 20 //一级属性最多20个

	//二级属性
	Creature_Prop_Two_Attack      = 20 //攻击力
	Creature_Prop_Two_Crit        = 21 //暴击 10000为基数
	Creature_Prop_Two_AttackSpeed = 22 //攻击速度
	Creature_Prop_Two_Miss        = 23 //闪避 10000为基数
	Creature_Prop_Two_Blood       = 24 //生命值
	Creature_Prop_Two_Get         = 25 //爆率
	Creature_Prop_Two_Defence     = 26 //防御
	Creature_Prop_Two_Magic       = 27 //魔法值
	Creature_Prop_Two_Hit         = 28 //命中，对立闪避 10000为基数
	Creature_Prop_Two_Tenacity    = 29 //抗暴，对立暴击 10000为基数
	Creature_Prop_Two_FAGain      = 30 //最终伤害增益
	Creature_Prop_Two_FAAdd       = 31 //最终额外伤害
	Creature_Prop_Two_BaseGain    = 32 //单次伤害增益
	Creature_Prop_Two_Max         = 50 //二级属性最多Creature_Prop_Two_Max -  Creature_Prop_One_Max = 30个
)

//monster属性
const (
	//其他属性
	Monster_Prop_Level   = 50 //等级
	Monster_Prop_Quality = 51 //金钱
	Monster_Prop_Max     = 100
)

//player属性
const (
	//其他属性
	Player_Prop_Level          = 50 //等级
	Player_Prop_Money          = 51 //金钱
	Player_Prop_Diamond        = 52 //钻石
	Player_Prop_VipLevel       = 53 //vip等级
	Player_Prop_Sex            = 54
	Player_Prop_Occupation     = 55
	Player_Prop_Exp            = 56
	Player_Prop_FreeFightCount = 57
	Player_Prop_MaxMapId       = 58
	Player_Prop_MaxRegionId    = 59
	Player_Equip_Reslove 	   = 60
	Player_Prop_SkillPoint     = 61
	Player_Prop_QuickFightCount = 62
	Player_Prop_QuickFightTime = 63
	Player_Prop_RechargeNum 	= 64
	Player_Prop_FightVal 		= 65 //战斗力
	Player_Prop_VipExp 			= 66
	Player_Prop_Advance_Level  = 67
	Player_Prop_Advance_Exp    = 68
	Player_Prop_VipReward_Time    = 69
	Player_Prop_Max            = 200 //角色属性最多 Player_Prop_Max - Creature_Prop_Two_Max = 150个
)

//生物类型
const (
	Creature_Type_None    = byte(1) + iota //基本生物类型：无类型
	Creature_Type_Monster                  //怪物
	Creature_Type_Player                   //角色
	Creature_Type_Pet                      //宠物.
)

const Crit_Per = 1500

//对外接口类
type Creature interface {
	GetProp(int) (int32, bool)
	SetProp(int, int32, bool) (int32, bool)
	GetType() byte
	GetName() []byte
	GetPic() []byte
	GetMapId() (int32, int32)
	SetMapId(int32, int32)
	Update(time.Time, int64)
	GetCfgId() int32

	OnFightEvent(bool, *Fight_Event_Info)

	GetSkillData() *SkillDBData

	GetInstanceMapId() int32
	SetInstanceMapId(int32)
	GetInstanceFightIndex() int
	SetInstanceFightIndex(int)
}

type Monster interface {
	Creature
}

type Player interface {
	Creature
	SetConnection(Connection)
	GetConnection() Connection
	OnRecv([]byte)
	GetDBId() int64
	GetPublicProps(*PlayerPublicProps)
	GetPrivateProps(*PlayerPrivateProps)
	GetPlayerBackPack()
	IsBagFull(cfgId int32, num int32, autoSell bool) bool
	IsBagFullMulti(items *RewardData, autoSell bool) bool
	AddItem(int32, int32, bool, bool) ([]*ItemDBData, []int32, []*SellItemCfgInfo, error)
	AddItems(*RewardData, bool, bool) ([]*ItemDBData, []int32, []*SellItemCfgInfo, error)
	SubItemByInstId(instId int32, num int32, sendMsg bool) ([]*ItemDBData, []int32, error)
	Online(bool)
	Offline()
	GetLastOnlineTime() int32
	GetLastOfflineTime() int32
	GetLoadDataTime() int64
	IsOnline() bool
	SetLoginType(int)
	SetOpenId(string)
	SetDeviceType(string)
	GetLoginType() int
	GetOpenId() string
	GetDeviceType() string
	Kick()
	GetVipEffectValue(int32) int32
	Recharge(int32)
}

type OffLinePlayer interface {
	Creature
	GetDBId() int64
}

type OBjectManager interface {
	Singleton

	Create() bool
	Stop() bool
	GetPlayer(int64) Player
	CreatePlayerFromDB(int64, *db.DB_PLayer_Props, Connection) (Player, error)
	CreatePlayer(Connection) (Player, error)
	PlayerIsOnline(int64) bool
	SetPlayerOnline(int64)
	GetOnlinePlayer() []int64
	CreateMonster(*MonsterCfg) Monster
	GenerateUUid() int64
	GetOfflinePlayer(int64) OffLinePlayer
}
