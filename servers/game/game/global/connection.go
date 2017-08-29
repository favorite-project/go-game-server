package global

import (
	"xianxia/common/dbengine"
)

type Connection interface {
	dbengine.IDBSink
	OnRecv([]byte)
	Send(NetMessage)
	OnClose()
	Close()
	OnAccept()
	IsClosed() bool
	Kick()
}
