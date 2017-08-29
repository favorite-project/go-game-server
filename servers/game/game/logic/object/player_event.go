package object

import (
	"github.com/name5566/leaf/log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/msg"
	"xianxia/servers/game/utils"
	"fmt"
)

func (p *player) generateFightWinReward(dropData string, randItem *rand.Rand) *global.RewardData {
	if len(dropData) == 0 {
		return nil
	}

	if randItem == nil {
		randSrc := rand.NewSource(time.Now().UnixNano())
		randItem = rand.New(randSrc)
	}

	// 掉落盒子id数组
	/*
	monCfgDropData := dropData
	monCfgDropDataArr := strings.Split(monCfgDropData, "+")
	if len(monCfgDropDataArr) == 0 {
		return nil
	}

	rewardItems := &global.RewardData{
		Items:make(map[int32]*global.RewardItem),
	}

	for _, monDropItem := range monCfgDropDataArr {
		dropBoxId, err := strconv.Atoi(monDropItem)
		if err != nil {
			continue
		}

		cfg := global.ServerG.GetConfigMgr().GetCfg("DropBox", int32(dropBoxId))
		if cfg == nil {
			log.Error("OnFightOver Get DropBoxCfg:%d empty", dropBoxId)
			continue
		}

		dcfg, _ := cfg.(*global.DropBoxCfg)
		itemid, num, err:= utils.GetResultNew(10000, dcfg.Reward)
		if itemid == 0 || num == 0 || err != nil {
			continue
		}
		rewardItems.Items[itemid] = &global.RewardItem{Id: itemid, Num: num}
	}
	*/

	monCfgDropDataArr := strings.Split(dropData, ";")
	if len(monCfgDropDataArr) == 0 {
		return nil
	}

	rewardItems := &global.RewardData{
		Items:make(map[int32]*global.RewardItem),
	}

	//1000+1001;2$2000#100-2001#50
	for _, monDropItem := range monCfgDropDataArr {
		dropIds := []int32{}
		if !strings.Contains(monDropItem, "$") {
			arr2 := strings.Split(monDropItem, "+")
			for _, idStr := range arr2 {
				dropBoxId, err := strconv.Atoi(idStr)
				if err != nil {
					log.Error("dropBox strconv.Atoi error:", err)
					return nil
				}

				dropIds = append(dropIds, int32(dropBoxId))
			}
		} else {
			arr2 := strings.Split(monDropItem, "$")
			if len(arr2) != 2 {
				log.Error("dropBox strings.Split len !=2")
				return nil
			}

			num, err := strconv.Atoi(arr2[0])
			if err != nil {
				log.Error("dropBox strconv.Atoi num error:", err)
				return nil
			}

			drops := []struct {
				Id int32
				Per int
			}{}

			totalPer := 0
			arr3 := strings.Split(arr2[1], "-")
			for _, dstr := range arr3 {
				arr4 := strings.Split(dstr, "#")
				if len(arr4) != 2 {
					log.Error("dropBox strings.Split dstr len != 2")
					return nil
				}

				id, err := strconv.Atoi(arr4[0])
				if err != nil {
					log.Error("dropBox trconv.Atoi(arr4[0]) error:", err)
					return nil
				}

				per, err := strconv.Atoi(arr4[1])
				if err != nil {
					log.Error("dropBox trconv.Atoi(arr4[1]) error:", err)
					return nil
				}

				totalPer += per
				drops = append(drops, struct {
					Id int32
					Per int
				}{
					Id:int32(id),
					Per:totalPer,
				})
			}

			for i:=0;i < num;i++ {
				rp := randItem.Intn(10000) + 1
				fmt.Println("randPer:", rp)
				for _, it := range drops {
					if rp <= it.Per {
						dropIds = append(dropIds, it.Id)
						break
					}
				}
			}
		}

		for _, dropBoxId := range  dropIds {
			cfg := global.ServerG.GetConfigMgr().GetCfg("DropBox", dropBoxId)
			if cfg == nil {
				log.Error("OnFightOver Get DropBoxCfg:%d empty", dropBoxId)
				continue
			}

			dcfg, _ := cfg.(*global.DropBoxCfg)
			itemid, num, err:= utils.GetResultNew(10000, dcfg.Reward)
			if itemid == 0 || num == 0 || err != nil {
				continue
			}
			rewardItems.Items[itemid] = &global.RewardItem{Id: itemid, Num: num}
		}
	}

	//检查一下vip金币掉落加成
	if value, ok := rewardItems.Items[global.MONEY_ITEM_ID]; ok {
		rewardItems.Items[global.MONEY_ITEM_ID].Num = value.Num + int32(p.GetVipEffectValue(global.Vip_Effect_AddMoneyDrop) * value.Num / 100)
	}

	//检查一下vip经验掉落加成
	if value, ok := rewardItems.Items[global.EXP_ITEM_ID]; ok {
		rewardItems.Items[global.EXP_ITEM_ID].Num = value.Num + int32(p.GetVipEffectValue(global.Vip_Effect_AddExpDrop) * value.Num / 100)
	}

	return rewardItems
}

