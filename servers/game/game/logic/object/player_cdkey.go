package object


import (
	"xianxia/servers/game/game/global"
	"strings"
	"strconv"
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/msg"
	"github.com/garyburd/redigo/redis"
	"fmt"
	"xianxia/common/dbengine"
	"encoding/json"
)

func (p *player) GetPlayerCDKeys() {
	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_GET_CDKEY, int64(0),"GET", fmt.Sprintf("playercdkeys:%d", p.dbId))
}

func (p *player) ReadCDKeyFromDB(ret *dbengine.CDBRet) (err error) {
	if ret.Err != nil {
		// TODO:打日志
		log.Error("player:%d ReadCDKeyFromDB error:%p", p.GetDBId(), ret.Err)
		return nil
	}

	values, err := redis.String(ret.Content, nil)
	if err == redis.ErrNil {
		p.initCDKeyData()
		return nil
	}

	p.initCDKeyData()
	err = json.Unmarshal([]byte(values), p.Player_CDKesy_Info)
	if err != nil {
		log.Error("player:%d ReadCDKeyFromDB error:%p", p.GetDBId(), err)
		return nil
	}

	return nil
}

func (p *player) SaveCDKey() {
	j, err := json.Marshal(p.Player_CDKesy_Info)
	if err != nil {
		log.Error("player:%d SaveCDKey Marshal Error:%p", p.GetDBId(), err)
		return
	}

	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_CDKEY, int64(0),"SET", fmt.Sprintf("playercdkeys:%d", p.dbId), j)
}

func (p *player) initCDKeyData() {
	p.Player_CDKesy_Info = &global.Player_CDKesy_Info{
		UseCDKeyArr:[]string{},
	}
}

//cdkey领奖
func (p *player) cdKeyReward(recvData []byte) {
	if recvData == nil || len(recvData) == 0 {
		return
	}

	//检查背包是否满了
	errMsg := ""
	var reward *global.RewardData
	autoSell := false

	cdkey := string(recvData)
	data, err := p.playerInfo.GetCDKey(cdkey)
	if err != nil {
		errMsg = "兑换码已过期"
	} else {
		if len(data.Content) == 0 {
			errMsg = "兑换码已过期"
			log.Error("兑换码：%s Content format error1", cdkey)
		} else {
			itemArr := strings.Split(data.Content, ";")
			if len(itemArr) == 0 {
				errMsg = "兑换码已过期"
				log.Error("兑换码：%s Content format error2", string(recvData))
			} else {
				reward = &global.RewardData{
					Items: make(map[int32]*global.RewardItem),
				}

				for _, itemStr := range itemArr {
					itemCountArr := strings.Split(itemStr, "+")
					if len(itemCountArr) != 2 {
						errMsg = "兑换码已过期"
						log.Error("兑换码：%s Content format error3", string(recvData))
						reward = nil
						break
					}

					itemId, err := strconv.Atoi(itemCountArr[0])
					if err != nil {
						errMsg = "兑换码已过期"
						log.Error("兑换码：%s Content format error3", string(recvData))
						reward = nil
						break
					}

					itemCount, err := strconv.Atoi(itemCountArr[1])
					if err != nil || itemCount <= 0 {
						errMsg = "兑换码已过期"
						log.Error("兑换码：%s Content format error4", string(recvData))
						reward = nil
						break
					}

					reward.Items[int32(itemId)] = &global.RewardItem{
						Id:  int32(itemId),
						Num: int32(itemCount),
					}
				}

				//检查背包是否能装下
				if p.IsBagFullMulti(reward, autoSell) {
					errMsg = "背包空间不足!!!"
					reward = nil
				}
			}
		}
	}

	if errMsg == "" && data.Type == global.CDKEY_TYPE_ALL {
		//检查一下使用过的cdkey
		if p.Player_CDKesy_Info == nil {
			log.Error("player:%d Player_CDKesy_Info nil", p.dbId)
			return
		}

		for _, u_cdkey := range p.Player_CDKesy_Info.UseCDKeyArr {
			if u_cdkey == cdkey {
				errMsg = "不能重复领取该兑换码"
				break
			}
		}
	}

	if errMsg != "" {
		m := &msg.GSCL_Error{
			Desc: []byte(errMsg),
		}
		p.conn.Send(m)
		return
	}

	if data.Type == global.CDKEY_TYPE_ONE {
		p.playerInfo.RemoveCDKey(cdkey)
	} else if data.Type == global.CDKEY_TYPE_ALL {
		p.Player_CDKesy_Info.UseCDKeyArr = append(p.Player_CDKesy_Info.UseCDKeyArr, cdkey)
	}

	_, _, sellItems, err := p.AddItems(reward, true, autoSell)
	if err != nil {
		log.Error("兑换码：addItems error")
		return
	}

	mItems := make(map[int32]int32)
	for _, sItem := range reward.Items {
		mItems[sItem.Id] = sItem.Num
	}

	mSellItems := make(map[int32]int32)
	for _, sItem := range sellItems {
		mSellItems[sItem.CfgId] = sItem.Num
	}

	msuc := &msg.GSCL_CDKeyReward{
		Items:     mItems,
		SellItems: mSellItems,
	}

	p.conn.Send(msuc)
}
