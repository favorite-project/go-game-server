package timer

import (
	"github.com/name5566/leaf/module"
	"time"
)

type ITimer interface {
	OnTimer() //定时触发
}

type TimerData struct {
	Count    int64
	RunCount int64
}

type CTimerMgr struct {
	MapTimers map[ITimer]*TimerData
	Skeleton  *module.Skeleton
	Run       bool
}

var TimerMgr *CTimerMgr

func init() {
	TimerMgr = new(CTimerMgr)
	TimerMgr.MapTimers = make(map[ITimer]*TimerData)
	TimerMgr.Run = false
}

func (tm *CTimerMgr) Start(skeleton *module.Skeleton) {
	if skeleton == nil {
		return
	}

	tm.Skeleton = skeleton
	tm.Run = true
	skeleton.AfterFunc(time.Millisecond * 50, tm.run)
}

func (tm *CTimerMgr) Stop() {
	tm.Run = false
	tm.Skeleton = nil
}

func (tm *CTimerMgr) run() {
	for timer, data := range tm.MapTimers {
		timer.OnTimer()

		data.RunCount++
		if data.Count != 0 && data.RunCount >= data.Count {
			delete(tm.MapTimers, timer)
		}
	}

	if tm.Run {
		tm.Skeleton.AfterFunc(time.Microsecond, tm.run) //1微秒执行一次
	}
}

func (tm *CTimerMgr) AddTimer(timer ITimer, count int64) {
	if _, ok := tm.MapTimers[timer]; ok {
		return
	}

	tm.MapTimers[timer] = &TimerData{Count: count, RunCount: 0}
}

func (tm *CTimerMgr) DelTimer(timer ITimer) {
	if _, ok := tm.MapTimers[timer]; !ok {
		return
	}

	delete(tm.MapTimers, timer)
}
