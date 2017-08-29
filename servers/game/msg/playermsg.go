package msg

import (
	"bytes"
	"encoding/binary"
	"xianxia/servers/game/game/global"
	//"fmt"
	"fmt"
)

//握手消息
type GSCL_Error struct {
	global.RootMessage
	Desc []byte
}

//错误信息
type GSCL_Hi struct {
	global.RootMessage
}

func (msg *GSCL_Hi) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_Sub_Hi

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)

	return buf.Bytes()
}

func (msg *GSCL_Error) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_Sub_Err

	if msg.Desc == nil { //兼容
		msg.Desc = make([]byte, 1)
	}

	data := make([]byte, 4+4+len(msg.Desc))
	RdWrEndian.PutUint32(data, msg.RootMessage.RootKey)
	RdWrEndian.PutUint32(data[4:], msg.RootMessage.RootKeySub)
	copy(data[8:], msg.Desc)

	return data
}

//登录返回角色所有信息
type GSCL_CreatePlayer struct {
	global.RootMessage
	Create byte
	Now 	int32
	global.PlayerPrivateProps
	*global.EquipDBData
	*global.SkillDBData
	Defines []interface{}
}

func (msg *GSCL_CreatePlayer) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_Create

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)

	binary.Write(buf, RdWrEndian, msg.Create)
	binary.Write(buf, RdWrEndian, msg.Now)

	binary.Write(buf, RdWrEndian, msg.PlayerPrivateProps)

	binary.Write(buf, RdWrEndian, int16(len(msg.EquipDBData.EquipData)))
	for bid, itemData := range msg.EquipDBData.EquipData {
		binary.Write(buf, RdWrEndian, bid)

		binary.Write(buf, RdWrEndian, itemData.Id)
		binary.Write(buf, RdWrEndian, itemData.CfgId)
		binary.Write(buf, RdWrEndian, itemData.Num)

		binary.Write(buf, RdWrEndian, int32(len([]byte(itemData.Data))))
		binary.Write(buf, RdWrEndian, []byte(itemData.Data))
		binary.Write(buf, RdWrEndian, itemData.Binding)
	}

	binary.Write(buf, RdWrEndian, int16(len(msg.SkillDBData.Equips)))
	for _, sItem := range msg.SkillDBData.Equips {
		binary.Write(buf, RdWrEndian, sItem)
	}
	binary.Write(buf, RdWrEndian, int16(len(msg.SkillDBData.Bags)))
	for _, sItem := range msg.SkillDBData.Bags {
		binary.Write(buf, RdWrEndian, sItem)
	}

	if msg.Defines != nil {
		for _, value := range msg.Defines {
			binary.Write(buf, RdWrEndian, value)
		}
	}

	return buf.Bytes()
}

//属性变化
type GSCL_PlayerUpdateProps struct {
	global.RootMessage
	Props map[int32]int32
}

func (msg *GSCL_PlayerUpdateProps) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_Props_Update

	buf := new(bytes.Buffer)

	binary.Write(buf, RdWrEndian, msg.RootMessage)

	for propId, propValue := range msg.Props {
		binary.Write(buf, RdWrEndian, propId)
		binary.Write(buf, RdWrEndian, propValue)
	}

	return buf.Bytes()
}

type GSCL_PlayerFightInfo struct {
	global.RootMessage
	Mode int32
	InstanceEnd bool
	Data *global.FightResultData
	//RewardData *global.FightRewardData
}

func (msg *GSCL_PlayerFightInfo) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_Fight

	buf := new(bytes.Buffer)

	binary.Write(buf, RdWrEndian, msg.RootMessage)

	binary.Write(buf, RdWrEndian, msg.Mode)
	binary.Write(buf, RdWrEndian, msg.InstanceEnd)

	if msg.Data.Attackers != nil {
		binary.Write(buf, RdWrEndian, uint16(len(msg.Data.Attackers)))
		for _, attack := range msg.Data.Attackers {
			binary.Write(buf, RdWrEndian, attack.ToBytes())
		}
	}

	if msg.Data.Defenders != nil {
		binary.Write(buf, RdWrEndian, uint16(len(msg.Data.Defenders)))
		for _, defender := range msg.Data.Defenders {
			binary.Write(buf, RdWrEndian, defender.ToBytes())
		}
	}

	binary.Write(buf, RdWrEndian, msg.Data.AttackWin)
	binary.Write(buf, RdWrEndian, msg.Data.BBoss)
	binary.Write(buf, RdWrEndian, uint16(len(msg.Data.Items)))
	for _, item := range msg.Data.Items {
		binary.Write(buf, RdWrEndian, item.ToBytes())
	}

	if msg.Data.Reward == nil || msg.Data.Reward.Items == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Data.Reward.Items)))
		for _, rItem := range msg.Data.Reward.Items {
			binary.Write(buf, RdWrEndian, rItem.Id)
			binary.Write(buf, RdWrEndian, rItem.Num)
		}
	}

	return buf.Bytes()
}

