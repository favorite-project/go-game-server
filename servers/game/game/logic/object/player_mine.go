package object

import (
	"xianxia/servers/game/conf"
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/game/global"
	"time"
	"strings"
	"strconv"
	"encoding/json"
	"xianxia/servers/game/utils"
	"xianxia/servers/game/msg"
)

func (p *player) mineCheck(pMineInfo *global.PlayerMineInfo) {
	if pMineInfo == nil {
		return
	}

	now := time.Now().Unix()
	passDay := utils.CheckIsSameDayBySec(now, int64(pMineInfo.LastUpdateTime), 0)
	if passDay {
		return
	}

	for _, mine := range pMineInfo.Mines {
		mine.Works = make(map[int32]int32)
	}

	pMineInfo.LastUpdateTime = int32(now)
}

func (p *player) mine_buy(recvData []byte) {
	if len(recvData) < 4 {
		return
	}

	cfgId := int32(conf.RdWrEndian.Uint32(recvData))
	cfg := global.ServerG.GetConfigMgr().GetCfg("Mine", cfgId)
	if cfg == nil {
		return
	}

	icfg := cfg.(*global.MineCfg)
	if len(icfg.BuyItems) > 0 {
		items := &global.RewardData{
			make(map[int32]*global.RewardItem),
		}
		arr1 := strings.Split(icfg.BuyItems, ";")
		for _, itemNumStr := range arr1 {
			arr2 := strings.Split(itemNumStr, "+")
			if len(arr2) != 2 {
				log.Error("mine_buy cfgId:%d format error", cfgId)
				return
			}

			itemId, err := strconv.Atoi(arr2[0])
			if err != nil {
				log.Error("mine_buy cfgId:%d strconv.Atoi(arr2[0]) error:%s", cfgId, err)
				return
			}

			num, err := strconv.Atoi(arr2[1])
			if err != nil {
				log.Error("mine_buy cfgId:%d strconv.Atoi(arr2[1]) error:%s", cfgId, err)
				return
			}

			if p.GetBagItemNum(int32(itemId)) < int32(num) {
				return
			}

			items.Items[int32(itemId)] = &global.RewardItem{
				Id:int32(itemId),
				Num:int32(-num),
			}
		}

		p.AddItems(items, true, true)
	}

	pMineInfo, err := p.playerInfo.GetPlayerMineInfo(p.dbId)
	if err != nil {
		log.Error("player::mine_buy playerId:%d GetPlayerMineInfo error:%s", p.dbId, err)
		return
	}

	for _, mine := range pMineInfo.Mines {
		if mine.CfgId == cfgId {
			return
		}
	}

	now := int32(time.Now().Unix())
	uniqueId, err := global.ServerG.GetDBEngine().GetUniqueID()
	if err != nil {
		log.Error("player::mine_buy playerId:%d GetUniqueID error:%s", p.dbId, err)
		return
	}

	newMine := &global.PlayerMine{
		Id: int32(uniqueId),
		CfgId:cfgId,
		LastCalcTime:now,
		Works:make(map[int32]int32),
		RobNum:0,
		Msgs:make([]*global.PlayerMineMsg, 0),
	}

	pMineInfo.Mines = append(pMineInfo.Mines, newMine)

	//检查一下
	p.mineCheck(pMineInfo)

	if err := p.playerInfo.SetPlayerMineInfo(p.dbId, pMineInfo); err != nil {
		log.Error("player::mine_buy playerId:%d SetPlayerMineInfo error:%s", p.dbId, err)
	}

	//塞入列表里
	node := &global.MinePoolNode{
		PlayerId:p.dbId,
		PlayerName:string(p.GetName()),
		CfgId:cfgId,
	}

	j, _ := json.Marshal(node)
	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_MINE_POOL, 0, "lpush", global.Mine_Pool_Key, j)
}

const (
	Player_Mine_Work_Ret_Suc  = int32(1) + iota
	Player_Mine_Work_Ret_CountFull
	Player_Mine_Work_Ret_BagFull
)

