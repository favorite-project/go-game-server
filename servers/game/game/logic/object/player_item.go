package object

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"xianxia/common/dbengine"
	"xianxia/servers/game/conf"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/msg"

	"github.com/garyburd/redigo/redis"
	"github.com/name5566/leaf/log"
)

func getRewardFromStr(str string) *global.RewardData {
	if len(str) == 0 {
		return nil
	}

	items := &global.RewardData{
		Items:make(map[int32]*global.RewardItem),
	}
	arr1 := strings.Split(str, ";")
	for _, item := range arr1 {
		arr2 := strings.Split(item, "+")
		if len(arr2) != 2 {
			continue
		}

		id, err := strconv.Atoi(arr2[0])
		if err != nil {
			continue
		}

		num, err := strconv.Atoi(arr2[1])
		if err != nil {
			continue
		}

		items.Items[int32(id)] = &global.RewardItem{
			Id:int32(id),
			Num:int32(num),
		}
	}

	return items
}

func (p *player) SaveBackPack() {

	j, err := json.Marshal(p.BackPackDBData)
	if err != nil {
		log.Error("player:%d SaveBackPack Marshal Error:%p", p.GetDBId(), err)
		return
	}

	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_BACKPACK, int64(0),"SET", fmt.Sprintf("backPack:%d", p.dbId), j)
}

/**********************用户背包和装备****************************/

func (p *player) GetPlayerBackPack() {
	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_GET_BACKPACK, int64(0),"GET", fmt.Sprintf("backPack:%d", p.dbId))
}

//首次没有背包数据
func (p *player) initBackPackData() {
	p.BackPackDBData = &global.BackPackDBData{}
	p.BackPackDBData.Num = global.BACKPACK_INIT_NUM
	p.BackPackDBData.BagData = make(map[int16][]*global.ItemDBData)
	for i := global.BACKPACK_ITEMS_BAG; i <= global.BACKPACK_SKILL_BAG; i++ {
		p.BackPackDBData.BagData[i] = make([]*global.ItemDBData, 0)
	}
}

func (p *player) ReadBackFromDB(ret *dbengine.CDBRet) (err error) {
	if ret.Err != nil {
		// TODO:打日志
		log.Error("player:%d ReadBackFromDB error:%p", p.GetDBId(), ret.Err)
		return nil
	}

	values, err := redis.String(ret.Content, nil)
	if err == redis.ErrNil {
		p.initBackPackData()
		return nil
	}

	p.initBackPackData()
	err = json.Unmarshal([]byte(values), p.BackPackDBData)
	if err != nil {
		log.Error("player:%d ReadBackFromDB error:%p", p.GetDBId(), err)
		return nil
	}

	return nil
}

func (p *player) SendBackPackToClient() {
	m := &msg.GSCL_CreateBackPack{BackPackDBData: p.BackPackDBData}

	p.conn.Send(m)
}

func (p *player) sendEquipAutoSellToClient() {
	m := &msg.GSCL_PlayerEquipAutoSell{EquipAutoSell: p.BackPackDBData.EquipSellSettings}

	p.conn.Send(m)
}

func (p *player) equipAutoSell(recvData []byte) {
	if len(recvData) < 2 {
		return
	}

	qlen := conf.RdWrEndian.Uint16(recvData)
	if len(recvData) != 2+int(qlen*2) {
		return
	}

	p.BackPackDBData.EquipSellSettings = []int16{}
	for i := 0; i < int(qlen); i++ {
		quality := int16(conf.RdWrEndian.Uint16(recvData[2+i*2:]))
		p.BackPackDBData.EquipSellSettings = append(p.BackPackDBData.EquipSellSettings, quality)
	}
}