type GSCL_PlayerFightReward struct {
	global.RootMessage
	SellInfo []*global.SellItemCfgInfo
}

func (msg *GSCL_PlayerFightReward) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_Fight_Reward

	buf := new(bytes.Buffer)

	binary.Write(buf, RdWrEndian, msg.RootMessage)
	if msg.SellInfo != nil {
		binary.Write(buf, RdWrEndian, int16(len(msg.SellInfo)))
		for _, item := range msg.SellInfo {
			binary.Write(buf, RdWrEndian, item)
		}
	} else {
		binary.Write(buf, RdWrEndian, int16(0))
	}

	return buf.Bytes()
}

//战斗请求失败的消息
type GSCL_PlayerFightNeedTime struct {
	global.RootMessage
	Mode int32
	Time int32
}

func (msg *GSCL_PlayerFightNeedTime) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_Fight_Need_Time

	buf := new(bytes.Buffer)

	binary.Write(buf, RdWrEndian, msg.RootMessage)

	binary.Write(buf, RdWrEndian, msg.Mode)

	binary.Write(buf, RdWrEndian, msg.Time)

	return buf.Bytes()
}

// 创建背包
type GSCL_CreateBackPack struct {
	global.RootMessage
	*global.BackPackDBData
}

func (msg *GSCL_CreateBackPack) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_BackPack

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.BackPackDBData.Num)
	binary.Write(buf, RdWrEndian, int16(len(msg.BackPackDBData.BagData)))
	for bid, bagData := range msg.BackPackDBData.BagData {
		binary.Write(buf, RdWrEndian, bid)
		binary.Write(buf, RdWrEndian, int32(len(bagData)))
		for _, itemData := range bagData {
			binary.Write(buf, RdWrEndian, itemData.Id)
			binary.Write(buf, RdWrEndian, itemData.CfgId)
			binary.Write(buf, RdWrEndian, itemData.Num)
			binary.Write(buf, RdWrEndian, int32(len([]byte(itemData.Data))))
			binary.Write(buf, RdWrEndian, []byte(itemData.Data))
			binary.Write(buf, RdWrEndian, itemData.Binding)
		}
	}

	return buf.Bytes()
}

//背包变化，主要是装备等一些变动了信息
type GSCL_PlayerUpdateBackPack struct {
	global.RootMessage
	AddItems []*global.ItemDBData
	DelItems []int32
}

func (msg *GSCL_PlayerUpdateBackPack) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_Update_BackPack

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)

	if msg.AddItems != nil {
		binary.Write(buf, RdWrEndian, int16(len(msg.AddItems)))
		for _, itemData := range msg.AddItems {
			binary.Write(buf, RdWrEndian, itemData.Id)
			binary.Write(buf, RdWrEndian, itemData.CfgId)
			binary.Write(buf, RdWrEndian, itemData.Num)

			binary.Write(buf, RdWrEndian, int32(len([]byte(itemData.Data))))
			binary.Write(buf, RdWrEndian, []byte(itemData.Data))
			binary.Write(buf, RdWrEndian, itemData.Binding)
		}
	} else {
		binary.Write(buf, RdWrEndian, int16(0))
	}

	if msg.DelItems != nil {
		binary.Write(buf, RdWrEndian, int16(len(msg.DelItems)))
		for _, instId := range msg.DelItems {
			binary.Write(buf, RdWrEndian, instId)
		}
	} else {
		binary.Write(buf, RdWrEndian, int16(0))
	}

	return buf.Bytes()
}

//熔炼值
type GSCL_PlayerEquipResloveVal struct {
	global.RootMessage
	ResloveVal int32
}

func (msg *GSCL_PlayerEquipResloveVal) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Equip_Resolve
	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.ResloveVal)
	return buf.Bytes()
}

