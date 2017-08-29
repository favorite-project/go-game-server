package internal

import (
	"github.com/name5566/leaf/module"
	"xianxia/servers/game/base"
	"xianxia/servers/game/game/logic"
	"xianxia/servers/game/game/http"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer
	ServerG  = logic.Init()
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton
	ServerG.Start(skeleton)
	http.Start(skeleton)
}

func (m *Module) Run(sig chan bool) {
	skeleton.Run(sig)
}

func (m *Module) OnDestroy() {
	ServerG.Stop()
}
