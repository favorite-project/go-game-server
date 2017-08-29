package object

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"xianxia/common/dbengine"
	"xianxia/servers/game/game/errorx"
	global "xianxia/servers/game/game/global"
	"xianxia/servers/game/msg"

	"github.com/garyburd/redigo/redis"
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/utils"
	"math"
)

func (p *player) GetPlayerEquipment() {
	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_GET_EQUIPMENT, int64(0),"GET", fmt.Sprintf("equip:%d", p.dbId))
}

func (p *player) ReadEquipFromDB(ret *dbengine.CDBRet) (err error) {
	if ret.Err != nil {
		// TODO:打日志
		log.Error("player:%d ReadEquipFromDB error:%p", p.GetDBId(), ret.Err)
		return nil
	}

	values, err := redis.String(ret.Content, nil)
	if err == redis.ErrNil {
		p.initEquipData()
		return nil
	}

	p.initEquipData()
	err = json.Unmarshal([]byte(values), p.EquipDBData)
	if err != nil {
		log.Error("player:%d ReadEquipFromDB error:%p", p.GetDBId(), err)
		return nil
	}

	// 解析装备,筛出穿在身上的装备
	for _, eItem := range p.EquipDBData.EquipData {
		p.updateEquipmentProps(eItem, false)
	}

	p.CaculateFightVal()

	return nil
}

func (p *player) initEquipData() {
	p.EquipDBData = &global.EquipDBData{}
	p.EquipDBData.EquipData = make(map[int16]*global.ItemDBData)
}

func (p *player) SaveEquip() {

	j, err := json.Marshal(p.EquipDBData)
	if err != nil {
		log.Error("player:%d SaveBackPack Marshal Error:%p", p.GetDBId(), err)
		return
	}

	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_EQUIPMENT, int64(0),"SET", fmt.Sprintf("equip:%d", p.dbId), j)
}

func (p *player) updateEquipmentProps(eItem *global.ItemDBData, unEquip bool) {
	if eItem == nil {
		return
	}

	propsMap := &global.EquipDBItemData{}
	err := json.Unmarshal([]byte(eItem.Data), propsMap)
	if err != nil {
		log.Error("player:%d updateEquipmentProps error:%p", p.GetDBId(), err)
		return
	}

	per := int32(1)
	if unEquip {
		per = -per
	}

	for pid, pv := range propsMap.BData {
		p.SetProp(int(pid), pv*per, true)
	}

	for pid, pv := range propsMap.OData {
		p.SetProp(int(pid), pv*per, true)
	}

}

func (p *player) Equip(Id int32, bEquip bool) bool {
	if bEquip {
		var eItem *global.ItemDBData
		//检查是否已经装备
		for _, eItem = range p.EquipDBData.EquipData {
			if eItem.Id == Id {
				return false
			}
		}

		eItem = nil
		var bagIndex int
		for bagIndex, eItem = range p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG] {
			if eItem.Id == Id {
				break
			}
		}

		//背包中未找到此装备
		if eItem == nil {
			return false
		}

		icfg := global.ServerG.GetConfigMgr().GetCfg("Equipment", eItem.CfgId)
		if icfg == nil {
			return false
		}

		cfg := icfg.(*global.EquipmentCfg)

		//检查等级
		pLevel, _ := p.GetProp(global.Player_Prop_Level)
		if pLevel < cfg.Level {
			return false
		}

		oldEItem, _ := p.EquipDBData.EquipData[cfg.SubType]
		p.EquipDBData.EquipData[cfg.SubType] = eItem
		message := &msg.GSCL_PlayerUpdateBackPack{
			AddItems: []*global.ItemDBData{},
			DelItems: []int32{},
		}

		message.DelItems = append(message.DelItems, eItem.Id)

		p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG] = append(p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG][:bagIndex], p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG][bagIndex+1:]...)
		if oldEItem != nil { //直接替换
			p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG] = append(p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG], oldEItem)
			message.AddItems = append(message.AddItems, oldEItem)

			//重新计算一下属性
			p.updateEquipmentProps(oldEItem, true)
		}

		//重新计算一下属性
		p.updateEquipmentProps(eItem, false)

		p.conn.Send(message)
	} else {
		//背包满了
		if int(p.BackPackDBData.Num) <= len(p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG]) {
			return false
		}

		var eItem *global.ItemDBData
		//检查是否已经装备
		for _, eItem = range p.EquipDBData.EquipData {
			if eItem.Id == Id {
				cfg, ok := global.ServerG.GetConfigMgr().GetCfg("Equipment", eItem.CfgId).(*global.EquipmentCfg)
				if !ok || cfg == nil {
					return false
				}

				delete(p.EquipDBData.EquipData, cfg.SubType)
				break
			}
		}

		if eItem == nil {
			return false
		}

		p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG] = append(p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG], eItem)
		message := &msg.GSCL_PlayerUpdateBackPack{
			AddItems: []*global.ItemDBData{},
		}

		message.AddItems = append(message.AddItems, eItem)

		//重新计算一下属性
		p.updateEquipmentProps(eItem, true)

		p.conn.Send(message)
	}
	p.CaculateFightVal()
	return true
}