//获取打造装备信息
type GSCL_EquipCreateInfo struct {
	global.RootMessage
	EquipCfgId       int32
	FreeTimes        int32
	CostEquipReslove int32
	PkId             int32
	PkVal            int32
	ActId            int32
}

func (msg *GSCL_EquipCreateInfo) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_EquipCreate_Info
	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.EquipCfgId)
	binary.Write(buf, RdWrEndian, msg.FreeTimes)
	binary.Write(buf, RdWrEndian, msg.CostEquipReslove)
	binary.Write(buf, RdWrEndian, msg.PkId)
	binary.Write(buf, RdWrEndian, msg.PkVal)
	binary.Write(buf, RdWrEndian, msg.ActId)
	return buf.Bytes()
}

//技能學習
type GSCL_PlayerSkilLStudy struct {
	global.RootMessage
	SkillItem *global.SkillDBItem
}

func (msg *GSCL_PlayerSkilLStudy) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_StudySkill

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.SkillItem)

	return buf.Bytes()
}

//技能改位置
type GSCL_PlayerSkilLChangePos struct {
	global.RootMessage
	NewSItem *global.SkillDBItem
	OldSItem *global.SkillDBItem
}

func (msg *GSCL_PlayerSkilLChangePos) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_Change_SkillPos

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.NewSItem)
	if msg.OldSItem != nil {
		binary.Write(buf, RdWrEndian, true)
		binary.Write(buf, RdWrEndian, msg.OldSItem)
	} else {
		binary.Write(buf, RdWrEndian, false)
	}
	return buf.Bytes()
}

//技能升級
type GSCL_PlayerSkillLevelUp struct {
	global.RootMessage
	DelSkillId int32
	AddSItem   *global.SkillDBItem
}

func (msg *GSCL_PlayerSkillLevelUp) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_SkillLvUp

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.DelSkillId)
	binary.Write(buf, RdWrEndian, msg.AddSItem)

	return buf.Bytes()
}

//技能裝備
type GSCL_PlayerSkillEquip struct {
	global.RootMessage
	EquipSItem *global.SkillDBItem
	BagSItem   *global.SkillDBItem
}

func (msg *GSCL_PlayerSkillEquip) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_SkillEquip

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)

	binary.Write(buf, RdWrEndian, msg.EquipSItem)
	if msg.BagSItem != nil {
		binary.Write(buf, RdWrEndian, true)
		binary.Write(buf, RdWrEndian, msg.BagSItem)
	} else {
		binary.Write(buf, RdWrEndian, false)
	}

	return buf.Bytes()
}

//离线奖励
type GSCL_PlayerOfflineReward struct {
	global.RootMessage
	OfflineSec int32
	FightCount int32
	Items      map[int32]int32
	SellItems  map[int32]int32
}

func (msg *GSCL_PlayerOfflineReward) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_OfflineReward

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.OfflineSec)
	binary.Write(buf, RdWrEndian, msg.FightCount)
	if msg.Items == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Items)))
		for cfgId, num := range msg.Items {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	if msg.SellItems == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.SellItems)))
		for cfgId, num := range msg.SellItems {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	return buf.Bytes()
}

//自动卖装
type GSCL_PlayerEquipAutoSell struct {
	global.RootMessage
	EquipAutoSell []int16
}

func (msg *GSCL_PlayerEquipAutoSell) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_AutoSellEquip

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, int16(len(msg.EquipAutoSell)))
	for _, quality := range msg.EquipAutoSell {
		binary.Write(buf, RdWrEndian, quality)
	}

	return buf.Bytes()
}

// 赌石结果
type GSCL_RandomStone struct {
	global.RootMessage
	Ret int32
	Cnt int32
	ItemIds map[int32]int32

}
func (msg *GSCL_RandomStone) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_RandomStone
	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.Ret)
	msg.Cnt = int32(len(msg.ItemIds))
	binary.Write(buf, RdWrEndian, msg.Cnt)
	fmt.Println("发送长度:",msg.Cnt, "结果:", msg.Cnt)
	if msg.Cnt > 0 {
		for item_id,cnt := range msg.ItemIds {
			fmt.Println("item_id:", item_id, "item_cnt:",cnt)
			binary.Write(buf, RdWrEndian, item_id)
			binary.Write(buf, RdWrEndian, cnt)
		}
	}
	return buf.Bytes()
}
type GSCL_RandomStoneCfg struct {
	global.RootMessage
	PrimaryPrice int32
	MiddlePrice int32
	HighPrice int32
}

