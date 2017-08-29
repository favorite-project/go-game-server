package global

import (
	"time"
)

//战斗模式
const (
	FIGHT_MODE_NORMAL = uint32(1) + iota
	FIGHT_MODE_INSTANCE
	FIGHT_MODE_BOSS
	FIGHT_MODE_PVP
	FIGHT_MODE_CHALLENGE
	FIGHT_MODE_MINE_ROB
	FIGHT_MODE_ADVANCE
)

//最大战斗人数
const MAX_FIGHTER_NUM = 3

// 最大战斗回合
const MAX_FIGHT_ROUNT = int8(50)

// 每个回合的时间(单位：秒)
const FIGHT_ROUND_TIME_SEC = int64(1.0 * float64(time.Second))

//快速战斗的最大次数
const QUICK_FIGHT_DURATION_SEC = int32(2 * 3600) //2小时
const QUICK_FIGHT_MAX_DIAMOND = int32(200) //20分钟

//战斗角色信息 需要具体实现
type IFightObject interface {
	ToBytes() []byte
	IsDead() bool
	GetBlood() int32
	SetBlood(int32)
	AddBuff(*BuffCfg)
	ClearBuff(int32)
	GetFighterSrc() Creature
	GetFightProp(int) int32
	SetFightProp(int, int32)
	CanBeAttacked() bool
	CanAttack() bool
	GetPos() int8
	BeAttacked() []IFightEventData
	Update(int8) []IFightEventData
	GetNowSkill() int32
	ResetNowSkill(int32)
	CanUseSkill() bool
}

type FightMgr interface {
	DoFight(Player, []Creature, []Creature, uint32, bool) *FightResultData
	DoRoundAttack(attacker IFightObject, defender IFightObject) IFightEventData
}
