package instance

import (
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/conf"
	"fmt"
	"encoding/json"
	"xianxia/servers/game/msg"
)

func (mgr *CInstanceMgr) msg_enter(player global.Player, recvData []byte) {
	if player == nil || recvData == nil || len(recvData) < 4 {
		return
	}

	instanceId := int32(conf.RdWrEndian.Uint32(recvData))
	if _, ok := mgr.mPlayerInstanceData[player.GetDBId()]; !ok {
		return
	}

	instanceData := mgr.mPlayerInstanceData[player.GetDBId()]
	icfg := global.ServerG.GetConfigMgr().GetCfg("Instance", instanceId)
	if icfg == nil {
		return
	}

	cfg := icfg.(*global.InstanceCfg)

	m := &msg.GSCL_EnterInstance {
		Suc: true,
	}

	if player.GetInstanceMapId() != 0 {
		m.Suc = false
		player.GetConnection().Send(m)
		return
	}

	//是否打成开启条件
	mapId, _ := player.GetMapId()
	if cfg.OpenMapId > mapId {
		m.Suc = false
		player.GetConnection().Send(m)
		return
	}

	useCount := int32(0)
	if _, ok := instanceData.MFreeCount[instanceId]; ok {
		useCount = instanceData.MFreeCount[instanceId]
	}

	if useCount >= cfg.FreeCount + player.GetVipEffectValue(global.Vip_Effect_AddInstanceCount){
		//判断道具
		_, _, _, err := player.AddItem(cfg.NeedItemId, -1, true, true)
		if err != nil {
			m.Suc = false
			player.GetConnection().Send(m)
			return
		}
	} else {
		instanceData.MFreeCount[instanceId] = useCount+1
	}

	player.SetInstanceFightIndex(0)
	player.SetInstanceMapId(instanceId)

	//更新db
	dbv, _:= json.Marshal(instanceData)
	global.ServerG.GetDBEngine().Request(mgr, Instance_DB_Set, player.GetDBId(), "setex",
		fmt.Sprintf("instance:%d", player.GetDBId()), mgr.getExpireTime(), string(dbv))

	player.GetConnection().Send(m)
}