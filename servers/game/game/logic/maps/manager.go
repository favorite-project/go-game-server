package maps

import (
	"errors"
	"github.com/name5566/leaf/log"
	_ "github.com/name5566/leaf/recordfile"
	"strconv"
	"strings"
	"time"
	"xianxia/common/event"
	"xianxia/servers/game/game/global"
)

type CMapMgr struct {
	Maps map[int32]*CMap
	InstanceMaps map[int32]*CMap
}

var MapMgr global.MapMgr

func init() {
	MapMgr = &CMapMgr{
		Maps: make(map[int32]*CMap),
		InstanceMaps: make(map[int32]*CMap),
	}
}

func (mgr *CMapMgr) Create() bool {
	//注册事件
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOnline, mgr)
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOffline, mgr)
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_ChangeMap, mgr)

	//读取所有常规地图
	csv := global.ServerG.GetConfigMgr().GetCsv("Map")
	if csv == nil {
		log.Error("CMapMgr Map GetCsv Config nil")
		return false
	}

	maps := csv.Indexes(0)
	if maps == nil {
		log.Error("CMapMgr Map GetCsv Config Indexes(0) nil")
		return false
	}

	for mapId, _ := range maps {
		mgr.Maps[mapId.(int32)] = &CMap{
			id:      mapId.(int32),
			players: make(map[int64]int),
			bInstance:false,
		}
	}

	//读取所有副本地图
	ins_csv := global.ServerG.GetConfigMgr().GetCsv("Instance")
	if ins_csv == nil {
		log.Error("CMapMgr Instance GetCsv Config nil")
		return false
	}

	ins_maps := ins_csv.Indexes(0)
	if ins_maps == nil {
		log.Error("CMapMgr Instance GetCsv Config Indexes(0) nil")
		return false
	}

	for mapId, _ := range ins_maps {
		mgr.InstanceMaps[mapId.(int32)] = &CMap{
			id:      mapId.(int32),
			players: make(map[int64]int),
			bInstance:true,
		}
	}

	return true
}

func (mgr *CMapMgr) Stop() bool {
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOnline, mgr)
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOffline, mgr)
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_ChangeMap, mgr)
	return true
}

//统一循环调度
func (mgr *CMapMgr) Update(now time.Time, elspNanoSecond int64) {
	for _, cmap := range mgr.Maps {
		cmap.Update(now, elspNanoSecond)
	}
}

func (mgr *CMapMgr) GetMap(id int32) global.Map {
	cmap, ok := mgr.Maps[id]
	if ok {
		return cmap
	}

	cmap, ok = mgr.InstanceMaps[id]
	if ok {
		return cmap
	}

	return nil
}

func (mgr *CMapMgr) OnEvent(ev *event.CEvent) {
	if ev == nil || ev.Obj == nil {
		return
	}

	switch ev.Type {
	case global.Event_Type_PlayerOffline:
		player, ok := ev.Obj.(global.Player)
		if !ok {
			return
		}

		mapId, _ := player.GetMapId()
		if m, ok := mgr.Maps[mapId]; ok {
			m.playerLeave(player.GetDBId())
		}
	case global.Event_Type_PlayerOnline:
		player, ok := ev.Obj.(global.Player)
		if !ok {
			return
		}

		mapId, _ := player.GetMapId()
		if m, ok := mgr.Maps[mapId]; ok {
			m.playerCome(player.GetDBId())
		}
	case global.Event_Type_ChangeMap:
		player, ok := ev.Obj.(global.Player)
		if !ok {
			return
		}

		cmEvent, ok := ev.Content.(*global.ChangeMap_Event_Info)
		if !ok {
			return
		}

		mapId := cmEvent.OMapId
		if m, ok := mgr.Maps[mapId]; ok {
			m.playerLeave(player.GetDBId())
		}

		mapId = cmEvent.MapId
		if m, ok := mgr.Maps[mapId]; ok {
			m.playerCome(player.GetDBId())
		}
	}
}

