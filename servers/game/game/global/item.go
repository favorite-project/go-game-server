package global

const (
	ITEM_TYPE_EQUIPMENT  = int16(1) + iota //装备
	ITEM_TYPE_GOLD                         //金币
	ITEM_TYPE_EXP                          //经验
	ITEM_TYPE_DIAMOND                      //钻石
	ITEM_TYPE_TREASURBOX                   //宝箱
	ITEM_TYPE_BLOOD                        //加血道具
	ITEM_TYPE_MAGIC                        //加蓝道具
	ITEM_TYPE_STONE                        //宝石
	ITEM_TYPE_TASK                         //任务道具
	ITEM_TYPE_SKILL                        //技能书
	ITEM_TYPE_CHANGE_NAME				   //改名卡
	ITEM_TYPE_ENQUIP_UPDATA_STONE          //装备升级石头
	ITEM_TYPE_NORMAL         		//无用道具
	ITEM_TYPE_GIFT         		 	//礼包
	ITEM_TYPE_VIP         			//vip
	ITEM_TYPE_ADVANCE				//魂值
)

const (
	EQUIPMENT_TYPE_WEAPON   = int16(1) + iota //武器
	EQUIPMENT_TYPE_ARMOR                      //衣服
	EQUIPMENT_TYPE_RING                       //戒指
	EQUIPMENT_TYPE_NECKLACE                   //项链
	EQUIPMENT_TYPE_HELMET                     //帽子
	EQUIPMENT_TYPE_SHOES                      //鞋子
	EQUIPMENT_TYPE_WATCH                      //护腕
	EQUIPMENT_TYPE_BELT                       //腰带
)

const BACKPACK_INIT_NUM int32 = 64
const EQUIPMENT_PROPS_MAX_NUM = 5
const BackPack_Expand_Max_Count = 10
const BackPack_Expand_Base_Count = 6
const BackPack_Expand_Base_Diamond = 100

const (
	MONEY_ITEM_ID = int32(200001)
	EXP_ITEM_ID = int32(300001)
	DIAMOND_ITEM_ID = int32(400001)
)

const (
	BACKPACK_ITEMS_BAG = int16(1) + iota
	BACKPACK_EQUIP_BAG
	BACKPACK_STONE_BAG
	BACKPACK_SKILL_BAG
)

const (
	ITEM_USE_TYPE_BUY = uint16(1) + iota
	ITEM_USE_TYPE_SELL
	ITEM_USE_TYPE_NORMAL //消耗品是加血，宝箱是开，装备是装或卸
	ITEM_USE_TYPE_OPEN //可使用道具
)

//装备品质
const (
	EQUIPMENT_QUALITY_WHITE  = int16(1) + iota //白装
	EQUIPMENT_QUALITY_BLUE                     //蓝装
	EQUIPMENT_QUALITY_GREEN                    //绿装
	EQUIPMENT_QUALITY_PURPLE                   //紫装
	EQUIPMENT_QUALITY_ORANGE                   //橙装
	EQUIPMENT_QUALITY_RED                      //红装
)

type RewardItem struct {
	Id  int32
	Num int32
}

type RewardData struct {
	Items map[int32]*RewardItem
}

type ItemDBData struct {
	Id      int32  `redis:"id"`
	CfgId   int32  `redis:"cfgId"`
	Num     int32  `redis:"num"`
	Data    string `redis:"data"`
	Binding byte   `redis:"binding"`
}

type BackPackDBData struct {
	Num               int32                   `redis:"num"`     //最大背包数量
	BagData           map[int16][]*ItemDBData `redis:"bagdata"` //背包数据
	EquipSellSettings []int16                 `redis:"equipsellsettings"`
}

type EquipDBData struct {
	EquipData map[int16]*ItemDBData
}

type EquipDBItemData struct {
	UpdateLv int32           //强化等级
	UseCnt   int32           //熟练度
	BData    map[int32]int32 //基础属性
	OData    map[int32]int32 //附件属性
}

type SellItemCfgInfo struct {
	CfgId int32
	Num   int32
}

const (
	Vip_Effect_DailyGift = int32(1) + iota
	Vip_Effect_AddInstanceCount
	Vip_Effect_QuickFightCount
	Vip_Effect_FightTimeSec
	Vip_Effect_FightWaitSec
	Vip_Effect_AddMoneyDrop
	Vip_Effect_AddExpDrop
	Vip_Effect_MineRobTime
)