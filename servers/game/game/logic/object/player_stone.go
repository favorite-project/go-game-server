package object

import (
	"fmt"
	"xianxia/servers/game/game/errorx"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/msg"
	"xianxia/servers/game/utils"
)

// 赌石
func (p *player) RandomStone(stoneMode int32, timesMode int) (rewardItem map[int32]int32, err error) {
	if stoneMode != 1 && stoneMode != 2 && stoneMode != 3 {
		err = errorx.WRONG_PARAMETER
		return
	}

	if timesMode < 0 && timesMode > 10 {
		err = errorx.WRONG_PARAMETER
		return
	}

	pmoney, ok := p.GetProp(global.Player_Prop_Money)
	if !ok {
		err = errorx.READ_PROPS_ERR
		return
	}

	randomStone := global.ServerG.GetConfigMgr().GetCfg("RandomStone", stoneMode).(*global.RandomStone)
	if randomStone == nil {
		err = errorx.CSV_CFG_EMPTY
		return
	}

	needCost := randomStone.Price * int32(timesMode)
	if pmoney < needCost {
		err = errorx.MONEY_NOT_ENOUGH
		return
	}

	_, ok = p.SetProp(global.Player_Prop_Money, -needCost, true)
	if !ok {
		err = errorx.SET_MONEY_PROPS_ERR
		return
	}

	cntMap := map[int32]int32{}
	for i := 0; i < timesMode; i++ {
		itemId, err := utils.GetResult(randomStone.Total, randomStone.Rate)
		if err != nil {
			continue
		}

		if itemId != 0 {
			if _, ok := cntMap[itemId]; !ok {
				cntMap[itemId] = 1
			} else {
				cntMap[itemId]++
			}
			p.AddItem(itemId, 1, true, false)
		}
	}
	return cntMap, nil
}

// 拉去赌石配置
func (p *player) GetRandomStoneCfg() (m *msg.GSCL_RandomStoneCfg, err error) {
	epCsv := global.ServerG.GetConfigMgr().GetCsv("RandomStone")
	if epCsv == nil {
		err = errorx.CSV_CFG_EMPTY
		return
	}
	m = &msg.GSCL_RandomStoneCfg{}
	for i := 0; i < epCsv.NumRecord(); i++ {
		epicfg := epCsv.Record(i)
		stoneCfg, ok := epicfg.(*global.RandomStone)
		if !ok {
			err = errorx.CSV_CFG_EMPTY
			return
		}
		// 由于怕map无序了,这里写死吧
		if stoneCfg.Id == 1 {
			m.PrimaryPrice = stoneCfg.Price
		}

		if stoneCfg.Id == 2 {
			m.MiddlePrice = stoneCfg.Price
		}

		if stoneCfg.Id == 3 {
			m.HighPrice = stoneCfg.Price
		}
	}

	fmt.Println("发送配置:", m)
	return m, nil
}
