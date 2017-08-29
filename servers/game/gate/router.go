package gate

import (
	"xianxia/servers/game/game"
	"xianxia/servers/game/msg"
)

func init() {
	msg.Processor.SetRouter([]byte(""), game.ChanRPC)
}
