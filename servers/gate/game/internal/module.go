package internal

import (
	"github.com/name5566/leaf/module"
	"xianxia/servers/gate/base"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton
}

func (m *Module) Run(sig chan bool) {
	skeleton.Run(sig)
}

func (m *Module) OnDestroy() {

}
