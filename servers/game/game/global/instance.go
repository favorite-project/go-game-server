package global

import (
	"xianxia/common/event"
	"xianxia/common/dbengine"
)

const Instance_Reset_Hour = 0

type InstanceMgr interface {
	Singleton
	OnRecv(interface{}, []byte)
	OnEvent(event *event.CEvent)
	OnRet(ret *dbengine.CDBRet)
	Create() bool
	Stop() bool
}