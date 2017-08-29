package internal

import (
	"github.com/name5566/leaf/gate"
	"xianxia/servers/game/conf"
	"xianxia/servers/game/game"
	"xianxia/servers/game/msg"
)

type Module struct {
	*gate.Gate
}

func (m *Module) OnInit() {
	m.Gate = &gate.Gate{
		MaxConnNum:      conf.Server.MaxConnNum,
		PendingWriteNum: conf.PendingWriteNum,
		MaxMsgLen:       conf.MaxMsgLen,
		WSAddr:          conf.Server.WSAddr,
		HTTPTimeout:     conf.HTTPTimeout,
		TCPAddr:         conf.Server.TCPAddr,
		LenMsgLen:       conf.LenMsgLen,
		LittleEndian:    conf.LittleEndian,
		Processor:       msg.Processor,
		AgentChanRPC:    game.ChanRPC,
	}
}

func (m *Module) Run(closeSig chan bool) {
	m.Gate.Run(closeSig)
}

func (m *Module) OnDestroy() {
	m.Gate.OnDestroy()
}