/*======================通用物品操作============================*/
func (p *player) arrangeBag(bid int16) []int32 {
	Indexes := make([]int, 0)

	delInstIds := []int32{}

	for i, item := range p.BackPackDBData.BagData[bid] {
		if item.Num <= 0 {
			Indexes = append(Indexes, i)
			delInstIds = append(delInstIds, item.Id)
		}
	}

	if len(Indexes) > 0 {
		newBagItems := make([]*global.ItemDBData, 0)
		for i := 0; i < len(Indexes); i++ {
			index := Indexes[i]

			var beginIndex int = 0
			if i > 0 {
				beginIndex = Indexes[i-1] + 1
			}

			newBagItems = append(newBagItems, p.BackPackDBData.BagData[bid][beginIndex:index]...)
		}

		if Indexes[len(Indexes)-1]+1 < len(p.BackPackDBData.BagData[bid]) {
			newBagItems = append(newBagItems, p.BackPackDBData.BagData[bid][Indexes[len(Indexes)-1]+1:]...)
		}

		p.BackPackDBData.BagData[bid] = newBagItems
	}

	return delInstIds
}

func (p *player) AddItem(cfgId int32, num int32, sendMsg bool, autoSell bool) ([]*global.ItemDBData, []int32, []*global.SellItemCfgInfo, error) {
	if num == 0 {
		return nil, nil, nil, errors.New("num == 0")
	}
	cfg := p.getItemCfgByCfgId(cfgId)
	if cfg == nil {
		log.Error("player:%d AddItem  get cfg empty", p.GetDBId())
		return nil, nil, nil, errors.New("get cfg empty")
	}

	//金币，经验，钻石直接加到属性上
	if cfg.GetType() == global.ITEM_TYPE_GOLD {
		p.SetProp(global.Player_Prop_Money, num, true)
		return nil, nil, nil, nil
	}

	if cfg.GetType() == global.ITEM_TYPE_EXP {
		p.SetProp(global.Player_Prop_Exp, num, true)
		return nil, nil, nil, nil
	}

	if cfg.GetType() == global.ITEM_TYPE_DIAMOND {
		p.SetProp(global.Player_Prop_Diamond, num, true)
		return nil, nil, nil, nil
	}

	if cfg.GetType() == global.ITEM_TYPE_ADVANCE {
		p.SetProp(global.Player_Prop_Advance_Exp, num, true)
		return nil, nil, nil, nil
	}

	bid := p.getBagTypeByCfgId(cfg)
	if num < 0 {
		baseNum := num
		Indexes := make([]int, 0)

		for i, item := range p.BackPackDBData.BagData[bid] {
			if item.CfgId == cfgId {
				baseNum += item.Num
				Indexes = append(Indexes, i)
				if baseNum >= 0 {
					break
				}
			}
		}

		if baseNum < 0 {
			return nil, nil, nil, errors.New("num not enougth")
		}

		baseNum = num
		addItems := []*global.ItemDBData{}
		for _, index := range Indexes {
			iNum := p.BackPackDBData.BagData[bid][index].Num
			p.BackPackDBData.BagData[bid][index].Num += baseNum
			baseNum += iNum

			if p.BackPackDBData.BagData[bid][index].Num > 0 {
				addItems = append(addItems, p.BackPackDBData.BagData[bid][index])
			}
		}

		DelInstIds := p.arrangeBag(bid)
		if sendMsg && (len(addItems) > 0 || len(DelInstIds) > 0) {
			m := &msg.GSCL_PlayerUpdateBackPack{
				AddItems:addItems,
				DelItems:DelInstIds,
			}
			p.conn.Send(m)
		}

		return addItems, DelInstIds, nil, nil
	}

	var sellInfo []*global.SellItemCfgInfo = nil
	if p.BackPackDBData.Num <= int32(len(p.BackPackDBData.BagData[bid])) {
		if autoSell {
			sellInfo = p.sellItemByCfgId(cfgId, num)
			return nil, nil, sellInfo, nil
		}

		return nil, nil, nil, errors.New("Bag Full")
	}

	//设置自动卖出
	if cfg.GetType() == global.ITEM_TYPE_EQUIPMENT {
		ecfg := cfg.(*global.EquipmentCfg)
		for _, quality := range p.BackPackDBData.EquipSellSettings {
			if quality == ecfg.Quality {
				sellInfo = p.sellItemByCfgId(cfgId, num)
				return nil, nil, sellInfo, nil
			}
		}
	}

	addItems := make([]*global.ItemDBData, 0)
	baseNum := num
	limitNum := cfg.GetMaxNum()
	for _, item := range p.BackPackDBData.BagData[bid] {
		if item.CfgId == cfgId {
			if item.Num >= limitNum {
				continue
			}

			addItems = append(addItems, item)
			bNum := item.Num
			item.Num += baseNum

			if item.Num > limitNum {
				item.Num = limitNum
			}

			baseNum -= (item.Num - bNum)
			if baseNum <= 0 {
				break
			}
		}
	}

	//最后剩下的格子可以 装得下
	if baseNum > 0 {
		for {
			uniqueId, err := global.ServerG.GetDBEngine().GetUniqueID()
			if err != nil {
				log.Error("player:%d AddItem  GetUniqueID  Error:", p.GetDBId(), err)
				return addItems, nil, nil, nil
			}

			addNum := baseNum
			if addNum > limitNum {
				addNum = limitNum
			}

			item := &global.ItemDBData{
				Id:      int32(uniqueId),
				CfgId:   cfgId,
				Num:     addNum,
				Data:    "",
				Binding: byte(0),
			}

			if cfg.GetType() == global.ITEM_TYPE_EQUIPMENT {
				ecfg, ok := cfg.(*global.EquipmentCfg)
				if !ok {
					log.Error("player:%d AddItem cfg.(*global.EquipmentCfg) error")
					return addItems, nil, nil, nil
				}

				//生成装备属性
				addProperty := &global.EquipDBItemData{
					BData:    make(map[int32]int32),
					OData:    make(map[int32]int32),
					UpdateLv: 0,
					UseCnt:   0,
				}

				randProps := global.ServerG.GetRandSrc()
				//基础属性生成
				propStrArr := strings.Split(ecfg.Data, ";")
				for _, propItemStr := range propStrArr {
					propkvStrArr := strings.Split(propItemStr, "+")
					if len(propkvStrArr) < 2 {
						continue
					}
					// pk:基础属性ID
					pk, err := strconv.Atoi(propkvStrArr[0])
					if err != nil {
						log.Error("player:%d AddItem  strconv.Atoi(propkvStrArr[0])  format error:", p.GetDBId())
						return addItems, nil, nil, nil
					}

					// pvMin:基础属性最小值
					pvMin, err := strconv.Atoi(propkvStrArr[1])
					if err != nil {
						log.Error("player:%d AddItem  strconv.Atoi(propkvStrArr[1])  format error:", p.GetDBId())
						return addItems, nil, nil, nil
					}

					pv := int32(pvMin)
					if len(propkvStrArr) > 2 {
						// 如果有第三个值,基础属性就是去
						pvMax, err := strconv.Atoi(propkvStrArr[2])
						if err != nil {
							log.Error("player:%d AddItem  strconv.Atoi(propkvStrArr[2])  format error:", p.GetDBId())
							return addItems, nil, nil, nil
						}

						if pvMax > pvMin {
							pv = int32(pvMin + randProps.Intn(pvMax-pvMin))
						}
					}

					addProperty.BData[int32(pk)] = pv
				}

				//生成附加属性
				if ecfg.OtherData >= 1.0 {
					propNum := p.GetEquipTwoPropsCnt(ecfg.Quality)
					epCsv := global.ServerG.GetConfigMgr().GetCsv("EquipmentProps")
					if epCsv != nil {
						randEPCfgs := []*global.EquipmentPropsCfg{}
						for i := 0; i < epCsv.NumRecord(); i++ {
							epicfg := epCsv.Record(i)
							epCfg, ok := epicfg.(*global.EquipmentPropsCfg)
							if !ok {
								continue
							}

							if epCfg.Level != ecfg.Level {
								continue
							}

							if epCfg.EType != 0 && epCfg.EType != ecfg.SubType {
								continue
							}

							randEPCfgs = append(randEPCfgs, epCfg)
						}

						for i := 0; i < propNum; i++ {
							epLen := len(randEPCfgs)
							if epLen == 0 {
								break
							}

							n := randProps.Intn(epLen)
							pv := randEPCfgs[n].MinValue
							duration := int(randEPCfgs[n].MaxValue - randEPCfgs[n].MinValue)
							if duration > 0 {
								pv += int32(randProps.Intn(duration))
							}

							addProperty.OData[randEPCfgs[n].PropId] = int32(math.Ceil(ecfg.OtherData * float64(pv)))

							randEPCfgs = append(randEPCfgs[:n], randEPCfgs[n+1:]...)
						}
					}
				}

				bData, err := json.Marshal(addProperty)
				if err != nil {
					log.Error("player:%d AddItem Marshal Error:%p", p.GetDBId(), err)
					return addItems, nil, nil, nil
				}
				item.Data = string(bData)
			}

			p.BackPackDBData.BagData[bid] = append(p.BackPackDBData.BagData[bid], item)
			addItems = append(addItems, item)

			//检查是否添加完了
			baseNum -= addNum
			if baseNum <= 0 {
				break
			}

			//检查背包是否满了
			if p.BackPackDBData.Num <= int32(len(p.BackPackDBData.BagData[bid])) {
				if autoSell {
					sellInfo = p.sellItemByCfgId(cfgId, baseNum)
				}

				break
			}
		}
	}

	if sendMsg && len(addItems) > 0 {
		m := &msg.GSCL_PlayerUpdateBackPack{
			AddItems: addItems,
		}

		p.conn.Send(m)
	}

	return addItems, nil, sellInfo, nil
}