func (msg *GSCL_RandomStoneCfg) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_RandomStoneCfg
	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.PrimaryPrice)
	binary.Write(buf, RdWrEndian, msg.MiddlePrice)
	binary.Write(buf, RdWrEndian, msg.HighPrice)
	return buf.Bytes()
}

//登录token过期
type GSCL_PlayerLoginTokenExpired struct {
	global.RootMessage
}

func (msg *GSCL_PlayerLoginTokenExpired) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Player_LoginToken_Expired

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)

	return buf.Bytes()
}

// 装备升级
type GSCL_EquipUpdate struct {
	global.RootMessage
	Pk int32
	Ret int32
	NewValue int32
	NewEquipLv int32
}

func (msg *GSCL_EquipUpdate) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_EquipUpdate
	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.Ret)
	binary.Write(buf, RdWrEndian, msg.Pk)
	binary.Write(buf, RdWrEndian, msg.NewValue)
	binary.Write(buf, RdWrEndian, msg.NewEquipLv)
	return buf.Bytes()
}

// 心跳包
type GSCL_HeartBeat struct {
	global.RootMessage
	NowSec int32
}

func (msg *GSCL_HeartBeat) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_HeartBeat
	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.NowSec)

	return buf.Bytes()
}

// 快速战斗
type GSCL_QuickFight struct {
	global.RootMessage
	FightCount int32
	Items      map[int32]int32
	SellItems  map[int32]int32
}

func (msg *GSCL_QuickFight) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_QuickFight

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.FightCount)
	if msg.Items == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Items)))
		for cfgId, num := range msg.Items {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	if msg.SellItems == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.SellItems)))
		for cfgId, num := range msg.SellItems {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	return buf.Bytes()
}

type GSCL_CDKeyReward struct {
	global.RootMessage
	Items      map[int32]int32
	SellItems  map[int32]int32
}


func (msg *GSCL_CDKeyReward) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_CDKey

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	if msg.Items == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Items)))
		for cfgId, num := range msg.Items {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	if msg.SellItems == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.SellItems)))
		for cfgId, num := range msg.SellItems {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	return buf.Bytes()
}

//系统邮件
type GSCL_PlayerMailsInfo struct {
	global.RootMessage
	Mails map[int32]*global.PlayerMailInfo
}

func (msg *GSCL_PlayerMailsInfo) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Mails

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)

	if msg.Mails == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Mails)))

		for _, mail := range msg.Mails {
			binary.Write(buf, RdWrEndian, mail.Id)
			binary.Write(buf, RdWrEndian, mail.SysId)
			binary.Write(buf, RdWrEndian, mail.Type)
			binary.Write(buf, RdWrEndian, uint16(len(mail.Title)))
			binary.Write(buf, RdWrEndian, []byte(mail.Title))
			binary.Write(buf, RdWrEndian, uint16(len(mail.Content)))
			binary.Write(buf, RdWrEndian, []byte(mail.Content))
			binary.Write(buf, RdWrEndian, mail.ExpireTime)
			binary.Write(buf, RdWrEndian, mail.CreateTime)
			binary.Write(buf, RdWrEndian, mail.State)
			if mail.Items == nil {
				binary.Write(buf, RdWrEndian, int16(0))
			} else {
				binary.Write(buf, RdWrEndian, int16(len(mail.Items)))
				for itemId, num := range mail.Items {
					binary.Write(buf, RdWrEndian, itemId)
					binary.Write(buf, RdWrEndian, num)
				}
			}
		}
	}

	return buf.Bytes()
}

//邮件领取消息
type GSCL_MailReward struct {
	global.RootMessage
	Ret int32
	Id int32
	Items      map[int32]int32
	SellItems  map[int32]int32
}

func (msg *GSCL_MailReward) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Mail_Reward

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.Ret)
	binary.Write(buf, RdWrEndian, msg.Id)
	if msg.Items == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Items)))
		for cfgId, num := range msg.Items {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	if msg.SellItems == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.SellItems)))
		for cfgId, num := range msg.SellItems {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	return buf.Bytes()
}

// 扫荡
type GSCL_QuickChallenge struct {
	global.RootMessage
	Ret int32
	Challenge_id int32
	ItemId int32
	Val int32
}

