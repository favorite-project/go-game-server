package instance
import (
	"xianxia/common/event"
	"xianxia/servers/game/game/global"
	"xianxia/common/dbengine"
	"xianxia/servers/game/game/global/db"
	"github.com/name5566/leaf/log"
	"fmt"
	"xianxia/servers/game/utils"
	"time"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"xianxia/servers/game/conf"
	"xianxia/servers/game/msg"
)

var InstanceMgr *CInstanceMgr
func init() {
	InstanceMgr = &CInstanceMgr{
		mPlayerInstanceData:make(map[int64]*db.Player_InstanceDB_Data),
	}
}

const (
	Instance_DB_Get = 1
	Instance_DB_Set = 2
)


type CInstanceMgr struct {
	mPlayerInstanceData map[int64]*db.Player_InstanceDB_Data
	m_lastUpdateTime *time.Time
}

func (mgr *CInstanceMgr) Create() bool {
	//注册消息
	global.ServerG.GetMsgDispatcher().Register(global.Message_RootKey_Instance, mgr)

	//注册事件
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOnline, mgr) //上线事件
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOffline, mgr)//下线事件

	return true
}

func (mgr *CInstanceMgr) Stop() bool {
	global.ServerG.GetMsgDispatcher().UnRegister(global.Message_RootKey_Instance)

	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOnline, mgr)
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOffline, mgr)

	return true
}


func (mgr *CInstanceMgr) OnRecv(obj interface{}, recvData []byte) {
	player, ok := obj.(global.Player)
	if !ok {
		return
	}

	if len(recvData) < 4 {
		return
	}

	subModule := conf.RdWrEndian.Uint32(recvData)
	switch (subModule) {
	case global.Message_RootKey_Instance_Enter:
		mgr.msg_enter(player, recvData[4:])
	}
}

func (mgr *CInstanceMgr) Update(now time.Time, elspNanoSecond int64) {
	//檢查是否跨天
	if mgr.m_lastUpdateTime != nil && !utils.CheckIsSameDay(mgr.m_lastUpdateTime, &now, global.Instance_Reset_Hour) {
		m := &msg.GSCL_PlayerInstanceInfo{}
		for dbid, _ := range mgr.mPlayerInstanceData {
			mgr.mPlayerInstanceData[dbid] = mgr.initPlayerInstanceDBData(dbid)

			player := global.ServerG.GetObjectMgr().GetPlayer(dbid)
			if player != nil {
				player.GetConnection().Send(m)
			}
		}
	}

	mgr.m_lastUpdateTime = &now
}

func (mgr *CInstanceMgr) OnEvent(event *event.CEvent) {
	if event == nil {
		return
	}

	switch(event.Type) {
	case global.Event_Type_PlayerOnline:
		if event.Obj == nil {
			return
		}

		player, ok := event.Obj.(global.Player)
		if !ok {
			return
		}

		global.ServerG.GetDBEngine().Request(mgr, Instance_DB_Get, player.GetDBId(),"get", fmt.Sprintf("instance:%d", player.GetDBId()))
	case global.Event_Type_PlayerOffline:
		if event.Obj == nil {
			return
		}

		player, ok := event.Obj.(global.Player)
		if !ok {
			return
		}

		//下线就离开副本
		player.SetInstanceMapId(int32(0))

		if _, ok := mgr.mPlayerInstanceData[player.GetDBId()];ok {
			delete(mgr.mPlayerInstanceData, player.GetDBId())
		}
	}
}

func (mgr *CInstanceMgr) OnRet(ret *dbengine.CDBRet) {
	if ret.Err != nil {
		log.Error("CInstanceMgr::OnRet error:%s", ret.Err)
		return
	}

	switch(ret.OpType) {
	case Instance_DB_Set:
	case Instance_DB_Get:
		player := global.ServerG.GetObjectMgr().GetPlayer(ret.DBId)
		if player == nil || !player.IsOnline() {
			return
		}

		//下发到客户端
		var instanceData *db.Player_InstanceDB_Data
		if nil == ret.Content {
			instanceData = mgr.initPlayerInstanceDBData(ret.DBId)
		} else {
			value, err := redis.String(ret.Content, nil)
			if err != nil {
				log.Error("CInstanceMgr::OnRet Instance_DB_Get content:%s redis.String error:%s", ret.Content, err)
				return
			}

			instanceData = &db.Player_InstanceDB_Data{}
			err = json.Unmarshal([]byte(value), instanceData)
			if err != nil {
				log.Error("CInstanceMgr::OnRet Instance_DB_Get content:%s json.Unmarshal error:%s", ret.Content, err)
				return
			}
		}

		mgr.mPlayerInstanceData[ret.DBId] = instanceData

		//下发客户端
		m := &msg.GSCL_PlayerInstanceInfo {
			Player_InstanceDB_Data:instanceData,
		}
		player.GetConnection().Send(m)
	}
}

func (mgr *CInstanceMgr) initPlayerInstanceDBData(dbId int64) *db.Player_InstanceDB_Data {
	return &db.Player_InstanceDB_Data{
		DbId:dbId,
		MFreeCount:make(map[int32]int32),
	}
}

//0点过期
func (mgr *CInstanceMgr) getExpireTime() int64 {
	now := time.Now()
	endSec, _ := utils.GetTodayEndUnixInHour(&now, global.Instance_Reset_Hour)
	return endSec - now.Unix()
}