//根據instid扣除道具
func (p *player) SubItemByInstId(instId int32, num int32, sendMsg bool) ([]*global.ItemDBData, []int32, error) {
	if num <= 0 {
		return nil, nil, errors.New("SubItemByInstId num error")
	}

	itemData := p.getItemDataInBagById(instId)
	if itemData == nil {
		return nil, nil, errors.New("SubItemByInstId instData nil")
	}

	cfg := p.getItemCfgByCfgId(itemData.CfgId)
	if cfg == nil {
		return nil, nil, errors.New("SubItemByInstId cfg nil")
	}

	itemData.Num -= num

	var delInstIds  []int32
	var updateItems []*global.ItemDBData
	if itemData.Num == 0 {
		delInstIds = p.arrangeBag(p.getBagTypeByCfgId(cfg))
	} else {
		updateItems = []*global.ItemDBData{itemData}
	}

	if sendMsg {
		m := &msg.GSCL_PlayerUpdateBackPack{
			DelItems:delInstIds,
			AddItems:updateItems,
		}

		p.conn.Send(m)
	}

	return updateItems, delInstIds, nil
}

// 根据品质
func (p *player) GetEquipTwoPropsCnt(quality int16) int {
	rand.Seed(time.Now().Unix())
	r := []int16{quality - 1, quality}
	ret := r[rand.Intn(len(r))]
	return int(ret)
}

