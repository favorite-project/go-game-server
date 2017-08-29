package global

//messageRootKey
const (
	Message_RootKey_Player = uint32(1) + iota //玩家消息
	Message_RootKey_Instance //副本消息
)

//messageRootKey_Player_Sub 玩家子消息
const (
	Message_RootKey_HeartBeat = uint32(1) + iota
	Message_RootKey_Player_Sub_Hi
	Message_RootKey_Player_Sub_Err
	Message_RootKey_Player_Create
	Message_RootKey_Player_Props_Update
	Message_RootKey_Player_Fight
	Message_RootKey_Player_BackPack
	Message_RootKey_Player_Add_BackPack_Item
	Message_RootKey_Player_Equip_ACT
	Message_RootKey_Player_Fight_Need_Time
	Message_RootKey_Player_Update_BackPack
	Message_RootKey_Player_UseItem
	Message_RootKey_Player_Fight_Reward
	Message_RootKey_Player_Change_Map
	Message_RootKey_Equip_Resolve
	Message_RootKey_Player_StudySkill
	Message_RootKey_Player_SkillLvUp
	Message_RootKey_Player_Change_SkillPos
	Message_RootKey_Player_SkillEquip
	Message_RootKey_Player_SkillUnEquip
	Message_RootKey_EquipCreate_Info
	Message_RootKey_EquipCreate_Refresh
	Message_RootKey_EquipCreate
	Message_RootKey_Player_OfflineReward
	Message_RootKey_Player_AutoSellEquip
	Message_RootKey_RandomStone
	Message_RootKey_RandomStoneCfg
	Message_RootKey_Player_LoginToken_Expired
	Message_RootKey_EquipUpdate
	Message_RootKey_ChangeName
	Message_RootKey_ExpandBag
	Message_RootKey_QuickFight
	Message_RootKey_CDKey
	Message_RootKey_Mails
	Message_RootKey_Mail_Reward
	Message_RootKey_Quick_Challenge
	Message_RootKey_Mine_Buy //买矿
	Message_RootKey_Mine_Reward   //挖矿收取
	Message_RootKey_Mine_Work		//打工
	Message_RootKey_SignInfo
	Message_RootKey_Sign_Reward
	Message_RootKey_LoginInfo
	Message_RootKey_Login_Reward

	Message_RootKey_Advance_LevelUp //渡劫升阶
	Message_RootKey_Vip_Reward
)

//messageRootKey_Instance_Sub 副本子消息
const (
	Message_RootKey_Instance_Enter = uint32(1) + iota //开启副本
	Message_RootKey_Instance_Info //同步副本消息到客户端
)

type NetMessage interface {
	MakeBuffer() []byte
}

type RootMessage struct {
	RootKey    uint32
	RootKeySub uint32
}

type PlayerPublicProps struct {
	Power        int32
	Agile        int32
	Intelligence int32
	Strength     int32
	Lucky        int32
	Endurance    int32

	Attack      int32
	Crit        int32
	AttackSpeed int32
	Miss        int32
	Blood       int32
	Get         int32
	Defence     int32
	Magic       int32

	Name       [NAME_MAX_LEN]byte
	Sex        int32
	Occupation int32
	Pic        [PIC_MAX_LEN]byte
	Level      int32
	MapId      int32
	RegionId   int32
	DBId       int64
	FightVal    int32
}

type PlayerPrivateProps struct {
	PlayerPublicProps
	Money          int32
	Diamond        int32
	VipLevel       int32
	Exp            int32
	FreeFightCount int32
	MaxMapId       int32
	MaxRegionId    int32
	EquipReslove   int32
	SkillPoint     int32
	QuickFightCount int32
	QuickFightTime int32
	RechargeNum    int32
	VipExp         int32
	AdvanceLevel    int32
	AdvanceExp      int32
	VipRewardTime    int32
}

type PlayerFightResult struct {
}

type PlayerBackPackItem struct {
	Uid          int32
	ItemSourceId int32
	Cnt          int32
	Name         string
	Lv           int32
	Description  string
	Category     int32
	SubCategory  int32
	Price        int32
	BaseAttr     string
	Icon         string
}

// 打造装备
const REFREASH_EQUIPINFO_ACT = 1
const LOADEQUIPINFO_ACT = 2
const CREATEQUIP_ACT = 3


type PlayerSignInfo struct {
	SignTime int32 `json:"signTime"`
	SignCnt  int32 `json:"signCnt"`
	ConSignCnt int32 `json:"conSignCnt"`
	RewardStates []int32 `json:"rewardStates"`
}

type PlayerLoginInfo struct {
	LoginTime int32 `json:"loginTime"`
	LoginCnt  int32 `json:"loginCnt"`
	ConLoginCnt int32 `json:"conLoginCnt"`
	TodayReward bool `json:"todayReward"`
	RewardStates []int32 `json:"rewardStates"`
}