func (msg *GSCL_QuickChallenge) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Quick_Challenge
	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.Ret)
	binary.Write(buf, RdWrEndian, msg.Challenge_id)
	binary.Write(buf, RdWrEndian, msg.ItemId)
	binary.Write(buf, RdWrEndian, msg.Val)
	return buf.Bytes()
}

//挖矿领取消息
type GSCL_MineReward struct {
	global.RootMessage
	CfgId int32
	Items      map[int32]int32
	SellItems  map[int32]int32
}

func (msg *GSCL_MineReward) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Mine_Reward

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.CfgId)

	if msg.Items == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Items)))
		for cfgId, num := range msg.Items {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	if msg.SellItems == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.SellItems)))
		for cfgId, num := range msg.SellItems {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	return buf.Bytes()
}

//挖矿打工消息
type GSCL_MineWork struct {
	global.RootMessage
	Ret int32
	CfgId int32
	MWorkCounts map[int32]int32
	Items      map[int32]int32
	SellItems  map[int32]int32
}

func (msg *GSCL_MineWork) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Mine_Work

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.Ret)
	binary.Write(buf, RdWrEndian, msg.CfgId)
	if msg.MWorkCounts == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.MWorkCounts)))
		for wid, num := range msg.MWorkCounts {
			binary.Write(buf, RdWrEndian, wid)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	if msg.Items == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Items)))
		for cfgId, num := range msg.Items {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	if msg.SellItems == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.SellItems)))
		for cfgId, num := range msg.SellItems {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, num)
		}
	}

	return buf.Bytes()
}


type CommonMsg_Reward struct {
	Items      *global.RewardData
	Sells []*global.SellItemCfgInfo
}

func(msg * CommonMsg_Reward) MakeBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}

	if msg.Items == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Items.Items)))
		for cfgId, data := range msg.Items.Items {
			binary.Write(buf, RdWrEndian, cfgId)
			binary.Write(buf, RdWrEndian, data.Num)
		}
	}

	if msg.Sells == nil {
		binary.Write(buf, RdWrEndian, int16(0))
	} else {
		binary.Write(buf, RdWrEndian, int16(len(msg.Sells)))
		for _, data := range msg.Sells {
			binary.Write(buf, RdWrEndian, data.CfgId)
			binary.Write(buf, RdWrEndian, data.Num)
		}
	}
}

//签到消息
type GSCL_SignInfo struct {
	global.RootMessage
	*global.PlayerSignInfo
}

func (msg *GSCL_SignInfo) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_SignInfo

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)

	binary.Write(buf, RdWrEndian, msg.SignTime)
	binary.Write(buf, RdWrEndian, msg.SignCnt)
	binary.Write(buf, RdWrEndian, msg.ConSignCnt)
	binary.Write(buf, RdWrEndian, int16(len(msg.RewardStates)))
	for _, s := range msg.RewardStates {
		binary.Write(buf, RdWrEndian, s)
	}

	return buf.Bytes()
}

type GSCL_SignReward struct {
	global.RootMessage
	Day  int32
	Reward *CommonMsg_Reward
}

func (msg *GSCL_SignReward) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Sign_Reward

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.Day)
	msg.Reward.MakeBuffer(buf)

	return buf.Bytes()
}

//登陆消息
type GSCL_LoginInfo struct {
	global.RootMessage
	*global.PlayerLoginInfo
}

func (msg *GSCL_LoginInfo) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_LoginInfo

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.PlayerLoginInfo)

	binary.Write(buf, RdWrEndian, msg.LoginTime)
	binary.Write(buf, RdWrEndian, msg.LoginCnt)
	binary.Write(buf, RdWrEndian, msg.ConLoginCnt)
	binary.Write(buf, RdWrEndian, msg.TodayReward)
	binary.Write(buf, RdWrEndian, int16(len(msg.RewardStates)))
	for _, s := range msg.RewardStates {
		binary.Write(buf, RdWrEndian, s)
	}

	return buf.Bytes()
}

type GSCL_LoginReward struct {
	global.RootMessage
	Day  int32
	Reward *CommonMsg_Reward
}

func (msg *GSCL_LoginReward) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Player
	msg.RootMessage.RootKeySub = global.Message_RootKey_Login_Reward

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.Day)
	msg.Reward.MakeBuffer(buf)

	return buf.Bytes()
}