func (p *player) AddItems(data *global.RewardData, sendMsg bool, autoSell bool) ([]*global.ItemDBData, []int32, []*global.SellItemCfgInfo, error) {

	if data == nil || data.Items == nil {
		return nil, nil, nil, errors.New("addItems rewardData nil")
	}

	addItems := make([]*global.ItemDBData, 0)
	delInstIds := []int32{}
	sellInfo := []*global.SellItemCfgInfo{}
	for _, rItem := range data.Items {
		//道具往背包加
		if items, ids, sells, err := p.AddItem(rItem.Id, rItem.Num, sendMsg, autoSell); err == nil {
			if items != nil {
				addItems = append(addItems, items...)
			}

			if ids != nil {
				delInstIds = append(delInstIds, ids...)
			}

			if sells != nil {
				sellInfo = append(sellInfo, sells...)
			}
		} else {
			break
		}
	}

	if sendMsg && (len(addItems) > 0 || len(delInstIds) > 0) {
		m := &msg.GSCL_PlayerUpdateBackPack{
			AddItems: addItems,
			DelItems: delInstIds,
		}

		p.conn.Send(m)
	}

	return addItems, delInstIds, sellInfo, nil
}

func (p *player) getBagTypeByCfgId(cfg global.ItemCfgInterface) int16 {
	if cfg == nil {
		return global.BACKPACK_ITEMS_BAG
	}

	bid := global.BACKPACK_ITEMS_BAG
	switch cfg.GetType() {
	case global.ITEM_TYPE_EQUIPMENT:
		bid = global.BACKPACK_EQUIP_BAG
	case global.ITEM_TYPE_STONE:
		bid = global.BACKPACK_STONE_BAG
	case global.ITEM_TYPE_SKILL:
		bid = global.BACKPACK_SKILL_BAG
	}

	return bid
}

