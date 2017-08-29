package event

type IMsgSink interface {
	OnRecv(interface{}, []byte) //定时触发
}

type CMsgDispatcher struct {
	MapSink map[uint32]IMsgSink
}

var MsgDisPatcher *CMsgDispatcher

func init() {
	MsgDisPatcher = new(CMsgDispatcher)
	MsgDisPatcher.MapSink = make(map[uint32]IMsgSink)
}

func (md *CMsgDispatcher) Register(ModuleId uint32, Sink IMsgSink) bool {
	_, ok := md.MapSink[ModuleId]
	if ok {
		return false
	}

	md.MapSink[ModuleId] = Sink
	return true
}

func (md *CMsgDispatcher) UnRegister(ModuleId uint32) {
	_, ok := md.MapSink[ModuleId]
	if ok {
		delete(md.MapSink, ModuleId)
	}
}

func (md *CMsgDispatcher) Dispatch(ModuleId uint32, obj interface{}, msg []byte) bool {
	sink, ok := md.MapSink[ModuleId]
	if !ok {
		return false
	}

	sink.OnRecv(obj, msg)

	return true
}