func (m *CMapMgr) NextMap(player global.Player) bool {
	if player == nil {
		return false
	}

	//切换地图
	mid, rid := player.GetMapId()
	cfg := global.ServerG.GetConfigMgr().GetCfg("Map", mid)
	if cfg == nil {
		log.Error("CMapMgr Player: %d NextMap get Map: %d error", player.GetDBId(), mid)
		return false
	}

	config, _ := cfg.(*global.MapCfg)
	reArr := strings.Split(config.Regions, "+")

	maxMapId, _ := player.GetProp(global.Player_Prop_MaxMapId)
	maxRegionId, _ := player.GetProp(global.Player_Prop_MaxRegionId)
	found := false
	foundMaxRid := false
	nextRegionId := 0
	for index, v := range reArr {
		iv, err := strconv.Atoi(v)
		if err != nil {
			continue
		}

		if int32(iv) == maxRegionId {
			foundMaxRid = true
		}

		if int32(iv) == rid {
			//拿到本地图下一个区域id
			if index < len(reArr)-1 {
				nextRegionId, err = strconv.Atoi(reArr[index+1])
				if err != nil {
					break
				}
			}

			found = true
			break
		}
	}

	if !found {
		return false
	}

	cmEvent := &global.ChangeMap_Event_Info{
		Player:    player.GetDBId(),
		OMapId:    mid,
		ORegionId: rid,
		MapId:     mid,
		RegionId:  int32(nextRegionId),
	}

	if nextRegionId == 0 {
		if config.NextMapId > 0 {
			cfg = global.ServerG.GetConfigMgr().GetCfg("Map", config.NextMapId)
			if cfg != nil {
				config, ok := cfg.(*global.MapCfg)
				if !ok {
					return false
				}

				reArr = strings.Split(config.Regions, "+")
				if len(reArr) == 0 {
					return false
				}

				//第一个区域
				nextRegionId, err := strconv.Atoi(reArr[0])
				if err != nil {
					return false
				}

				//检查区域是否有配置
				rcfg := global.ServerG.GetConfigMgr().GetCfg("Region", int32(nextRegionId))
				if rcfg == nil {
					return false
				}

				if config.Id > maxMapId {
					player.SetProp(global.Player_Prop_MaxMapId, config.Id, false)
					player.SetProp(global.Player_Prop_MaxRegionId, int32(nextRegionId), false)
				}

				player.SetMapId(config.Id, int32(nextRegionId))
				cmEvent.MapId = config.Id
				cmEvent.RegionId = int32(nextRegionId)
			}
		}
	} else {
		player.SetMapId(mid, int32(nextRegionId))
		if mid == maxMapId {
			if foundMaxRid {
				player.SetProp(global.Player_Prop_MaxRegionId, int32(nextRegionId), false)
			}
		}
	}

	global.ServerG.GetEventRouter().DoEvent(global.Event_Type_ChangeMap, player, cmEvent)

	return true
}

func (m *CMapMgr) ChangeMap(player global.Player, cMapId int32, cRegionId int32) error {
	if player == nil {
		return errors.New("player nil")
	}

	mid, rid := player.GetMapId()
	if mid == cMapId && cRegionId == rid {
		return errors.New("in curMap curRegion")
	}

	maxMapId, _ := player.GetProp(global.Player_Prop_MaxMapId)
	if cMapId > maxMapId {
		return errors.New("no pass cur Mao error")
	}

	if cMapId == maxMapId {
		maxRegionId, _ := player.GetProp(global.Player_Prop_MaxRegionId)

		cfg := global.ServerG.GetConfigMgr().GetCfg("Map", cMapId)
		if cfg == nil {
			log.Error("CMapMgr Player: %d ChangeMap get Map: %d error", player.GetDBId(), mid)
			return errors.New("get map cfg error")
		}

		config, _ := cfg.(*global.MapCfg)
		reArr := strings.Split(config.Regions, "+")
		// foundMaxRid := false
		found := false
		hadPassRegionId := []int32{}
		for _, v := range reArr {
			av, err := strconv.Atoi(v)
			if err != nil {
				panic(err)
			}

			vv := int32(av)
			if vv > maxRegionId {
				continue
			}
			// 获取到已经通过区域数组
			hadPassRegionId = append(hadPassRegionId, vv)
		}

		for _, v := range hadPassRegionId {
			if v == cRegionId {
				found = true
			}
		}

		if !found {
			return errors.New("get region cfg error")
		}
	}

	//切换地图

	cmEvent := &global.ChangeMap_Event_Info{
		Player:    player.GetDBId(),
		OMapId:    mid,
		ORegionId: rid,
		MapId:     cMapId,
		RegionId:  cRegionId,
	}

	player.SetMapId(cMapId, cRegionId)
	global.ServerG.GetEventRouter().DoEvent(global.Event_Type_ChangeMap, player, cmEvent)

	return nil
}