const PLAY_Max_Refresh_Time = 4

// 获取免费刷新次数
func (p *player) GetFreeRefreshTimes() int32 {
	timesToday, err := p.playerInfo.GetPlayerResloveTimes(p.dbId)
	if err != nil && err != redis.ErrNil {
		fmt.Println(err)
		return 0
	}

	if err == redis.ErrNil {
		timesToday = 0
	}

	freeTime := PLAY_Max_Refresh_Time - timesToday
	if freeTime <= 0 {
		freeTime = 0
	}
	fmt.Println("免费刷新次数:", freeTime)
	return int32(freeTime)
}

// 获取打造炼化消耗
func (p *player) GetResloveVal(equip_level int32, equip_quality int32) int32 {

	return 0
}

// 获取装备基础属性信息
func (p *player) GetEquipBaseAttribute(cfgId int32) (pkid, pkval int) {
	cfg, err := getEquipCfgById(cfgId)
	if err != nil {
		return
	}
	propkvStrArr := strings.Split(cfg.Data, "+")
	if len(propkvStrArr) < 2 {
		return
	}

	pkid, err = strconv.Atoi(propkvStrArr[0])
	if err != nil {
		return
	}
	pkval, err = strconv.Atoi(propkvStrArr[1])
	if err != nil {
		return
	}

	return
}

func (p *player) EquipCreate() (err error) {
	// 读取打造装备的id
	equipCfgId, err := p.playerInfo.GetEquipCreateCfgIdCache(p.dbId)
	if err != nil {
		return
	}
	resloveCost := p.GetEquipCreateResloveValue(equipCfgId)
	playerReslove, _ := p.GetProp(global.Player_Equip_Reslove)
	if playerReslove < resloveCost {
		fmt.Println("你的熔炼值不住,你的:", playerReslove, "需要:", resloveCost)
		return
	}

	// 扣熔炼值
	retReslove, ok := p.SetProp(global.Player_Equip_Reslove, -resloveCost, true)
	if !ok {
		fmt.Println("失败")
		return
	}

	p.AddItem(equipCfgId, 1, true, true)
	fmt.Println("剩余熔炼值:", retReslove)
	return
}

func (p *player) RefreshNewEquipInfo(useTimes bool, actId int32) (m *msg.GSCL_EquipCreateInfo, err error) {
	equipCfgId, err := p.GetEquipCreateInfo(useTimes)
	if err != nil {
		fmt.Println("nothing")
		return
	}
	costResloveVal := p.GetEquipCreateResloveValue(equipCfgId)
	pkid, pkval := p.GetEquipBaseAttribute(equipCfgId)
	m = &msg.GSCL_EquipCreateInfo{
		EquipCfgId:       equipCfgId,
		CostEquipReslove: costResloveVal,
		FreeTimes:        p.GetFreeRefreshTimes(),
		PkId:             int32(pkid),
		PkVal:            int32(pkval),
		ActId:            actId,
	}
	return
}

// 计算获取装备需要熔炼值
func (p *player) GetEquipCreateResloveValue(equipCfgId int32) (resloveCost int32) {
	return  int32(p.GetEquipCreateResloveVal(equipCfgId))
}