func (p *player) mine_work(recvData []byte) {
	if len(recvData) < 20 {
		return
	}

	dstDbId := int64(conf.RdWrEndian.Uint64(recvData))
	cfgId := int32(conf.RdWrEndian.Uint32(recvData[8:]))
	workId := int32(conf.RdWrEndian.Uint32(recvData[12:]))
	workNum := int32(conf.RdWrEndian.Uint32(recvData[16:]))
	if dstDbId == p.dbId ||  workNum <= 0 {
		return
	}

	cfg := global.ServerG.GetConfigMgr().GetCfg("Mine", cfgId)
	if cfg == nil {
		return
	}
	icfg := cfg.(*global.MineCfg)

	found := false
	if len(icfg.Works) > 0 {
		cfgWorkArr :=  strings.Split(icfg.Works, "+")
		for _, work := range cfgWorkArr {
			workDataId, err := strconv.Atoi(work)
			if err != nil {
				log.Error("MineCfg workId:%d strconv.Atoi error:%s", workDataId, err)
				return
			}

			if workDataId == int(workId) {
				found = true
				break
			}
		}
	}

	if !found {
		return
	}

	wcfg := global.ServerG.GetConfigMgr().GetCfg("MineWork", workId)
	if wcfg == nil {
		return
	}

	wicfg := wcfg.(*global.MineWorkCfg)

	pMineInfo, err := p.playerInfo.GetPlayerMineInfo(dstDbId)
	if err != nil {
		log.Error("player::mine_work playerId:%d dstPlayerId:%d GetPlayerMineInfo error:%s", p.dbId, dstDbId, err)
		return
	}

	//检查一下
	p.mineCheck(pMineInfo)

	var pMine *global.PlayerMine = nil
	for _, mine := range pMineInfo.Mines {
		if mine.CfgId == cfgId {
			pMine = mine
			break
		}
	}

	if pMine == nil {
		return
	}

	m := &msg.GSCL_MineWork{
		CfgId:cfgId,
		Ret:Player_Mine_Work_Ret_Suc,
		MWorkCounts:pMine.Works,
	}

	if workedNum, ok:= pMine.Works[workId]; ok {
		if workedNum >= wicfg.MaxCount {
			m.Ret = Player_Mine_Work_Ret_CountFull
			p.conn.Send(m)
			return
		}
	}

	if p.IsBagFull(icfg.ProductItemId, wicfg.ItemNum * workNum, true) {
		m.Ret = Player_Mine_Work_Ret_BagFull
		p.conn.Send(m)
		return
	}

	_, _, sellItems, _ := p.AddItem(wicfg.ItemId, wicfg.ItemNum * workNum, true, true)//检查背包

	if workedNum, ok:= pMine.Works[workId]; ok {
		if workedNum + workNum >  wicfg.MaxCount {
			workNum =  wicfg.MaxCount - workedNum
		}
		pMine.Works[workId] += workNum
	} else {
		if workNum > wicfg.MaxCount {
			workNum =  wicfg.MaxCount
		}
		pMine.Works[workId] = workNum
	}

	mItems := make(map[int32]int32)
	mItems[wicfg.ItemId] = wicfg.ItemNum * workNum

	mSellItems := make(map[int32]int32)
	for _, sItem := range sellItems {
		mSellItems[sItem.CfgId] = sItem.Num
	}

	m.Items = mItems
	m.SellItems = mSellItems
	p.conn.Send(m)

	now := int32(time.Now().Unix())
	mmsg := &global.PlayerMineMsg{
		Type: global.Mine_Msg_Type_Work,
		Num: workNum,
		PlayerId: p.dbId,
		WorkId: workId,
		Time:now,
	}

	pMine.Msgs = append(pMine.Msgs, mmsg)
	if len(pMine.Msgs) > global.Mine_Msg_Max_Len {
		pMine.Msgs = pMine.Msgs[len(pMine.Msgs)-global.Mine_Msg_Max_Len:]
	}

	if err := p.playerInfo.SetPlayerMineInfo(dstDbId, pMineInfo); err != nil {
		log.Error("player::mine_work playerId:%d dstDbID:%d SetPlayerMineInfo error:%s", p.dbId, dstDbId, err)
	}
}

func (p *player) mine_reward(recvData []byte) {
	if len(recvData) < 4 {
		return
	}

	cfgId := int32(conf.RdWrEndian.Uint32(recvData))
	cfg := global.ServerG.GetConfigMgr().GetCfg("Mine", cfgId)
	if cfg == nil {
		return
	}

	icfg := cfg.(*global.MineCfg)
	pMineInfo, err := p.playerInfo.GetPlayerMineInfo(p.dbId)
	if err != nil {
		log.Error("player::mine_reward playerId:%d GetPlayerMineInfo error:%s", p.dbId, err)
		return
	}

	//检查一下
	p.mineCheck(pMineInfo)

	var pMine *global.PlayerMine = nil
	for _, mine := range pMineInfo.Mines {
		if mine.CfgId == cfgId {
			pMine = mine
			break
		}
	}

	if pMine == nil {
		return
	}

	now := int32(time.Now().Unix())
	passSec := now - pMine.LastCalcTime
	if passSec > icfg.FullSec {
		passSec = icfg.FullSec
	}

	if passSec <= 0 {
		return
	}

	addNum := passSec * icfg.PerSecNum
	addNum -= pMine.RobNum
	if addNum <= 0 {
		log.Error("Player::mine_reward PlayerId:%d addNum <= 0", p.dbId)
		return
	}

	//检查背包
	if p.IsBagFull(icfg.ProductItemId, addNum, true) {
		m := &msg.GSCL_Error{
			Desc:[]byte("背包已满"),
		}

		p.conn.Send(m)
		return
	}

	_, _, sellItems, _ := p.AddItem(icfg.ProductItemId, addNum, true, true)

	mItems := make(map[int32]int32)
	mItems[icfg.ProductItemId] = addNum

	mSellItems := make(map[int32]int32)
	for _, sItem := range sellItems {
		mSellItems[sItem.CfgId] = sItem.Num
	}

	m := &msg.GSCL_MineReward{
		CfgId:cfgId,
		Items:mItems,
		SellItems:mSellItems,
	}
	p.conn.Send(m)

	pMine.LastCalcTime = now
	pMine.RobNum = 0
	if err := p.playerInfo.SetPlayerMineInfo(p.dbId, pMineInfo); err != nil {
		log.Error("player::mine_reward playerId:%d SetPlayerMineInfo error:%s", p.dbId, err)
	}
}

