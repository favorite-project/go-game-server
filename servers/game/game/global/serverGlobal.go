package global

import (
	"github.com/name5566/leaf/module"
	"math/rand"
	"time"
	"xianxia/common/dbengine"
	"xianxia/common/event"
	"xianxia/common/timer"
	"xianxia/common/log"
)

type Singleton interface {
	Update(time.Time, int64)
}

type ServerGlobal interface {
	timer.ITimer //定时器
	Start(*module.Skeleton)
	Stop()
	GetState() int
	GetCurTime() int64
	GetObjectMgr() OBjectManager
	GetTimerMgr() *timer.CTimerMgr
	GetDBEngine() *dbengine.CDBEngine
	GetEventRouter() *event.CEventRouter
	GetMsgDispatcher() *event.CMsgDispatcher

	GetConfigMgr() ConfigMgr
	GetSkeleton() *module.Skeleton
	GetMapMgr() MapMgr
	GetFightMgr() FightMgr
	GetRandSrc() *rand.Rand
	GetSkillMgr() SkillMgr
	GetLog() *log.Log
	GetMailMgr() MailMgr
}

var ServerG ServerGlobal