func (p *player) GetEquipCreateInfo(useTimes bool) (equipCfgId int32, err error) {
	if useTimes {
		// 使用次数+1
		times, err := p.playerInfo.IncrByPlayerResloveTimes(p.dbId)
		if err != nil {
			fmt.Println(err)
			return equipCfgId, err
		}

		// 当达到每日最大刷新数
		if times > PLAY_Max_Refresh_Time {
			diamond,ret := p.GetProp(global.Player_Prop_Diamond)
			if !ret {
				err = errorx.READ_PROPS_ERR
				return equipCfgId,err
			}

			if diamond < 100 {
				err = errorx.MONEY_NOT_ENOUGH
				return equipCfgId,err
			}

			p.SetProp(global.Player_Prop_Diamond,-100,true)
		}
	}
	// 获取用户等级相似的装备
	player_level, ret := p.GetProp(global.Player_Prop_Level)
	if !ret {
		return
	}

	cfg := global.ServerG.GetConfigMgr().GetCfg("EquipCreate", int32(1))
	if cfg == nil  {
		fmt.Println("找不到配置")
		err = errorx.CSV_CFG_EMPTY
		return
	}
	equipCreateCfg:= cfg.(*global.EquipCreateCfg)
	q,err := utils.GetResult(1000,equipCreateCfg.Rate)
	if err != nil {
		return
	}
	// q 大于10取大一级的装备
	var quality int16
	equip_lv := int(player_level - (player_level % 10))

	if q > 10 {
		quality = int16(q%10)
		if player_level < 10 {
			equip_lv = 10
		} else {
			equip_lv = equip_lv + 10
		}
	} else {
		quality = int16(q)
		if player_level < 10{
			equip_lv = 1
		}
	}

	query := map[string]interface{}{
		"lv":      []int32{int32(equip_lv)},
		"quality": []int16{quality},
	}

	equipCfg, err := p.EquipQuery(query)
	rand.Seed(time.Now().Unix())
	finalEquip := equipCfg[rand.Intn(len(equipCfg))]

	// 缓存起来,下次打开直接读取
	p.playerInfo.SetPlayerEquipCreateCfgId(p.dbId, finalEquip.Id)
	return finalEquip.Id, nil
}

func getEquipCfgById(equipCfgId int32) (equipCfg *global.EquipmentCfg, err error) {
	equipCfg = global.ServerG.GetConfigMgr().GetCfg("Equipment", equipCfgId).(*global.EquipmentCfg)
	return
}

func (p *player) EquipQuery(query map[string]interface{}) (equipCfg []*global.EquipmentCfg, err error) {
	epCsv := global.ServerG.GetConfigMgr().GetCsv("Equipment")
	if epCsv == nil {
		return equipCfg, errors.New("noEquipCfg")
	}
	equipCfg = []*global.EquipmentCfg{}
	for i := 0; i < epCsv.NumRecord(); i++ {
		epicfg := epCsv.Record(i)
		epCfg, ok := epicfg.(*global.EquipmentCfg)
		if !ok {
			continue
		}

		// 查询特定的装备id
		if query_ids, ok := query["id"]; ok {
			ret := false
			for _, id := range query_ids.([]int) {
				if epCfg.Id == int32(id) {
					ret = true
				}
			}

			if !ret {
				continue
			}
		}

		// 查询特定的装备等级
		if query_lvs, ok := query["lv"]; ok {
			ret := false
			for _, lv := range query_lvs.([]int32) {
				if epCfg.Level == int32(lv) {
					ret = true
				}
			}
			if !ret {
				continue
			}
		}

		// 查询特定装备部位
		if query_subTypes, ok := query["subType"]; ok {
			ret := false
			for _, subType := range query_subTypes.([]int) {
				if epCfg.SubType == int16(subType) {
					ret = true
				}
			}

			if !ret {
				continue
			}
		}

		// 查询装备特定品质
		if query_qualitys, ok := query["quality"]; ok {
			ret := false
			for _, quality := range query_qualitys.([]int16) {
				if epCfg.Quality == int16(quality) {
					ret = true
				}
			}
			if !ret {
				continue
			}
		}
		equipCfg = append(equipCfg, epCfg)
	}

	return
}

