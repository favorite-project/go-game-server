package event

type IEventListener interface {
	OnEvent(event *CEvent) //定时触发
}

type CEvent struct {
	Type    int         //类型
	Obj     interface{} //发送对象
	Content interface{} //发送内容
}

type MapListener map[IEventListener]struct{}

type CEventRouter struct {
	MapEvents map[int]MapListener
}

var EventRouter *CEventRouter

func init() {
	EventRouter = new(CEventRouter)
	EventRouter.MapEvents = make(map[int]MapListener)
}

func (er *CEventRouter) AddEventListener(evType int, listener IEventListener) bool {
	if _, ok := er.MapEvents[evType]; !ok {
		er.MapEvents[evType] = make(MapListener)
	}

	if _, ok := er.MapEvents[evType][listener]; ok {
		return false
	}

	er.MapEvents[evType][listener] = struct{}{}
	return true
}

func (er *CEventRouter) DelEventListener(evType int, listener IEventListener) {
	if _, ok := er.MapEvents[evType]; !ok {
		return
	}

	if _, ok := er.MapEvents[evType][listener]; !ok {
		return
	}

	delete(er.MapEvents[evType], listener)
}

func (er *CEventRouter) DoEvent(evType int, obj interface{}, content interface{}) {
	if _, ok := er.MapEvents[evType]; !ok {
		return
	}

	event := &CEvent{
		Type:    evType,
		Obj:     obj,
		Content: content,
	}

	for l, _ := range er.MapEvents[evType] {
		l.OnEvent(event)
	}
}