func (p *player) getItemDataInBagById(instId int32) *global.ItemDBData {
	for _, bagData := range p.BackPackDBData.BagData {
		for _, itemData := range bagData {
			if itemData.Id == instId {
				return itemData
			}
		}
	}

	return nil
}

func (p *player) getItemCfgByCfgId(cfgId int32) global.ItemCfgInterface {
	bcfg := global.ServerG.GetConfigMgr().GetCfg("Equipment", cfgId)
	if bcfg == nil {
		bcfg = global.ServerG.GetConfigMgr().GetCfg("Item", cfgId)
		if bcfg == nil {

			return nil
		}
	}

	return bcfg.(global.ItemCfgInterface)
}

func (p *player) UseItem(instId int32, num int32, useType uint16) bool {
	switch useType {
	case global.ITEM_USE_TYPE_BUY:
	case global.ITEM_USE_TYPE_SELL:
		p.sellItemByInstId(instId, num)
	case global.ITEM_USE_TYPE_NORMAL:
	case global.ITEM_USE_TYPE_OPEN:
		p.useItemByInstId(instId, num)
	}

	return true
}
func (p *player) useItemByInstId(instId int32, num int32)  {
	itemData := p.getItemDataInBagById(instId)
	if itemData == nil || num <= 0 {
		log.Error("player:%d sellItemByInstId  get itemData empty", p.GetDBId())
		return
	}

	cfg := global.ServerG.GetConfigMgr().GetCfg("Item", itemData.CfgId)
	if cfg == nil {
		log.Error("player:%d sellItemByInstId  get cfg empty", p.GetDBId())
		return
	}

	itemCfg := cfg.(*global.ItemCfg)
	switch itemCfg.Type {
	case global.ITEM_TYPE_GIFT:
		content := strings.Split(itemCfg.Data,";")
		rewardData := &global.RewardData{
			Items:make(map[int32]*global.RewardItem),
		}

		for _,v := range content {
			itemInfo := strings.Split(v,"+")
			item_id,err := strconv.Atoi(itemInfo[0])
			if err != nil {
				log.Error("@player:%d use item fail(convert fail)", p.GetDBId())
				return
			}

			item_val,err := strconv.Atoi(itemInfo[1])
			if err != nil {
				log.Error("#player:%d use item fail(convert fail)", p.GetDBId())
				return
			}
			if _, ok := rewardData.Items[int32(item_id)]; ok {
				rewardData.Items[int32(item_id)].Num += int32(item_val)
			} else {
				rewardData.Items[int32(item_id)] = &global.RewardItem{
					Id:int32(item_id),
					Num:int32(item_val),
				}
			}
		}

		for _, item := range rewardData.Items {
			item.Num *= num
		}

		autoSell := true
		if p.IsBagFullMulti(rewardData, autoSell) {
			m := &msg.GSCL_Error{
				Desc:[]byte("背包已满!!!"),
			}

			p.conn.Send(m)
			return
		}

		p.AddItem(itemCfg.Id, -num, true, true)
		p.AddItems(rewardData, true, autoSell)
	case global.ITEM_TYPE_VIP:
		vipValue, err := strconv.Atoi(itemCfg.Data)
		if err != nil {
			log.Error("Player::useItemByInstId cfgId:%d strconv.Atoi data error:%s", itemCfg.Id, err)
			return
		}

		p.AddItem(itemCfg.Id, -num, true, true)
		p.SetProp(global.Player_Prop_VipExp, int32(vipValue) * num, true)
	}
}