// 装备熔炼
func (p *player) EquipReslove(items []int) (int32, bool) {
	// 删除背包装备,并且计算熔炼值
	resloveVal := 0
	updateItems := []*global.ItemDBData{}
	delInstIds := []int32{}
	for _, items_id := range items {
		itemData := p.getItemDataInBagById(int32(items_id))
		if itemData == nil {
			log.Error("player:%d EquipReslove equipment:%d getItemDataInBagById nil", p.GetDBId(), items_id)
			continue
		}

		ui, di, err := p.SubItemByInstId(int32(items_id), 1, false) //sendmsg每次都發，整合一次發
		if err != nil {
			log.Error("player:%d EquipReslove equipment:%d error:%s", p.GetDBId(), items_id, err)
		} else {
			resloveVal += p.GetEquipResloveVal(itemData.CfgId)

			if ui != nil {
				updateItems = append(updateItems, ui...)
			}

			if di != nil {
				delInstIds = append(delInstIds, di...)
			}
		}
	}

	if len(updateItems) > 0 || len(delInstIds) > 0 {
		m := &msg.GSCL_PlayerUpdateBackPack{
			AddItems: updateItems,
			DelItems: delInstIds,
		}

		p.conn.Send(m)
	}
	/*
	for _, items_id := range items {
		for index, itemData := range p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG] {
			if items_id == int(itemData.Id) {
				p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG] = append(p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG][:index],
					p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG][index+1:]...)
				resloveVal += p.GetEquipResloveVal(itemData.CfgId)
				message.DelItems = append(message.DelItems, itemData.Id)
			}
		}
	}
	*/


	return p.SetProp(global.Player_Equip_Reslove, int32(resloveVal), true)
}

// 熔炼装备得到的熔炼值
func (p *player) GetEquipResloveVal(cfgId int32) (val int) {
	bagCfg := global.ServerG.GetConfigMgr().GetCfg("Equipment", cfgId).(*global.EquipmentCfg)
	//resloveVal := 1
	switch bagCfg.Quality {
	case global.EQUIPMENT_QUALITY_WHITE:
		val = 5
	case global.EQUIPMENT_QUALITY_BLUE:
		val = 10
	case global.EQUIPMENT_QUALITY_GREEN:
		val = 20
	case global.EQUIPMENT_QUALITY_PURPLE:
		val = 50
	case global.EQUIPMENT_QUALITY_ORANGE:
		val = 100
	case global.EQUIPMENT_QUALITY_RED:
		val = 200
	}/*
	switch bagCfg.Level {
	case 1:
		resloveVal = 2
	case 10:
		resloveVal = 5
	case 20:
		resloveVal = 10
	case 30:
		resloveVal = 20
	case 40:
		resloveVal = 30
	case 50:
		resloveVal = 40
	case 60:
		resloveVal = 50
	case 70:
		resloveVal = 60
	case 80:
		resloveVal = 70
	}

	r,quality := resloveVal, bagCfg.Quality
	return int(quality) * r*/
	return
}

// 打造装备需要的熔炼值
func (p *player) GetEquipCreateResloveVal(cfgId int32) (val int) {
	bagCfg := global.ServerG.GetConfigMgr().GetCfg("Equipment", cfgId).(*global.EquipmentCfg)
	switch bagCfg.Quality {
	case global.EQUIPMENT_QUALITY_WHITE:
		val = 999
	case global.EQUIPMENT_QUALITY_BLUE:
		val = 999
	case global.EQUIPMENT_QUALITY_GREEN:
		val = 500
	case global.EQUIPMENT_QUALITY_PURPLE:
		val = 2000
	case global.EQUIPMENT_QUALITY_ORANGE:
		val = 8000
	case global.EQUIPMENT_QUALITY_RED:
		val = 25000
	}

	return
}

