package maps

import (
	"github.com/name5566/leaf/log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"xianxia/servers/game/game/global"
)

type CMap struct {
	id      int32
	players map[int64]int
	bInstance bool
}

func (m *CMap) GetPlayerList() []int64 {
	list := make([]int64, len(m.players))
	i := 0
	for dbId, _ := range m.players {
		list[i] = dbId
		i += 1
	}

	return list
}

func (m *CMap) Update(now time.Time, elspNanoSecond int64) {

}

func (m *CMap) GetRegionMonsters(regionId int32, bBoss bool) []global.Monster {
	cfg := global.ServerG.GetConfigMgr().GetCfg("Map", m.id)
	if cfg == nil {
		log.Error("Map: %d GetRegionMonsters Config nil", m.id)
		return nil
	}
	config, _ := cfg.(*global.MapCfg)

	reArr := strings.Split(config.Regions, "+")
	found := false
	for _, v := range reArr {
		iv, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Map: %d GetRegionMonsters Regions Atoi Error:%v", m.id, err)
			return nil
		}

		if int32(iv) == regionId {
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	rcfg := global.ServerG.GetConfigMgr().GetCfg("Region", regionId)
	if rcfg == nil {
		log.Error("Map: %d, Region:%d GetRegionMonsters Config nil", m.id, regionId)
		return nil
	}
	rconfig, _ := rcfg.(*global.RegionCfg)

	var monArr []string
	if bBoss {
		monArr = strings.Split(rconfig.BossId, "+")
	} else {
		monArr = strings.Split(rconfig.MonData, "+")
	}

	if monArr == nil || len(monArr) == 0 {
		log.Error("Map: %d, Region:%d GetRegionMonsters Regions Format Error", m.id, regionId)
		return nil
	}

	choseMonArr := monArr
	if !bBoss {
		monNum := rconfig.MonNum
		if monNum > int32(len(monArr)) {
			monNum = int32(len(monArr))
		}

		if monNum == 0 {
			log.Error("Map: %d, Region:%d GetRegionMonsters Regions MonsterNum error", m.id, regionId)
			return nil
		}

		randMon := rand.New(rand.NewSource(time.Now().UnixNano()))
		choseMonArr = make([]string, monNum)
		index := 0
		for {
			if index == int(monNum) {
				break
			}

			ri := randMon.Intn(len(monArr))
			choseMonArr[index] = monArr[ri]
			index++
			monArr = append(monArr[:ri], monArr[ri+1:]...)
		}
	}

	monObjArr := make([]global.Monster, len(choseMonArr))
	ObjectMgr := global.ServerG.GetObjectMgr()

	for i, v := range choseMonArr {
		iv, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Map: %d, Region:%d GetRegionMonsters Regions Atoi Error:%v", m.id, regionId, err)
			return nil
		}

		micfg := global.ServerG.GetConfigMgr().GetCfg("Monster", int32(iv))
		if micfg == nil {
			log.Error("Map: %d, Region:%d Monster: %d GetRegionMonsters Config nil", m.id, regionId, iv)
			return nil
		}

		mconfig := micfg.(*global.MonsterCfg)
		monObjArr[i] = ObjectMgr.CreateMonster(mconfig)
	}

	return monObjArr
}

func(m *CMap) GetInstanceMonsters(fightIndex int) ([]global.Monster, bool) {
	if fightIndex < 0 {
		log.Error("Map: %d GetInstanceMonsters FightIndex error1", m.id)
		return nil, false
	}

	cfg := global.ServerG.GetConfigMgr().GetCfg("Instance", m.id)
	if cfg == nil {
		log.Error("Map: %d GetInstanceMonsters Config nil", m.id)
		return nil, false
	}
	config, _ := cfg.(*global.InstanceCfg)
	if len(config.MonData) == 0 {
		log.Error("Map: %d GetInstanceMonsters Config MonData empty", m.id)
		return nil, false
	}

	monsStrArr := strings.Split(config.MonData, ";")
	if fightIndex >= len(monsStrArr) {
		log.Error("Map: %d GetInstanceMonsters FightIndex error2", m.id)
		return nil, false
	}

	bBoss := false
	if  fightIndex == len(monsStrArr) - 1 {
		bBoss = true
	}

	monsIndexStr := monsStrArr[fightIndex]
	if len(monsIndexStr) == 0 {
		log.Error("Map: %d GetInstanceMonsters FightIndex:%d mons empty", m.id, fightIndex)
		return nil, false
	}

	monsIndexArr := strings.Split(monsIndexStr, "+")
	monObjArr := []global.Monster{}
	ObjectMgr := global.ServerG.GetObjectMgr()
	for _, v := range monsIndexArr {
		iv, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Map: %d GetInstanceMonsters Monster:%s Atoi Error:%v", m.id, v, err)
			return nil, false
		}

		micfg := global.ServerG.GetConfigMgr().GetCfg("Monster", int32(iv))
		if micfg == nil {
			log.Error("Map: %d, GetInstanceMonsters Monster:%d GetRegionMonsters Config nil", m.id, iv)
			return nil, false
		}

		mconfig := micfg.(*global.MonsterCfg)
		monObjArr = append(monObjArr, ObjectMgr.CreateMonster(mconfig))
	}

	return monObjArr, bBoss
}

func (m *CMap) IsInstance() bool {
	return m.bInstance
}

func (m *CMap) playerCome(dbId int64) {
	delete(m.players, dbId)
}

func (m *CMap) playerLeave(dbId int64) {
	m.players[dbId] = 1
}