func (p *player) sellItemByInstId(instId int32, num int32) {
	itemData := p.getItemDataInBagById(instId)
	if itemData == nil {
		log.Error("player:%d sellItemByInstId  get itemData empty", p.GetDBId())
		return
	}

	cfg := p.getItemCfgByCfgId(itemData.CfgId)
	if cfg == nil {
		log.Error("player:%d sellItemByInstId  get cfg empty", p.GetDBId())
		return
	}

	if num <= 0 || cfg.GetSellPrice() <= 0 {
		return
	}

	if itemData.Num < num {
		return
	}

	itemData.Num -= num
	m := &msg.GSCL_PlayerUpdateBackPack{}

	if itemData.Num == 0 {
		m.DelItems = p.arrangeBag(p.getBagTypeByCfgId(cfg))
	} else {
		m.AddItems = []*global.ItemDBData{itemData}
	}

	p.conn.Send(m)

	addMoney := cfg.GetSellPrice() * num
	p.SetProp(global.Player_Prop_Money, addMoney, true)
}

func (p *player) sellItemByCfgId(cfgId int32, num int32) []*global.SellItemCfgInfo {
	cfg := p.getItemCfgByCfgId(cfgId)
	if cfg == nil || num <= 0 {
		log.Error("player:%d sellItemByCfgId  get cfg empty", p.GetDBId())
		return nil
	}

	sellInfo := []*global.SellItemCfgInfo{}
	sellItem := &global.SellItemCfgInfo{
		CfgId: cfgId,
		Num:   num,
	}

	sellInfo = append(sellInfo, sellItem)

	addMoney := cfg.GetSellPrice() * num
	p.SetProp(global.Player_Prop_Money, addMoney, true)

	return sellInfo
}

func (p *player) GetBagItemNum(cfgId int32) int32 {
	cfg := p.getItemCfgByCfgId(cfgId)
	if cfg == nil {
		return 0
	}

	//金币，经验，钻石直接加到属性上
	if cfg.GetType() == global.ITEM_TYPE_GOLD {
		num, _ := p.GetProp(global.Player_Prop_Money)
		return num
	}

	if cfg.GetType() == global.ITEM_TYPE_EXP {
		num, _ := p.GetProp(global.Player_Prop_Exp)
		return num
	}

	if cfg.GetType() == global.ITEM_TYPE_DIAMOND {
		num, _ := p.GetProp(global.Player_Prop_Diamond)
		return num
	}

	if cfg.GetType() == global.ITEM_TYPE_ADVANCE {
		num, _ := p.GetProp(global.Player_Prop_Advance_Exp)
		return num
	}

	num := int32(0)
	for _, itemData := range p.BackPackDBData.BagData[p.getBagTypeByCfgId(cfg)] {
		if itemData.CfgId == cfgId {
			num += itemData.Num
		}
	}

	return num
}

func (p *player) RegisterSendItem() {
	// 赠送改名卡
	fmt.Println("赠送改名卡")
	p.AddItem(60003, 1, false, false)
	fmt.Println("赠送礼包")
	//p.AddItem(60010, 1, false, false)
}

func (p *player) expandBag(recvData []byte) {
	//是否已经最大
	if p.BackPackDBData.Num >= global.BACKPACK_INIT_NUM + global.BackPack_Expand_Base_Count * global.BackPack_Expand_Max_Count {
		return
	}

	//检查钻石
	index := int32((p.BackPackDBData.Num - global.BACKPACK_INIT_NUM )/ global.BackPack_Expand_Base_Count)
	needDiamond := int32((index + 1) * global.BackPack_Expand_Base_Diamond)
	playeDiamond, _ := p.GetProp(global.Player_Prop_Diamond)
	if playeDiamond < needDiamond {
		return
	}

	p.SetProp(global.Player_Prop_Diamond, -needDiamond, true)

	p.BackPackDBData.Num += global.BackPack_Expand_Base_Count
}

