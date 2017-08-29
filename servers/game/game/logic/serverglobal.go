package logic

import (
	"github.com/name5566/leaf/module"
	"math/rand"
	"time"
	"xianxia/common/dbengine"
	"xianxia/common/event"
	"xianxia/common/timer"
	"xianxia/servers/game/conf"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/game/logic/cfg"
	"xianxia/servers/game/game/logic/fight"
	"xianxia/servers/game/game/logic/fight/skill"
	"xianxia/servers/game/game/logic/maps"
	"xianxia/servers/game/game/logic/object"
	"xianxia/common/log"
	"xianxia/servers/game/game/logic/instance"
	"xianxia/servers/game/game/logic/mail"
)

type CServerGlobal struct {
	timerMgr      *timer.CTimerMgr
	objectMgr     *object.CObjectManager
	dbEngine      *dbengine.CDBEngine
	state         int
	curTime       time.Time
	processTime   int64
	skeleton      *module.Skeleton
	msgDispathcer *event.CMsgDispatcher
	eventRouter   *event.CEventRouter
	configMgr     *cfg.CConfigMgr
	mapMgr        global.MapMgr
	fightMgr      global.FightMgr
	randSrc       *rand.Rand
	skillMgr      global.SkillMgr
	log 		  *log.Log
	instanceMgr *instance.CInstanceMgr
	mailMgr *mail.CMailMgr
}

func Init() global.ServerGlobal {
	global.ServerG = &CServerGlobal{
		timerMgr:      timer.TimerMgr,
		objectMgr:     object.ObjectMgr,
		state:         global.Server_State_Closed,
		dbEngine:      dbengine.DBEngine,
		msgDispathcer: event.MsgDisPatcher,
		eventRouter:   event.EventRouter,
		configMgr:     cfg.ConfigMgr,
		mapMgr:        maps.MapMgr,
		fightMgr:      fight.FightMgr,
		skillMgr:      skill.SkillMgr,
		//randSrc:       rand.New(rand.NewSource(time.Now().UnixNano())),
		log: &log.Log{
			WriteGoRouterNum:10,
			WriteChanNum:100,
			Platform: conf.Server.PlatformName,
			ServerId: conf.Server.ServerID,
		},
		instanceMgr: instance.InstanceMgr,
		mailMgr: mail.MailMgr,
	}

	return global.ServerG
}

func (sg *CServerGlobal) GetObjectMgr() global.OBjectManager {
	return sg.objectMgr
}

func (sg *CServerGlobal) GetTimerMgr() *timer.CTimerMgr {
	return sg.timerMgr
}

func (sg *CServerGlobal) GetDBEngine() *dbengine.CDBEngine {
	return sg.dbEngine
}

func (sg *CServerGlobal) GetEventRouter() *event.CEventRouter {
	return sg.eventRouter
}

func (sg *CServerGlobal) GetMsgDispatcher() *event.CMsgDispatcher {
	return sg.msgDispathcer
}

func (sg *CServerGlobal) GetConfigMgr() global.ConfigMgr {
	return sg.configMgr
}

func (sg *CServerGlobal) GetSkeleton() *module.Skeleton {
	return sg.skeleton
}

func (sg *CServerGlobal) GetMapMgr() global.MapMgr {
	return sg.mapMgr
}

func (sg *CServerGlobal) GetFightMgr() global.FightMgr {
	return sg.fightMgr
}

func (sg *CServerGlobal) GetMailMgr() global.MailMgr {
	return sg.mailMgr
}

func (sg *CServerGlobal) OnTimer() { //定时器
	//时间处理
	now := time.Now()
	elsp := now.UnixNano() - sg.curTime.UnixNano()
	sg.curTime = now
	//sg.randSrc = rand.New(rand.NewSource(sg.curTime.UnixNano()))

	//角色管理器
	sg.objectMgr.Update(sg.curTime, elsp)

	//副本管理器
	sg.instanceMgr.Update(sg.curTime, elsp)

	sg.mailMgr.Update(sg.curTime, elsp)

	sg.processTime = time.Now().UnixNano() - sg.curTime.UnixNano()
}

func (sg *CServerGlobal) GetCurTime() int64 {
	return int64(sg.curTime.UnixNano())
}

func (sg *CServerGlobal) Start(skeleton *module.Skeleton) {
	if sg.state != global.Server_State_Closed {
		return
	}

	sg.state = global.Server_State_Starting

	sg.skeleton = skeleton

	sg.log.Start(conf.Server.RedisAddrs.LogAddr)

	//redis db
	sg.dbEngine.Start(skeleton, conf.Server.RedisAddrs.DBAddr)

	//csv配置
	sg.configMgr.Start()

	//角色管理器
	sg.objectMgr.Create()

	//地图
	sg.mapMgr.Create()

	//副本
	sg.instanceMgr.Create()

	//邮件
	sg.mailMgr.Create()

	//定时器
	sg.timerMgr.AddTimer(sg, 0)
	sg.timerMgr.Start(skeleton)

	sg.curTime = time.Now()
	sg.state = global.Server_State_Started
}

func (sg *CServerGlobal) Stop() {
	if sg.state != global.Server_State_Started {
		return
	}

	sg.state = global.Server_State_Closing

	sg.mapMgr.Stop()

	//角色管理器
	sg.objectMgr.Stop()

	//定时器
	sg.timerMgr.Stop()

	//redis db
	sg.dbEngine.Close()

	//副本
	sg.instanceMgr.Stop()

	//邮件
	sg.mailMgr.Stop()

	sg.skeleton = nil

	sg.log.Close()

	sg.state = global.Server_State_Closed
}

func (sg *CServerGlobal) GetState() int {
	return sg.state
}

func (sg *CServerGlobal) GetRandSrc() *rand.Rand {
	//return sg.randSrc
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (sg *CServerGlobal) GetSkillMgr() global.SkillMgr {
	return sg.skillMgr
}

func (sg *CServerGlobal) GetLog() *log.Log {
	return sg.log
}
