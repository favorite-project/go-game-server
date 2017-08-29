package global

import (
	"github.com/name5566/leaf/recordfile"
)

type ConfigMgr interface {
	Start() bool
	GetCsv(string) *recordfile.RecordFile
	GetCfg(string, interface{}) interface{}
	Reload()
}

type MonsterCfg struct {
	Id          int32 "index"
	Name        string
	Desc        string
	Level       int32
	Quality     int32
	Icon        string
	DropData    string
	Defence     int32
	Attack      int32
	Blood       int32
	Attackspeed int32
	Miss        int32
	Hit         int32
	Crit        int32
	Tenacity    int32
	Skills      string
}

type MapCfg struct {
	Id        int32 "index"
	Name      string
	Desc      string
	OpenLevel int32
	Icon      string
	Regions   string
	NextMapId int32
}

type RegionCfg struct {
	Id      int32 "index"
	Name    string
	Desc    string
	Icon    string
	MonData string
	MonNum  int32
	BossId  string
}

type PropsCaculationCfg struct {
	Id          int32
	PropId      int
	EffectProps string
	ValueMin    int32
	ValueMax    int32
}

type PlayerLevelCfg struct {
	Level   		int32 "index"
	NeedExp  		int32
	AddProps 		string
	AddSkillPoint 	int32
}

type ItemCfgInterface interface {
	GetId() int32
	GetType() int16
	GetMaxNum() int32
	GetSellPrice() int32
}

type ItemCfg struct {
	Id        int32 "index"
	Name      string
	Type      int16
	Level     int32
	Desc      string
	SellPrice int32  //-1表示不可卖
	Data      string //数据
	Icon      string
	MaxNum    int32 //最大叠加数量
	Quality   int16
}

func (i *ItemCfg) GetId() int32 {
	return i.Id
}

func (i *ItemCfg) GetType() int16 {
	return i.Type
}

func (i *ItemCfg) GetMaxNum() int32 {
	return i.MaxNum
}

func (i *ItemCfg) GetSellPrice() int32 {
	return i.SellPrice
}

type EquipmentCfg struct {
	Id        int32 "index"
	Name      string
	Type      int16
	Level     int32
	Desc      string
	SellPrice int32  //-1表示不可卖
	Data      string //数据
	Icon      string
	MaxNum    int32 //最大叠加数量
	Quality   int16
	OtherData float64
	SuitId    int32
	SubType   int16
}

func (i *EquipmentCfg) GetId() int32 {
	return i.Id
}

func (i *EquipmentCfg) GetType() int16 {
	return i.Type
}

func (i *EquipmentCfg) GetMaxNum() int32 {
	return i.MaxNum
}

func (i *EquipmentCfg) GetSellPrice() int32 {
	return i.SellPrice
}

type DropBoxCfg struct {
	Id     int32 "index"
	Reward string
	Desc   string
}

type EquipmentPropsCfg struct {
	Id       int32 "index"
	Level    int32
	EType    int16
	PropId   int32
	MinValue int32
	MaxValue int32
	Desc     string
}

//buff基本配置 一个buff最多支持三组效果，后续可扩展
type BuffCfg struct {
	Id    int32 "index"
	Name  string
	Desc  string
	Icon  string
	Type  int16  //作用类型
	Value string //作用值
	Round int8   //持续几个回合
}

//技能基本配置 目前支持3种类型和1种buff
type SkillCfg struct {
	Id             int32 "index"
	Name           string
	Desc           string
	Icon           string
	Friend         bool //是否是友军技能
	Type           int8
	RType          int32
	SrcSkillId     int32
	CanPlayerStrud bool
	ActiveCond     string
	Level          int32
	NextLevel      int32
	MaxLevel       int32
	UpLvCostPoint  int32  //升级消耗技能点数
	UpLvCostItems  string //升级消耗的金钱，钻石，物品之类
	CDRound        int8   //持续几个回合
	TargetType     int16  //施法对象规则
	Param1         int32
	Value1         int32
	Param2         int32
	Value2         int32
	Param3         int32
	Value3         int32
	Param4         int32
	Value4         int32
	StrParams      string
}

// 赌石配置
type RandomStone struct {
	Id    int32 "index"
	Name  string
	Price int32
	Total int
	Rate  string
}

// 装备强化配置
type EquipUpdate struct {
	Id    int32 "index"
	UpdateGrowth float64
	ItemId int32
	StoneCnt int32
	Money int32
	SuccRate int32
}

// 装备打造配置
type EquipCreateCfg struct {
	Id    int32 "index"
	Rate string
}

//副本配置
type InstanceCfg struct {
	Id int32 "index"
	Name string
	Desc string
	Level int32
	OpenMapId int32
	FreeCount int32
	NeedItemId int32
	Icon string
	MonData string
	DropData string
	Fight int
	DropShowData string
}

//vip配置
type VipCfg struct {
	Level int32 "index"
	Name string
	Desc string
	Icon string
	Recharge int32
	DailyGiftId int32
	AddInstanceCount int32
	QuickFightCount int32
	FightTimeSec int32
	FightWaitSec int32
	AddMoneyDrop int32
	AddExpDrop   int32
	MineRobTime  int32
}

// 挑战配置
type Challenge struct {
	Id int32 "index"
	Name string
	Produce string
	Cnt int32
	Desc string
	Icon string
	CompleteReward string
	Reward string
	QuickReward string
}
type ChallengeMonster struct {
	Id          int32 "index"
	Lid 		int32
	Cid         int32
	Mid 		int32
	DropData    string
}

type MineCfg struct {
	Id int32 "index"
	Name string
	Desc string
	Icon string
	Quality int
	Level int32
	NextId int32
	BuyItems string
	LvUpItems string
	ProductItemId int32
	RobPer int //1000
	FullSec int32
	PerSecNum int32
	Works string
}

type MineWorkCfg struct {
	Id int32 "index"
	Name string
	Desc string
	MaxCount int32
	ItemId int32
	ItemNum int32
}

//套装
type SuitCfg struct {
	Id int32 "index"
	Name string
	Data1 string
	Data2 string
	Data3 string
	Data4 string
	Data5 string
	Data6 string
	Data7 string
	Data8 string
}


//签到奖励
type SignRewardCfg struct {
	Day int32 "index"
	Name string
	Desc string
	Items string
}

//登陆奖励
type LoginRewardCfg struct {
	Day int32 "index"
	Name string
	Desc string
	Items string
}

type AdvanceCfg struct {
	Level int32 "index"
	Name  string
	Desc string
	NeedExp int32
	AddProps string
	BossId int32
	PlayerLevel int32
	ItemId1 int32
	ItemNum1 int32
	ItemId2 int32
	ItemNum2 int32
	ItemId3 int32
	ItemNum3 int32
}