func (p *player) IsBagFull(cfgId int32, num int32, autoSell bool) bool {
	if num <= 0 {
		return true
	}

	cfg := p.getItemCfgByCfgId(cfgId)
	if cfg == nil {
		log.Error("player:%d IsBagFull  get cfg empty", p.GetDBId())
		return true
	}

	//金币，经验，钻石直接加到属性上
	cfgType := cfg.GetType()
	if cfgType == global.ITEM_TYPE_GOLD  ||
		cfgType == global.ITEM_TYPE_EXP  ||
		cfgType == global.ITEM_TYPE_DIAMOND ||
		cfgType == global.ITEM_TYPE_ADVANCE {
		return false
	}

	bid := p.getBagTypeByCfgId(cfg)
	curBagNum :=int32(len(p.BackPackDBData.BagData[bid]))
	if p.BackPackDBData.Num <= curBagNum {
		if autoSell {
			return false
		}

		return true
	}

	//设置自动卖出
	if cfg.GetType() == global.ITEM_TYPE_EQUIPMENT {
		ecfg := cfg.(*global.EquipmentCfg)
		for _, quality := range p.BackPackDBData.EquipSellSettings {
			if quality == ecfg.Quality {
				return false
			}
		}
	}

	baseNum := num
	limitNum := cfg.GetMaxNum()
	for _, item := range p.BackPackDBData.BagData[bid] {
		if item.CfgId == cfgId {
			if item.Num >= limitNum {
				continue
			}

			addNum := limitNum - item.Num
			baseNum -= addNum
			if baseNum <= 0 {
				break
			}
		}
	}

	//最后剩下的格子可以 装得下
	if baseNum > 0 {
		for {
			addNum := baseNum
			if addNum > limitNum {
				addNum = limitNum
			}

			curBagNum++

			//检查是否添加完了
			baseNum -= addNum
			if baseNum <= 0 {
				break
			}

			//检查背包是否满了
			if p.BackPackDBData.Num <= curBagNum {
				if autoSell {
					return false
				}

				return true
			}
		}
	}

	return false
}

func (p *player) IsBagFullMulti(items *global.RewardData, autoSell bool) bool {
	if items == nil || items.Items == nil {
		return true
	}

	tmpBagMums := make(map[int16]int32)
	for _, item := range items.Items {
		cfgId := item.Id
		num := item.Num
		if num <= 0 {
			return true
		}

		cfg := p.getItemCfgByCfgId(cfgId)
		if cfg == nil {
			log.Error("player:%d IsBagFullMulti  get cfg empty,cfgId:%d", p.GetDBId(),cfgId)
			return true
		}

		//金币，经验，钻石直接加到属性上
		cfgType := cfg.GetType()
		if cfgType == global.ITEM_TYPE_GOLD ||
			cfgType == global.ITEM_TYPE_EXP ||
			cfgType == global.ITEM_TYPE_DIAMOND ||
			cfgType == global.ITEM_TYPE_ADVANCE {
			continue
		}

		bid := p.getBagTypeByCfgId(cfg)
		var curBagNum int32
		if _, ok := tmpBagMums[bid]; ok {
			curBagNum = tmpBagMums[bid]
		} else {
			curBagNum =int32(len(p.BackPackDBData.BagData[bid]))
			tmpBagMums[bid] = curBagNum
		}

		if p.BackPackDBData.Num <= curBagNum {
			if autoSell {
				continue
			}

			return true
		}

		//设置自动卖出
		if cfg.GetType() == global.ITEM_TYPE_EQUIPMENT {
			ecfg := cfg.(*global.EquipmentCfg)
			canSell := false
			for _, quality := range p.BackPackDBData.EquipSellSettings {
				if quality == ecfg.Quality {
					canSell = true
					break
				}
			}

			if canSell {
				continue
			}
		}

		baseNum := num
		limitNum := cfg.GetMaxNum()
		for _, item := range p.BackPackDBData.BagData[bid] {
			if item.CfgId == cfgId {
				if item.Num >= limitNum {
					continue
				}

				addNum := limitNum - item.Num
				baseNum -= addNum
				if baseNum <= 0 {
					break
				}
			}
		}

		//最后剩下的格子可以 装得下
		if baseNum > 0 {
			for {
				addNum := baseNum
				if addNum > limitNum {
					addNum = limitNum
				}

				curBagNum++
				tmpBagMums[bid] = curBagNum

				//检查是否添加完了
				baseNum -= addNum
				if baseNum <= 0 {
					break
				}

				//检查背包是否满了
				if p.BackPackDBData.Num <= curBagNum {
					if autoSell {
						break
					}

					return true
				}
			}
		}
	}

	return false
}