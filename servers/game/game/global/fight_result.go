package global

import (
	"bytes"
	"encoding/binary"
	"xianxia/servers/game/conf"
)

//战斗条目结果
const (
	FIGHT_EVENT_SKILL_NORMAL = int8(1) + iota //普通伤害
	FIGHT_EVENT_SKILL_NOHIT                   //未命中
	FIGHT_EVENT_SKILL_CRIT                    //暴击
	FIGHT_EVENT_SKILL_RELIVE                  //复活
	FIGHT_EVENT_BUFF_ADD                      //添加buff
	FIGHT_EVENT_BUFF_DEL                      //删除buff
	FIGHT_EVENT_BUFFEFFECT                    //buff回合效果
	FIGHT_EVENT_SKILL_FRIEND                  //友军技能
)

//战斗事件类型
const (
	FIGHT_ITEM_ATTACK = int8(1) + iota //攻击,施法
	FIGHT_ITEM_BUFF                    //buff
)

type IFightEventData interface {
	ToBytes() []byte
}

type IFightItemData interface {
	ToBytes() []byte
}

//整个战斗过程的数据
type FightResultData struct {
	Attackers []IFightObject
	Defenders []IFightObject

	AttackWin bool
	BBoss     bool
	Items     []IFightItemData
	Reward    *RewardData
}

//定义单次战斗条目结合
type FightItemData_Base struct {
	CurRound   int8
	IType      int8
	EventDatas []IFightEventData
}

func (fi *FightItemData_Base) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, conf.RdWrEndian, fi.IType)
	binary.Write(buf, conf.RdWrEndian, fi.CurRound)

	if fi.EventDatas != nil {
		binary.Write(buf, conf.RdWrEndian, int8(len(fi.EventDatas)))
		for _, ei := range fi.EventDatas {
			binary.Write(buf, conf.RdWrEndian, ei.ToBytes())
		}
	} else {
		binary.Write(buf, conf.RdWrEndian, int8(0))
	}

	return buf.Bytes()
}

type FightItemData_Attack struct {
	FightItemData_Base
	AttackerPos int8
	SkillId     int32
}

func (fi *FightItemData_Attack) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, conf.RdWrEndian, fi.FightItemData_Base.ToBytes())

	binary.Write(buf, conf.RdWrEndian, fi.AttackerPos)
	binary.Write(buf, conf.RdWrEndian, fi.SkillId)

	return buf.Bytes()
}

type FightItemData_Buff struct {
	FightItemData_Base
}

//定义单次战斗中所有事件集合
type FightEventData_Base struct {
	EType int8
	Pos   int8
}

func (fe *FightEventData_Base) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, conf.RdWrEndian, fe.EType)
	binary.Write(buf, conf.RdWrEndian, fe.Pos)
	return buf.Bytes()
}

type FightEventData_Skill struct {
	FightEventData_Base
	ChangeProps map[int32]int32
}

func (fe *FightEventData_Skill) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, conf.RdWrEndian, fe.FightEventData_Base.ToBytes())
	if fe.ChangeProps != nil {
		binary.Write(buf, conf.RdWrEndian, int8(len(fe.ChangeProps)))
		for pid, pv := range fe.ChangeProps {
			binary.Write(buf, conf.RdWrEndian, pid)
			binary.Write(buf, conf.RdWrEndian, pv)
		}
	} else {
		binary.Write(buf, conf.RdWrEndian, int8(0))
	}
	return buf.Bytes()
}

type FightEventData_Buff struct {
	FightEventData_Base
	BuffId int32
}

func (fe *FightEventData_Buff) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, conf.RdWrEndian, fe.FightEventData_Base.ToBytes())
	binary.Write(buf, conf.RdWrEndian, fe.BuffId)
	return buf.Bytes()
}

type FightEventData_BuffEffect struct {
	FightEventData_Base
	BuffId      int32
	ChangeProps map[int32]int32
}

func (fe *FightEventData_BuffEffect) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, conf.RdWrEndian, fe.FightEventData_Base.ToBytes())
	binary.Write(buf, conf.RdWrEndian, fe.BuffId)
	if fe.ChangeProps != nil {
		binary.Write(buf, conf.RdWrEndian, int8(len(fe.ChangeProps)))
		for pid, pv := range fe.ChangeProps {
			binary.Write(buf, conf.RdWrEndian, pid)
			binary.Write(buf, conf.RdWrEndian, pv)
		}
	} else {
		binary.Write(buf, conf.RdWrEndian, int8(0))
	}

	return buf.Bytes()
}