func (p *player) OnFightEvent(bAttacker bool, fev *global.Fight_Event_Info) {
	if fev == nil {
		return
	}

	now := global.ServerG.GetCurTime()
	if fev.Mode == global.FIGHT_MODE_NORMAL { //普通战斗
		p.lastNormalFightReard.Items = nil
		p.nextNormalFightTime = now + int64(p.GetVipEffectValue(global.Vip_Effect_FightWaitSec)) + int64(fev.FightRound)*global.FIGHT_ROUND_TIME_SEC
		//todo设置奖励
		if fev.Win {
			//设置下次可战斗时间
			randSrc := rand.NewSource(time.Now().UnixNano())
			randItem := rand.New(randSrc)

			p.lastNormalFightReard.Items = make(map[int32]*global.RewardItem)
			for _, c := range fev.Defencers {
				mon, ok := c.(global.Monster)
				if !ok {
					continue
				}

				cfg := global.ServerG.GetConfigMgr().GetCfg("Monster", mon.GetCfgId())
				if cfg == nil {
					continue
				}

				monCfg, _ := cfg.(*global.MonsterCfg)
				rItems := p.generateFightWinReward(monCfg.DropData, randItem)
				if rItems != nil {
					for _, item := range rItems.Items {
						if _, ok := p.lastNormalFightReard.Items[item.Id]; ok {
							p.lastNormalFightReard.Items[item.Id].Num += item.Num
						} else {
							p.lastNormalFightReard.Items[item.Id] = &global.RewardItem{Id: item.Id, Num: item.Num}
						}
					}
				}
			}

			if fev.BBoss {
				global.ServerG.GetMapMgr().NextMap(p)
			}
		}
	} else if fev.Mode == global.FIGHT_MODE_INSTANCE { //副本战斗
		//设置下次可战斗时间
		p.nextInstanceFightTime = now + int64(p.GetVipEffectValue(global.Vip_Effect_FightWaitSec)) + int64(fev.FightRound)*global.FIGHT_ROUND_TIME_SEC
		p.lastInstanceFightReard.Items = nil
		if fev.Win {
			if fev.BBoss {
				//结算奖励
				p.lastInstanceFightReard.Items = make(map[int32]*global.RewardItem)
				icfg := global.ServerG.GetConfigMgr().GetCfg("Instance",  p.GetInstanceMapId())
				if icfg != nil {
					cfg := icfg.(*global.InstanceCfg)
					rItems := p.generateFightWinReward(cfg.DropData, nil)
					if rItems != nil {
						for _, item := range rItems.Items {
							if _, ok := p.lastInstanceFightReard.Items[item.Id]; ok {
								p.lastInstanceFightReard.Items[item.Id].Num += item.Num
							} else {
								p.lastInstanceFightReard.Items[item.Id] = &global.RewardItem{Id: item.Id, Num: item.Num}
							}
						}
					}
				} else {
					log.Error("Player::OnFightEvent get instance:%d nil", p.GetInstanceMapId())
				}

				p.nextInstanceFightTime = int64(0)
				p.SetInstanceMapId(0)
			} else {
				p.SetInstanceFightIndex(p.GetInstanceFightIndex() + 1)
			}
		} else { //挑战副本失败了
			p.nextInstanceFightTime = int64(0)
			p.SetInstanceMapId(0)
		}
	} else if fev.Mode == global.FIGHT_MODE_CHALLENGE { //挖矿抢劫
		p.curRobReward.Items = nil
		if fev.Win {
			level_id,err := p.GetPlayerChallengeInfo(p.curChallengeId)
			if err != nil {
				log.Error("player_fight_event GetPlayerChallengeInfo error:%s", err)
				p.curChallengeId = 0
				return
			}

			ccfg := p.GetChallengeMonCfg(p.curChallengeId, level_id + 1)
			if ccfg == nil {
				log.Error("player_fight_event GetChallengeMonCfg nil cid:%d lid:%d", p.curChallengeId, level_id)
				p.curChallengeId = 0
				return
			}

			randSrc := rand.NewSource(time.Now().UnixNano())
			randItem := rand.New(randSrc)
			p.lastChallengeFightReard.Items = make(map[int32]*global.RewardItem)
			for i:=0;i < Challenge_Mon_Num;i++ {
				rItems := p.generateFightWinReward(ccfg.DropData, randItem)
				if rItems != nil {
					for _, item := range rItems.Items {
						if _, ok := p.lastChallengeFightReard.Items[item.Id]; ok {
							p.lastChallengeFightReard.Items[item.Id].Num += item.Num
						} else {
							p.lastChallengeFightReard.Items[item.Id] = &global.RewardItem{Id: item.Id, Num: item.Num}
						}
					}
				}
			}
		} else {
			p.curChallengeId = 0
		}
	} else if fev.Mode == global.FIGHT_MODE_MINE_ROB { //挖矿抢劫
		p.curRobReward.Items = nil

		var dstDBId int64 = 0
		if p1, ok := fev.Defencers[0].(global.Player); ok {
			dstDBId = p1.GetDBId()
		} else if p2, ok := fev.Defencers[0].(global.OffLinePlayer); ok {
			dstDBId = p2.GetDBId()
		} else {
			return
		}

		cfg := global.ServerG.GetConfigMgr().GetCfg("Mine", p.curRobCfgId)
		if cfg == nil {
			return
		}

		icfg := cfg.(*global.MineCfg)
		pMineInfo, err := p.playerInfo.GetPlayerMineInfo(dstDBId)
		if err != nil {
			log.Error("player::fightOver playerId:%d  dstDBID:%d GetPlayerMineInfo error:%s", p.dbId, dstDBId, err)
			return
		}

		var pMine *global.PlayerMine = nil
		for _, mine := range pMineInfo.Mines {
			if mine.CfgId ==  p.curRobCfgId {
				pMine = mine
				break
			}
		}

		if pMine == nil {
			return
		}

		now := int32(time.Now().Unix())
		robNum := int32(-1)
		if fev.Win {
			calcSec := now - pMine.LastCalcTime
			if calcSec > icfg.FullSec {
				calcSec = icfg.FullSec
			}

			totalItemNum := calcSec * icfg.PerSecNum
			totalItemNum -= pMine.RobNum
			if totalItemNum <= 0 {
				return
			}

			robNum = int32(int(totalItemNum)* icfg.RobPer / 1000)
			pMine.RobNum += robNum
		}

		//消息
		robMsg := &global.PlayerMineMsg{
			Type:global.Mine_Msg_Type_Rob,
			Num:robNum,
			PlayerId:p.dbId,
			Time:now,
		}
		pMine.Msgs = append(pMine.Msgs, robMsg)

		if len(pMine.Msgs) > global.Mine_Msg_Max_Len {
			pMine.Msgs = pMine.Msgs[len(pMine.Msgs) - global.Mine_Msg_Max_Len:]
		}

		if err := p.playerInfo.SetPlayerMineInfo(dstDBId, pMineInfo); err != nil {
			log.Error("player::fightOver playerId:%d dstDBID:%d SetPlayerMineInfo error:%s", p.dbId, dstDBId, err)
			return
		}

		p.curRobReward.Items = make(map[int32]*global.RewardItem)
		p.curRobReward.Items[icfg.ProductItemId] = &global.RewardItem{
			Id:icfg.ProductItemId,
			Num:robNum,
		}

	} else if fev.Mode == global.FIGHT_MODE_ADVANCE { //
		if !fev.Win {
			return
		}

		advanceLv, _ := p.GetProp(global.Player_Prop_Advance_Level)
		icfg := global.ServerG.GetConfigMgr().GetCfg("Advance", advanceLv)
		if icfg == nil {
			return
		}

		cfg := icfg.(*global.AdvanceCfg)
		items := &global.RewardData{
			Items:make(map[int32]*global.RewardItem),
		}

		if cfg.ItemId1 != 0 {
			items.Items[cfg.ItemId1] = &global.RewardItem{
				Id: cfg.ItemId1,
				Num: -cfg.ItemNum1,
			}
		}

		if cfg.ItemId2 != 0 {
			items.Items[cfg.ItemId2] = &global.RewardItem{
				Id: cfg.ItemId2,
				Num: -cfg.ItemNum2,
			}
		}

		if cfg.ItemId3 != 0 {
			items.Items[cfg.ItemId3] = &global.RewardItem{
				Id: cfg.ItemId3,
				Num: -cfg.ItemNum3,
			}
		}

		lv1 := int(advanceLv / 10)
		newLv := int32((lv1 + 1) * 10 + 1)
		icfg = global.ServerG.GetConfigMgr().GetCfg("Advance", newLv)
		if icfg == nil {
			return
		}

		p.AddItems(items, true, true)
		p.SetProp(global.Player_Prop_Advance_Level, newLv, false)
	}
}

func (p *player) checkCanFight(mode uint32) bool {
	m := &msg.GSCL_PlayerFightNeedTime{
		Mode: int32(mode),
		Time: 0,
	}

	//防止背包还未拉取到
	if p.BackPackDBData == nil {
		m.Time = 1
		p.conn.Send(m)
		return false
	}

	now := global.ServerG.GetCurTime()
	nextTime := p.nextNormalFightTime
	if mode == global.FIGHT_MODE_INSTANCE { //普通战斗
		nextTime = p.nextInstanceFightTime
	}

	needTime := int32((nextTime-now)/int64(time.Second)) + int32(1)
	if needTime > 0 {
		m.Time = needTime
		p.conn.Send(m)
		return false
	}

	return true
}