// 装备升级
func (p *player) EquipUpdate(id int32) (isSuccess bool,pk int32,newVal int32,newEquipLv int32, err error) {
	// todo 强化满级判断
	isSuccess = false
	playerUpdateEquipItem := &global.ItemDBData{}
	// 装备所处位置,1-在背包,2-穿戴身上
	euipLocation := 1
	equip_index := -1

	// 读取装备信息(从背包读取)
	for index, itemData := range p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG] {
		if id == itemData.Id {
			playerUpdateEquipItem = itemData
			equip_index = index
			break
		}
	}

	// 背包读取不到从装备读取
	if equip_index < 0{
		for index, itemData := range p.EquipData {
			if id == itemData.Id {
				playerUpdateEquipItem = itemData
				equip_index = int(index)
				euipLocation = 2
				break
			}
		}
	}

	if playerUpdateEquipItem == nil {
		err = errorx.GET_DATA_EMPTY
		return
	}

	playerEuippropsMap := &global.EquipDBItemData{}
	err = json.Unmarshal([]byte(playerUpdateEquipItem.Data),playerEuippropsMap)
	if err != nil {
		err = errorx.JSON_ERR
		return
	}

	// 升级该装备所需升级石头类型,数量:
	newEquipLv = playerEuippropsMap.UpdateLv + 1
	// 读取装备等级:
	eq_cfg := global.ServerG.GetConfigMgr().GetCfg("Equipment", playerUpdateEquipItem.CfgId)
	if eq_cfg == nil {
		err = errorx.CSV_CFG_EMPTY
		return
	}
	equip_cfg := eq_cfg.(*global.EquipmentCfg)
	up_cfg := global.ServerG.GetConfigMgr().GetCfg("EquipUpdate", newEquipLv)
	if up_cfg == nil {
		err = errorx.CSV_CFG_EMPTY
		return
	}
	equip_update_cfg := up_cfg.(*global.EquipUpdate)
	stoneItemCnt := equip_update_cfg.StoneCnt
	pmoney, ok := p.GetProp(global.Player_Prop_Money)
	if !ok {
		err = errorx.READ_PROPS_ERR
		return
	}

	if pmoney < equip_update_cfg.Money {
		err = errorx.MONEY_NOT_ENOUGH
		return
	}

	_, ok = p.SetProp(global.Player_Prop_Money, -equip_update_cfg.Money, true)
	if !ok {
		err = errorx.SET_MONEY_PROPS_ERR
		return
	}

	playerHadItem := false
	// 读取用户身上道具数量,并且扣除
	for _, itemData := range p.BackPackDBData.BagData[global.BACKPACK_ITEMS_BAG] {
		if equip_update_cfg.ItemId == itemData.CfgId && itemData.Num >= stoneItemCnt {
			p.AddItem(itemData.CfgId, -stoneItemCnt, true, true)
			playerHadItem = true
			break
		}
	}

	if !playerHadItem {
		err = errorx.NOT_ENOUGH_ITEM
		return
	}


	// 升级:
	canUpdate := utils.HundredRandomTrue(equip_update_cfg.SuccRate)
	if !canUpdate {
		return isSuccess,0,0,playerEuippropsMap.UpdateLv,nil
	}

	// 升级数据: 基础值 * 成长
	baseAttrArr := strings.Split(equip_cfg.Data, "+")
	baseProId,err := strconv.Atoi(baseAttrArr[0])
	if err != nil {
		// TODO:打日志
		return
	}
	pk = int32(baseProId)
	baseVal,err := strconv.ParseFloat(baseAttrArr[1],32)
	if err != nil {
		// TODO:打日志
		return
	}
	isSuccess = true
	newVal = int32(math.Ceil(baseVal * equip_update_cfg.UpdateGrowth))
	oldVal := playerEuippropsMap.BData[pk]
	playerEuippropsMap.BData[pk] = newVal
	playerEuippropsMap.UpdateLv = newEquipLv
	bData, err := json.Marshal(playerEuippropsMap)
	playerUpdateEquipItem.Data = string(bData)
	if euipLocation == 1{
		p.BackPackDBData.BagData[global.BACKPACK_EQUIP_BAG][equip_index] = playerUpdateEquipItem
	} else {
		p.EquipData[int16(equip_index)] = playerUpdateEquipItem
		addProps := newVal - oldVal
		fmt.Println("装备升级属性增加:",addProps)
		p.SetProp(baseProId,addProps,true)
		p.CaculateFightVal()
	}
	return
}