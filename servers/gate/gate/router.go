package gate

import (
	"xianxia/servers/gate/game"
	"xianxia/servers/gate/msg"
)

func init() {
	msg.Processor.SetRouter([]byte{}, game.ChanRPC)
}
