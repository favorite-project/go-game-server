package object

import (
	_ "encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"xianxia/common/dbengine"
	"xianxia/servers/game/conf"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/game/global/db"
	"xianxia/servers/game/game/model"
	"xianxia/servers/game/msg"
	"github.com/garyburd/redigo/redis"
	"github.com/name5566/leaf/log"
	"encoding/json"
	"xianxia/servers/game/utils"
)

const CREATE_PIC = "com_head"

const (
	EVENTYPE_DB_RET_SET_BACKPACK  = 1
	EVENTYPE_DB_RET_GET_BACKPACK  = 2
	EVENTYPE_DB_RET_SET_PROPS     = 3
	EVENTYPE_DB_RET_GET_EQUIPMENT = 4
	EVENTYPE_DB_RET_SET_EQUIPMENT = 5
	EVENTYPE_DB_RET_GET_SKILL     = 6
	EVENTYPE_DB_RET_SET_SKILL     = 7
	EVENTYPE_DB_RET_GET_CDKEY    = 8
	EVENTYPE_DB_RET_SET_CDKEY    = 9
	EVENTYPE_DB_RET_SET_RANK    = 10
	EVENTYPE_DB_RET_SET_MINE_POOL = 11
)

type PlayerProps [global.Player_Prop_Max - global.Creature_Prop_Two_Max]int32
type PlayerBitFlag [global.Player_Prop_Max - global.Creature_Prop_Two_Max]int32
type player struct {
	creature
	dbId    int64
	props   PlayerProps
	bitFlag PlayerBitFlag
	conn    global.Connection

	*global.BackPackDBData
	*global.EquipDBData
	nextNormalFightTime  int64             //下一次普通战斗时间
	lastNormalFightReard global.RewardData //普通战斗的奖励

	online          bool
	lastOnlineTime  int32
	lastOfflineTime int32
	createTime      int32
	loadDataTime    int64
	playerInfo      model.PlayerInfo
	openId string
	deviceType string
	loginType int

	nextInstanceFightTime  int64             //下一次副本战斗时间
	lastInstanceFightReard global.RewardData //副本战斗的奖励

	lastHeartBeatTimeSec int64 //上一次心跳包时间
	bKicked bool

	lastSaveDBTime int64 //上一次保存db的时间

	//cdkey
	*global.Player_CDKesy_Info
	LockState bool

	//mine
	curRobCfgId int32
	curRobReward global.RewardData //挖矿抢劫的奖励

	//challenge
	curChallengeId  int32             //当前挑战的id
	lastChallengeFightReard global.RewardData //挑战的奖励
}

func (p *player) initProps(dbData *db.DB_PLayer_Props) bool {
	if dbData == nil {
		return false
	}
	p.playerInfo.DB = global.ServerG.GetDBEngine()
	p.initBase()

	//一级属性
	/*
	p.creature.props[global.Creature_Prop_One_Power] = dbData.Power
	p.creature.props[global.Creature_Prop_One_Agile] = dbData.Agile
	p.creature.props[global.Creature_Prop_One_Strength] = dbData.Strength
	p.creature.props[global.Creature_Prop_One_Intelligence] = dbData.Intelligence
	p.creature.props[global.Creature_Prop_One_Lucky] = dbData.Lucky
	p.creature.props[global.Creature_Prop_One_Endurance] = dbData.Endurance



	//todo 计算出二级属性
	p.caculateProps(global.Creature_Prop_One_Power, dbData.Power)
	p.caculateProps(global.Creature_Prop_One_Agile, dbData.Agile)
	p.caculateProps(global.Creature_Prop_One_Strength, dbData.Strength)
	p.caculateProps(global.Creature_Prop_One_Intelligence, dbData.Intelligence)
	p.caculateProps(global.Creature_Prop_One_Lucky, dbData.Lucky)
	p.caculateProps(global.Creature_Prop_One_Endurance, dbData.Endurance)
*/
	//通过等级计算一级和二级属性
	p.SetProp(global.Player_Prop_Level, dbData.Level, false)

	//设置vip经验，并计算出等级
	p.SetProp(global.Player_Prop_VipExp, dbData.VipExp, false)

	//测试一下暴击
	//p.SetProp(global.Creature_Prop_Two_Crit, 1000, false)

	//基本数据
	p.creature.cType = global.Creature_Type_Player
	p.creature.mapId = dbData.MapId
	p.creature.regionId = dbData.RegionId
	copy(p.creature.name[:len(p.creature.name)], dbData.Name)

	//角色属性

	//p.props[global.Player_Prop_Level-global.Player_Prop_Level] = dbData.Level
	p.props[global.Player_Prop_Money-global.Player_Prop_Level] = dbData.Money
	p.props[global.Player_Prop_Diamond-global.Player_Prop_Level] = dbData.Diamond
	p.props[global.Player_Prop_Sex-global.Player_Prop_Level] = dbData.Sex
	p.props[global.Player_Prop_Occupation-global.Player_Prop_Level] = dbData.Occupation
	p.props[global.Player_Prop_Exp-global.Player_Prop_Level] = dbData.Exp
	p.props[global.Player_Prop_FreeFightCount-global.Player_Prop_Level] = dbData.FreeFightCount
	p.props[global.Player_Prop_MaxMapId-global.Player_Prop_Level] = dbData.MaxMapId
	p.props[global.Player_Prop_MaxRegionId-global.Player_Prop_Level] = dbData.MaxRegionId
	p.props[global.Player_Equip_Reslove-global.Player_Prop_Level] = dbData.EquipReslove
	p.props[global.Player_Prop_SkillPoint-global.Player_Prop_Level] = dbData.SkillPoint
	p.props[global.Player_Prop_QuickFightCount-global.Player_Prop_Level] = dbData.QuickFightCount
	p.props[global.Player_Prop_QuickFightTime-global.Player_Prop_Level] = dbData.QuickFightTime
	p.props[global.Player_Prop_RechargeNum-global.Player_Prop_Level] = dbData.RechargeNum
	p.props[global.Player_Prop_FightVal-global.Player_Prop_Level] = dbData.FightVal

	p.SetProp(global.Player_Prop_Advance_Level, dbData.AdvanceLevel, true)
	p.props[global.Player_Prop_Advance_Exp-global.Player_Prop_Level] = dbData.AdvanceExp

	p.props[global.Player_Prop_VipReward_Time-global.Player_Prop_Level] = dbData.VipRewardTime

	p.createTime = dbData.CreateTime
	p.lastOnlineTime = dbData.LastOnlineTime
	p.lastOfflineTime = dbData.LastOffTimeTime
	p.deviceType = dbData.DeviceType
	p.openId = dbData.OpenId
	p.loginType = dbData.LoginType

	copy(p.creature.pic[:len(p.creature.pic)], dbData.Pic)

	p.lastHeartBeatTimeSec = time.Now().Unix()

	p.CaculateFightVal()

	return true
}

func (p *player) fromProps() *db.DB_PLayer_Props {
	propsDBData := &db.DB_PLayer_Props{
		Name: make([]byte, global.NAME_MAX_LEN),
		Pic:  make([]byte, global.NAME_MAX_LEN),
	}

	/*
	propsDBData.Power, _ = p.GetProp(global.Creature_Prop_One_Power)
	propsDBData.Agile, _ = p.GetProp(global.Creature_Prop_One_Agile)
	propsDBData.Strength, _ = p.GetProp(global.Creature_Prop_One_Strength)
	propsDBData.Intelligence, _ = p.GetProp(global.Creature_Prop_One_Intelligence)
	propsDBData.Lucky, _ = p.GetProp(global.Creature_Prop_One_Lucky)
	propsDBData.Endurance, _ = p.GetProp(global.Creature_Prop_One_Endurance)
	*/

	propsDBData.Level, _ = p.GetProp(global.Player_Prop_Level)
	propsDBData.Money, _ = p.GetProp(global.Player_Prop_Money)
	propsDBData.Diamond, _ = p.GetProp(global.Player_Prop_Diamond)
	propsDBData.VipLevel, _ = p.GetProp(global.Player_Prop_VipLevel)
	propsDBData.Occupation, _ = p.GetProp(global.Player_Prop_Occupation)
	propsDBData.Sex, _ = p.GetProp(global.Player_Prop_Sex)
	propsDBData.Exp, _ = p.GetProp(global.Player_Prop_Exp)
	propsDBData.FreeFightCount, _ = p.GetProp(global.Player_Prop_FreeFightCount)
	propsDBData.MaxMapId, _ = p.GetProp(global.Player_Prop_MaxMapId)
	propsDBData.MaxRegionId, _ = p.GetProp(global.Player_Prop_MaxRegionId)
	propsDBData.EquipReslove, _ = p.GetProp(global.Player_Equip_Reslove)
	propsDBData.SkillPoint, _ = p.GetProp(global.Player_Prop_SkillPoint)
	propsDBData.QuickFightCount, _ = p.GetProp(global.Player_Prop_QuickFightCount)
	propsDBData.QuickFightTime, _ = p.GetProp(global.Player_Prop_QuickFightTime)
	propsDBData.RechargeNum, _= p.GetProp(global.Player_Prop_RechargeNum)
	propsDBData.FightVal, _= p.GetProp(global.Player_Prop_FightVal)
	propsDBData.VipExp, _= p.GetProp(global.Player_Prop_VipExp)
	propsDBData.AdvanceLevel, _= p.GetProp(global.Player_Prop_Advance_Level)
	propsDBData.AdvanceExp, _= p.GetProp(global.Player_Prop_Advance_Exp)
	propsDBData.VipRewardTime, _= p.GetProp(global.Player_Prop_VipReward_Time)

	//propsDBData.CreateTime
	propsDBData.MapId = p.creature.mapId
	propsDBData.RegionId = p.creature.regionId
	propsDBData.DBId = p.dbId
	propsDBData.LastOnlineTime = p.lastOnlineTime
	propsDBData.LastOffTimeTime = p.lastOfflineTime
	propsDBData.CreateTime = p.createTime
	propsDBData.OpenId = p.openId
	propsDBData.DeviceType = p.deviceType
	propsDBData.LoginType = p.loginType
	propsDBData.Lock = p.LockState

	copy(propsDBData.Name, p.creature.name[:len(p.creature.name)])
	copy(propsDBData.Pic, p.creature.pic[:len(p.creature.pic)])

	return propsDBData
}

func (p *player) initFromDB(dbData *db.DB_PLayer_Props) {
	p.initProps(dbData)

	p.loadDataTime = time.Now().UnixNano()

	p.GetPlayerEquipment()
}

func (p *player) create() error {
	p.playerInfo.DB = global.ServerG.GetDBEngine()

	bCsv := global.ServerG.GetConfigMgr().GetCsv("Map")
	if bCsv == nil {
		log.Error("CreatePlayer Map Config error")
		return errors.New("CreatePlayer Map Config error")
	}

	//取第一张地图进行初始化
	micfg := bCsv.Record(0)
	if micfg == nil {
		log.Error("CreatePlayer Map Config Record 0 error")
		return errors.New("CreatePlayer Map Config Record 0 error")
	}

	mcfg := micfg.(*global.MapCfg)
	regionIdArr := strings.Split(mcfg.Regions, "+")
	if len(regionIdArr) <= 0 {
		log.Error("CreatePlayer Map Config Record 0 has no region")
		return errors.New("CreatePlayer Map Config Record 0  has no region")
	}

	rid, err := strconv.Atoi(regionIdArr[0])
	if err != nil {
		log.Error("CreatePlayer Map Config Record 0 region Atoi error")
		return errors.New("CreatePlayer Map Config Record 0 region Atoi error")
	}

	cRegionId := int32(rid)

	id, err := global.ServerG.GetDBEngine().GetUniqueID()
	if err != nil {
		log.Error("CreatePlayer GetUniqueID Error:%p", err)
		return errors.New("CreatePlayer GetUniqueID Error")
	}

	//基本数据
	p.creature.cType = global.Creature_Type_Player
	p.dbId = int64(id)
	p.initBase()

	//设置地图
	p.creature.mapId = mcfg.Id
	p.creature.regionId = cRegionId
	p.SetProp(global.Player_Prop_MaxMapId, mcfg.Id, false)
	p.SetProp(global.Player_Prop_MaxRegionId, cRegionId, false)
	p.createTime = int32(time.Now().UnixNano() / int64(time.Second))
	p.lastOnlineTime = int32(0)
	p.lastOfflineTime = int32(0)

	copy(p.creature.name[:len(p.creature.name)], []byte(fmt.Sprintf("玩家%d", p.dbId)))
	copy(p.creature.pic[:len(p.creature.pic)], CREATE_PIC)

	//默认初始化是1级别
	p.SetProp(global.Player_Prop_Level, 1, true)

	p.lastHeartBeatTimeSec = time.Now().Unix()

	p.loadDataTime = time.Now().UnixNano()

	//装备数据
	p.initEquipData()

	//技能数据
	p.initSkillData()

	p.CaculateFightVal()

	return nil
}

func (p *player) Update(now time.Time, elsp int64) {
	//检测心跳包
	if p.bKicked {
		return
	} else {
		if  now.Unix() - p.lastHeartBeatTimeSec >= global.HEARTBEAT_SEC * 2 {
			p.Kick()
			p.bKicked = true
			return
		}
	}

	p.creature.Update(now, elsp)

	//检测快速战斗是否跨天
	qfTime, _ := p.GetProp(global.Player_Prop_QuickFightTime)
	if qfTime == 0 {
		qfTime = int32(now.Unix())
		p.SetProp(global.Player_Prop_QuickFightTime, qfTime, false)
	} else {
		if !utils.CheckIsSameDayBySec(now.Unix(), int64(qfTime), 0) {
			p.SetProp(global.Player_Prop_QuickFightTime, int32(now.Unix()), false)
			p.SetProp(global.Player_Prop_QuickFightCount, 0, false)
		}
	}

	//广播变动的属性
	if p.conn != nil && p.bitflags != nil && len(p.bitflags) > 0 {
		m := &msg.GSCL_PlayerUpdateProps{
			Props: make(map[int32]int32),
		}

		for pi, _ := range p.bitflags {
			m.Props[int32(pi)], _ = p.GetProp(pi)
			delete(p.bitflags, pi)
		}

		p.conn.Send(m)
	}

	//每隔1分钟保存一次数据库
	if p.lastSaveDBTime == 0 {
		p.lastSaveDBTime = now.UnixNano()
	} else {
		if now.UnixNano() - p.lastSaveDBTime > int64(time.Minute) {
			p.lastSaveDBTime = now.UnixNano()
			p.OnSave()
		}
	}
}

func (p *player) GetProp(index int) (int32, bool) {
	value, suc := p.creature.GetProp(index)
	if suc {
		return value, suc
	}

	if index >= global.Player_Prop_Max || index < global.Player_Prop_Level {
		return 0, false
	}

	return p.props[index-global.Player_Prop_Level], true
}

func getPropsArr(str string) map[int]int32 {
	result := make(map[int]int32)

	if len(str) == 0 {
		return result
	}

	propArr := strings.Split(str, ";")
	for _, propItemStr := range propArr {
		propItemArr := strings.Split(propItemStr, "+")
		if len(propItemArr) != 2 {
			continue
		}

		k, err := strconv.Atoi(propItemArr[0])
		if err != nil {
			continue
		}

		v, err := strconv.Atoi(propItemArr[1])
		if err != nil {
			continue
		}

		result[k] = int32(v)
	}

	return result
}

func (p *player) SetProp(index int, value int32, bAdd bool) (int32, bool) {
	v, suc := p.creature.SetProp(index, value, bAdd)
	if suc {
		return v, suc
	}

	if index >= global.Player_Prop_Max || index < global.Player_Prop_Level {
		return 0, false
	}

	ov := p.props[index-global.Player_Prop_Level]

	if bAdd {
		p.props[index-global.Player_Prop_Level] += value
	} else {
		p.props[index-global.Player_Prop_Level] = value
	}

	//等级和经验药特殊判断
	if index == global.Player_Prop_Exp {
		for {
			level, _ := p.GetProp(global.Player_Prop_Level)
			bcfg := global.ServerG.GetConfigMgr().GetCfg("PlayerLevel", level)
			if bcfg == nil {
				return 0, false
			}

			cfg := bcfg.(*global.PlayerLevelCfg)
			if p.props[index-global.Player_Prop_Level] >= cfg.NeedExp { //满足升级经验
				bcfg = global.ServerG.GetConfigMgr().GetCfg("PlayerLevel", level+1)
				if bcfg == nil { //满级了
					p.props[index-global.Player_Prop_Level] = cfg.NeedExp
					break
				} else {
					p.SetProp(global.Player_Prop_Level, 1, true)
					p.SetProp(global.Player_Prop_SkillPoint, cfg.AddSkillPoint, true)
					p.props[index-global.Player_Prop_Level] -= cfg.NeedExp
				}
			} else {
				break
			}
		}
	} else if index == global.Player_Prop_Level {
		bcfg := global.ServerG.GetConfigMgr().GetCfg("PlayerLevel", p.props[index-global.Player_Prop_Level])
		if bcfg == nil { //等级不存在或者已经满级了
			p.props[index-global.Player_Prop_Level] = ov
			return 0, false
		}

		if p.props[index-global.Player_Prop_Level] != ov { //重新计算等级相关的属性
			bCsv := global.ServerG.GetConfigMgr().GetCsv("PlayerLevel")
			if bCsv != nil {
				ld := int32(1)
				begin := ov
				end := p.props[index-global.Player_Prop_Level]
				if p.props[index-global.Player_Prop_Level] < ov {
					ld = -ld
					begin = p.props[index-global.Player_Prop_Level]
					end = ov
				}

				for i := begin; i < end; i++ {
					bcfg := bCsv.Record(int(i))
					if bcfg == nil {
						p.props[index-global.Player_Prop_Level] = ov
						return 0, false
					}

					cfg := bcfg.(*global.PlayerLevelCfg)
					propArr := strings.Split(cfg.AddProps, ";")
					for _, propItemStr := range propArr {
						propItemArr := strings.Split(propItemStr, "+")
						if len(propItemArr) != 2 {
							p.props[index-global.Player_Prop_Level] = ov
							return 0, false
						}

						k, err := strconv.Atoi(propItemArr[0])
						if err != nil {
							p.props[index-global.Player_Prop_Level] = ov
							return 0, false
						}

						v, err := strconv.Atoi(propItemArr[1])
						if err != nil {
							p.props[index-global.Player_Prop_Level] = ov
							return 0, false
						}

						p.SetProp(k, int32(v)*ld, true)
					}
				}
			}
		}
	} else if index == global.Player_Prop_VipExp {
		//检查是否达成vip升级条件
		for {
			vipLevel, _ := p.GetProp(global.Player_Prop_VipLevel)
			cfg := global.ServerG.GetConfigMgr().GetCfg("Vip", vipLevel + 1)
			if cfg == nil {
				break
			}

			icfg := cfg.(*global.VipCfg)
			if p.props[index - global.Player_Prop_Level] >= icfg.Recharge {
				p.SetProp(global.Player_Prop_VipLevel, vipLevel + 1, false)
			} else {
				break
			}
		}
	} else if index == global.Player_Prop_Advance_Level {
		oicfg := global.ServerG.GetConfigMgr().GetCfg("Advance", ov)
		oparr := make(map[int]int32)
		if oicfg != nil {
			oparr = getPropsArr(oicfg.(*global.AdvanceCfg).AddProps)
		}

		icfg := global.ServerG.GetConfigMgr().GetCfg("Advance", p.props[index-global.Player_Prop_Level])
		if icfg != nil {
			parr := getPropsArr(icfg.(*global.AdvanceCfg).AddProps)
			for ppid, ppv := range parr {
				if oppv, ok := oparr[ppid]; ok {
					ppv -= oppv
				}

				p.SetProp(ppid, ppv, true)
			}
		}
	}

	if p.bitflags != nil && ov != p.props[index-global.Player_Prop_Level] {
		p.bitflags[index] = byte(1)
	}

	return p.props[index-global.Player_Prop_Level], true
}

func (p *player) SetConnection(conn global.Connection) {
	p.conn = conn
}

func (p *player) GetConnection() global.Connection {
	return p.conn
}

func (p *player) OnSave() {
	//保存基本数据
	dbData := p.fromProps()
	global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_PROPS, int64(0),"hmset", redis.Args{}.Add(fmt.Sprintf("player:%d", p.dbId)).AddFlat(dbData)...)

	//保存背包
	p.SaveBackPack()

	//保存装备
	p.SaveEquip()

	//保存技能
	p.SaveSkill()

	//保存cdkey
	p.SaveCDKey()
}

func (p *player) GetDBId() int64 {
	return p.dbId
}

func (p *player) GetPublicProps(publicContent *global.PlayerPublicProps) {
	if publicContent == nil {
		return
	}

	publicContent.Power, _ = p.GetProp(global.Creature_Prop_One_Power)
	publicContent.Agile, _ = p.GetProp(global.Creature_Prop_One_Agile)
	publicContent.Intelligence, _ = p.GetProp(global.Creature_Prop_One_Intelligence)
	publicContent.Strength, _ = p.GetProp(global.Creature_Prop_One_Strength)
	publicContent.Lucky, _ = p.GetProp(global.Creature_Prop_One_Lucky)
	publicContent.Endurance, _ = p.GetProp(global.Creature_Prop_One_Endurance)

	publicContent.Attack, _ = p.GetProp(global.Creature_Prop_Two_Attack)
	publicContent.Crit, _ = p.GetProp(global.Creature_Prop_Two_Crit)
	publicContent.AttackSpeed, _ = p.GetProp(global.Creature_Prop_Two_AttackSpeed)
	publicContent.Miss, _ = p.GetProp(global.Creature_Prop_Two_Miss)
	publicContent.Blood, _ = p.GetProp(global.Creature_Prop_Two_Blood)
	publicContent.Get, _ = p.GetProp(global.Creature_Prop_Two_Get)
	publicContent.Defence, _ = p.GetProp(global.Creature_Prop_Two_Defence)
	publicContent.Magic, _ = p.GetProp(global.Creature_Prop_Two_Magic)

	copy(publicContent.Name[:global.NAME_MAX_LEN], p.GetName())
	publicContent.Level, _ = p.GetProp(global.Player_Prop_Level)
	publicContent.Occupation, _ = p.GetProp(global.Player_Prop_Occupation)
	publicContent.Sex, _ = p.GetProp(global.Player_Prop_Sex)
	copy(publicContent.Pic[:global.PIC_MAX_LEN], p.creature.pic[:])
	publicContent.MapId, publicContent.RegionId = p.GetMapId()
	publicContent.DBId = p.dbId
}

func (p *player) GetPrivateProps(privateContent *global.PlayerPrivateProps) {
	if privateContent == nil {
		return
	}

	p.GetPublicProps(&privateContent.PlayerPublicProps)
	privateContent.Money, _ = p.GetProp(global.Player_Prop_Money)
	privateContent.Diamond, _ = p.GetProp(global.Player_Prop_Diamond)
	privateContent.VipLevel, _ = p.GetProp(global.Player_Prop_VipLevel)
	privateContent.Exp, _ = p.GetProp(global.Player_Prop_Exp)
	privateContent.MaxMapId, _ = p.GetProp(global.Player_Prop_MaxMapId)
	privateContent.MaxRegionId, _ = p.GetProp(global.Player_Prop_MaxRegionId)
	privateContent.EquipReslove, _ = p.GetProp(global.Player_Equip_Reslove)
	privateContent.SkillPoint, _ = p.GetProp(global.Player_Prop_SkillPoint)
	privateContent.QuickFightCount, _ = p.GetProp(global.Player_Prop_QuickFightCount)
	privateContent.QuickFightTime, _ = p.GetProp(global.Player_Prop_QuickFightTime)
	privateContent.RechargeNum, _ = p.GetProp(global.Player_Prop_RechargeNum)
	privateContent.FightVal, _ = p.GetProp(global.Player_Prop_FightVal)
	privateContent.VipExp, _ = p.GetProp(global.Player_Prop_VipExp)
	privateContent.AdvanceLevel, _ = p.GetProp(global.Player_Prop_Advance_Level)
	privateContent.AdvanceExp, _ = p.GetProp(global.Player_Prop_Advance_Exp)
	privateContent.VipRewardTime, _ = p.GetProp(global.Player_Prop_VipReward_Time)
}

func (p *player) OnRecv(recvData []byte) {
	if len(recvData) < 4 {
		return
	}

	subModule := msg.RdWrEndian.Uint32(recvData)
	switch subModule {
	case global.Message_RootKey_HeartBeat: //心跳包
		p.lastHeartBeatTimeSec = time.Now().Unix()

		m := &msg.GSCL_HeartBeat{
			NowSec: int32(p.lastHeartBeatTimeSec),
		}

		p.conn.Send(m)
	case global.Message_RootKey_Player_Add_BackPack_Item:
		//非测试模式此接口不开放
		if !conf.TestMode {
			return
		}

		if len(recvData[4:]) < 8 {
			return
		}

		itemId := msg.RdWrEndian.Uint32(recvData[4:])
		num := msg.RdWrEndian.Uint32(recvData[8:])
		p.AddItem(int32(itemId), int32(num), true, true)
	case global.Message_RootKey_Player_Fight:
		//在副本中不允许战斗
		if len(recvData[4:]) < 5 {
			return
		}

		fightMode := conf.RdWrEndian.Uint32(recvData[4:])
		if !p.checkCanFight(fightMode) {
			return
		}

		bBoss := recvData[8] == byte(1)
		cmons :=  []global.Creature{}
		instanceId := p.GetInstanceMapId()

		if fightMode == global.FIGHT_MODE_NORMAL {
			if instanceId != 0 {
				return
			}

			mid, rid := p.GetMapId()
			cmap := global.ServerG.GetMapMgr().GetMap(mid)
			if cmap == nil {
				return
			}

			mons := cmap.GetRegionMonsters(rid, bBoss)
			if mons == nil {
				return
			}

			for _, mon := range mons {
				cmons = append(cmons, mon)
			}
		} else if fightMode == global.FIGHT_MODE_INSTANCE {
			cmap := global.ServerG.GetMapMgr().GetMap(instanceId)
			if cmap == nil {
				return
			}

			mons, boss := cmap.GetInstanceMonsters(p.GetInstanceFightIndex())
			if mons == nil {
				return
			}

			for _, mon := range mons {
				cmons = append(cmons, mon)
			}

			bBoss = boss
		} else if fightMode == global.FIGHT_MODE_CHALLENGE {
			if len(recvData[4:]) < 9 {
				return
			}

			challengeId := int32(conf.RdWrEndian.Uint32(recvData[9:]))
			level_id,err := p.GetPlayerChallengeInfo(challengeId)
			if err != nil {
				return
			}
			next_lv_id := level_id + 1
			mons, boss := p.GetChallengeMonsters(challengeId,next_lv_id)
			if mons == nil {
				return
			}

			for _, mon := range mons {
				cmons = append(cmons, mon)
			}
			bBoss = boss

			p.curChallengeId = challengeId

		}  else if fightMode == global.FIGHT_MODE_MINE_ROB {
			if len(recvData[4:]) < 17 {
				return
			}

			robDBId := int64(conf.RdWrEndian.Uint64(recvData[9:]))
			p.curRobCfgId = int32(conf.RdWrEndian.Uint32(recvData[17:]))

			if robDBId == p.dbId {
				return
			}
			
			if p.playerInfo.GetMineRobTimeKey(p.dbId, robDBId, p.curRobCfgId) {
				return
			}

			var robPlayer global.Creature = global.ServerG.GetObjectMgr().GetPlayer(robDBId)
			if robPlayer == nil {
				robPlayer = global.ServerG.GetObjectMgr().GetOfflinePlayer(robDBId) //取离线数据
				if robPlayer == nil {
					log.Error("Player::OnRecv Fight MineRob dbid:%d empty", robDBId)
					return
				}
			}

			cmons = append(cmons, robPlayer)
		} else if fightMode == global.FIGHT_MODE_ADVANCE {
			cmons = p.getAdvanceBoss()
			if cmons == nil {
				return
			}
		}else {
			return
		}

		content := global.ServerG.GetFightMgr().DoFight(p, []global.Creature{p}, cmons, fightMode, bBoss)
		if content == nil {
			return
		}

		cmsg := &msg.GSCL_PlayerFightInfo{
			Mode: int32(fightMode),
			InstanceEnd:bBoss,
			Data: content,
		}

		if content.AttackWin {
			cmsg.Data.Reward = &p.lastNormalFightReard
			if fightMode == global.FIGHT_MODE_INSTANCE {
				if bBoss {
					cmsg.Data.Reward = &p.lastInstanceFightReard
				}
			} else  if fightMode == global.FIGHT_MODE_CHALLENGE {
				cmsg.Data.Reward = &p.lastChallengeFightReard
			} else if fightMode == global.FIGHT_MODE_MINE_ROB {
				cmsg.Data.Reward = &p.curRobReward

				robDBId := int64(conf.RdWrEndian.Uint64(recvData[9:]))
				p.playerInfo.SetMineRobTimeKey(p.dbId, robDBId, p.curRobCfgId, p.GetVipEffectValue(global.Vip_Effect_MineRobTime))
			}
		} else {
			if fightMode == global.FIGHT_MODE_INSTANCE  {
				cmsg.InstanceEnd = true
			}
		}

		p.conn.Send(cmsg)
	case global.Message_RootKey_Player_Equip_ACT:
		if len(recvData[4:]) < 5 {
			return
		}

		itemId := msg.RdWrEndian.Uint32(recvData[4:])
		bEquip := recvData[8] == byte(1)
		p.Equip(int32(itemId), bEquip)
	case global.Message_RootKey_Player_UseItem:
		if len(recvData[4:]) < 10 {
			return
		}

		itemId := int32(msg.RdWrEndian.Uint32(recvData[4:]))
		num := int32(msg.RdWrEndian.Uint32(recvData[8:]))
		useType := msg.RdWrEndian.Uint16(recvData[12:])
		p.UseItem(itemId, num, useType)
	case global.Message_RootKey_Player_Fight_Reward:
		if len(recvData[4:]) < 4 {
			return
		}

		fightMode := conf.RdWrEndian.Uint32(recvData[4:])
		if fightMode == global.FIGHT_MODE_NORMAL {
			_, _, sellInfo, _ := p.AddItems(&p.lastNormalFightReard, true, true)
			m := &msg.GSCL_PlayerFightReward{
				SellInfo: sellInfo,
			}

			p.conn.Send(m)
			p.lastNormalFightReard.Items = nil
		} else if fightMode == global.FIGHT_MODE_INSTANCE {
			_, _, sellInfo, _ := p.AddItems(&p.lastInstanceFightReard, true, true)
			m := &msg.GSCL_PlayerFightReward{
				SellInfo: sellInfo,
			}

			p.conn.Send(m)
			p.lastInstanceFightReard.Items = nil

		} else if fightMode == global.FIGHT_MODE_CHALLENGE {
			_, _, sellInfo, _ := p.AddItems(&p.lastChallengeFightReard, true, true)
			m := &msg.GSCL_PlayerFightReward{
				SellInfo: sellInfo,
			}

			p.conn.Send(m)
			p.lastChallengeFightReard.Items = nil
			now_level_id,_ := p.GetPlayerChallengeInfo(p.curChallengeId)
			new_lv_id := now_level_id + 1
			p.ChallengeEnd(p.dbId, p.curChallengeId, new_lv_id)
			p.curChallengeId = 0
		} else if fightMode == global.FIGHT_MODE_MINE_ROB {
			_, _, sellInfo, _ := p.AddItems(&p.curRobReward, true, true)
			m := &msg.GSCL_PlayerFightReward{
				SellInfo: sellInfo,
			}

			p.conn.Send(m)
			p.curRobReward.Items = nil
		}
	case global.Message_RootKey_Player_Change_Map:
		if len(recvData[4:]) < 8 {
			return
		}
		cMapId := int32(conf.RdWrEndian.Uint32(recvData[4:]))
		cRegionId := int32(conf.RdWrEndian.Uint32(recvData[8:]))
		global.ServerG.GetMapMgr().ChangeMap(p, cMapId, cRegionId)
	case global.Message_RootKey_Equip_Resolve:
		resolveCnt := int(msg.RdWrEndian.Uint32(recvData[4:]))
		if resolveCnt <= 0 || resolveCnt > 6 {
			fmt.Println("resolvcn小于0:", resolveCnt)
			return
		}
		fmt.Println("cnt:", resolveCnt, "长度:", len(recvData))
		itemId := []int{}
		for i := 1; i <= resolveCnt; i++ {
			itemId = append(itemId, int(msg.RdWrEndian.Uint32(recvData[((i+1)*4):])))
		}

		resloveVal, _ := p.EquipReslove(itemId)
		m := &msg.GSCL_PlayerEquipResloveVal{
			ResloveVal: resloveVal,
		}
		p.conn.Send(m)
	case global.Message_RootKey_EquipCreate_Info:
		equipCfgId, err := p.playerInfo.GetEquipCreateCfgIdCache(p.dbId)
		if err != nil && err != redis.ErrNil {
			fmt.Println("err:", err)
			return
		}
		var costResloveVal int32
		// 查不到数据说明是当天第一次打开
		if err == redis.ErrNil {
			equipCfgId, err = p.GetEquipCreateInfo(true)
			if err != nil {
				fmt.Println("errx:", err)
				return
			}
		}
		costResloveVal = p.GetEquipCreateResloveValue(equipCfgId)
		pkid, pkval := p.GetEquipBaseAttribute(equipCfgId)
		m := &msg.GSCL_EquipCreateInfo{
			EquipCfgId:       equipCfgId,
			CostEquipReslove: costResloveVal,
			FreeTimes:        p.GetFreeRefreshTimes(),
			PkId:             int32(pkid),
			PkVal:            int32(pkval),
			ActId:            global.LOADEQUIPINFO_ACT,
		}
		p.conn.Send(m)
	case global.Message_RootKey_EquipCreate_Refresh:
		// 用户刷新操作
		m, err := p.RefreshNewEquipInfo(true, global.REFREASH_EQUIPINFO_ACT)
		if err != nil {
			return
		}
		p.conn.Send(m)
	case global.Message_RootKey_EquipCreate:
		// 打造装备
		p.EquipCreate()
		// 返回下一把装备信息
		m, err := p.RefreshNewEquipInfo(false, global.CREATEQUIP_ACT)
		if err != nil {
			return
		}
		p.conn.Send(m)
	case global.Message_RootKey_Player_StudySkill:
		p.skill_study(recvData[4:])
	case global.Message_RootKey_Player_SkillLvUp:
		p.skill_levelUp(recvData[4:])
	case global.Message_RootKey_Player_Change_SkillPos:
		p.skill_changePos(recvData[4:])
	case global.Message_RootKey_Player_SkillEquip:
		p.skill_equip(recvData[4:])
	case global.Message_RootKey_Player_SkillUnEquip:
		p.skill_unequip(recvData[4:])
	case global.Message_RootKey_Player_AutoSellEquip:
		p.equipAutoSell(recvData[4:])
	case global.Message_RootKey_RandomStone:
		stoneMode := int32(msg.RdWrEndian.Uint32(recvData[4:]))
		timesMode := msg.RdWrEndian.Uint32(recvData[8:])
		itemIds, err := p.RandomStone(stoneMode, int(timesMode))
		ret := 0
		if err != nil {
			ret = -1
			fmt.Println(err)
		}
		m := &msg.GSCL_RandomStone{
			Ret:     int32(ret),
			ItemIds: itemIds,
		}
		p.conn.Send(m)
	case global.Message_RootKey_RandomStoneCfg:
		msg, err := p.GetRandomStoneCfg()
		if err != nil {
			fmt.Println(err)
			return
		}
		p.conn.Send(msg)
	case global.Message_RootKey_EquipUpdate:
		item_id := int32(msg.RdWrEndian.Uint32(recvData[4:]))
		isSuccess,pk,newVal,newEquipLv,err := p.EquipUpdate(item_id)
		ret := 1
		if err != nil {
			ret = -1
			fmt.Println(err)
		} else if !isSuccess {
			ret = 0
		}
		m := &msg.GSCL_EquipUpdate{
			Pk:pk,
			NewValue:newVal,
			Ret:int32(ret),
			NewEquipLv:newEquipLv,
		}
		p.conn.Send(m)
	case global.Message_RootKey_ChangeName:
		p.changeName(recvData[4:])
	case global.Message_RootKey_ExpandBag:
		p.expandBag(recvData[4:])
	case global.Message_RootKey_QuickFight:
		p.quickFight(recvData[4:])
	case global.Message_RootKey_CDKey:
		p.cdKeyReward(recvData[4:])
	case global.Message_RootKey_Mail_Reward:
		mailId := int32( msg.RdWrEndian.Uint32(recvData[4:]))
		global.ServerG.GetMailMgr().Reward(p, mailId)
	case global.Message_RootKey_Quick_Challenge:
		challenge_id := int32( msg.RdWrEndian.Uint32(recvData[4:]))
		item_id,val,err := p.QuickChallenge(challenge_id)
		ret := 1
		if err != nil {
			ret = -1
		}
		m := &msg.GSCL_QuickChallenge{
			Ret:int32(ret),
			Challenge_id:challenge_id,
			ItemId:item_id,
			Val:val,
		}
		p.conn.Send(m)
	case global.Message_RootKey_Mine_Buy:
		p.mine_buy(recvData[4:])
	case global.Message_RootKey_Mine_Reward:
		p.mine_reward(recvData[4:])
	case global.Message_RootKey_Mine_Work:
		p.mine_work(recvData[4:])
	case global.Message_RootKey_Login_Reward:
		p.msg_login(recvData[4:])
	case global.Message_RootKey_Sign_Reward:
		p.msg_sign(recvData[4:])
	case global.Message_RootKey_Advance_LevelUp:
		p.msg_advance_levelUp(recvData[4:])
	case global.Message_RootKey_Vip_Reward:
		p.vip_reward(recvData[4:])
	default:
		fmt.Println("传的啥JB玩意儿!!!")
	}
}

func (p *player) OnRet(ret *dbengine.CDBRet) {
	if ret == nil {
		return
	}

	if ret.Err != nil {
		log.Error("player:%d OpType:%d error:%p", p.GetDBId(), ret.OpType, ret.Err)
		return
	}

	//網絡都斷開了
	if p.conn == nil || p.conn.IsClosed() {
		return
	}

	switch ret.OpType {
	case EVENTYPE_DB_RET_GET_EQUIPMENT:
		p.ReadEquipFromDB(ret)
		p.GetPlayerSkill()
	case EVENTYPE_DB_RET_GET_SKILL:
		p.ReadSkillFromDB(ret)
		p.Online(false)
	case EVENTYPE_DB_RET_GET_BACKPACK:
		err := p.ReadBackFromDB(ret)
		if err != nil {
			fmt.Println("不发送用户背包数据")
			return
		}

		p.SendBackPackToClient()

		//自动卖出
		p.sendEquipAutoSellToClient()

		//离线奖励
		p.caculateOfflineReward()
	case EVENTYPE_DB_RET_GET_CDKEY:
		p.ReadCDKeyFromDB(ret)
	}
}

func (p *player) Online(create bool) {
	now := time.Now().UnixNano()
	p.loadDataTime = now

	if p.conn != nil {
		p.bitflags = make(map[int]byte)

		//发送上线消息
		m := &msg.GSCL_CreatePlayer{
			Create:  byte(0),
			Now:int32(time.Now().Unix()),
			Defines: []interface{}{},
		}

		if create {
			m.Create = byte(1)
		}

		m.EquipDBData = p.EquipDBData
		m.SkillDBData = p.SkillDBData
		p.GetPrivateProps(&m.PlayerPrivateProps)

		//todo 一些常量定义
		m.Defines = append(m.Defines, global.FIGHT_ROUND_TIME_SEC) //每回合战斗时间

		p.conn.Send(m)

		p.online = true
		p.lastOnlineTime = int32(now / int64(time.Second))
		if p.lastOfflineTime == 0 {
			p.lastOfflineTime = p.lastOnlineTime
		}

		//拉去背包
		if create {
			p.initBackPackData()

			p.initCDKeyData()

			// 往背包塞数据
			p.RegisterSendItem()
			p.SendBackPackToClient()
		} else {
			p.GetPlayerBackPack()
			p.GetPlayerCDKeys()
		}

		//打印
		logInfo, err := json.Marshal(&struct{
			Id 		int64
			OpenId 	string
			Register bool
			Time int
		}{
			Id: p.dbId,
			OpenId:p.openId,
			Register: create,
			Time: int(now / int64(time.Second)),
		})

		if err != nil {
			log.Error("player::online json.Marshal logError:%s", err)
		} else {
			global.ServerG.GetLog().Info(global.Log_FileName_Login, p.GetDeviceType(), string(logInfo))
		}

		global.ServerG.GetEventRouter().DoEvent(global.Event_Type_PlayerOnline, p, create)

		p.sendSignInfo()
		p.sendLoginInfo()
	}
}

func (p *player) Offline() {

	p.conn = nil
	p.bitflags = nil

	p.online = false

	now := time.Now().UnixNano()
	p.lastOfflineTime = int32(now / int64(time.Second))
	p.loadDataTime = now
	p.OnSave()

	//清理背包
	p.BackPackDBData = nil

	global.ServerG.GetEventRouter().DoEvent(global.Event_Type_PlayerOffline, p, nil)
}

func (p *player) GetLastOnlineTime() int32 {
	return p.lastOnlineTime
}

func (p *player) GetLastOfflineTime() int32 {
	return p.lastOfflineTime
}

func (p *player) GetLoadDataTime() int64 {
	return p.loadDataTime
}

func (p *player) IsOnline() bool {
	return p.online
}

func (p *player) GetSkillData() *global.SkillDBData {
	return p.SkillDBData
}

//一段时间内计算战斗奖励
func (p *player) caculateFightInTimeDuration(durationSec int32) (int32, map[int32]int32, map[int32]int32, error){
	if durationSec <= 0 {
		return 0, nil, nil, errors.New("caculateFightInTimeDuration offlineSec <= 0")
	}

	//todo读取配置，根据会员vip等级做递减
	fightBaseTime := p.GetVipEffectValue(global.Vip_Effect_FightTimeSec)
	if durationSec < fightBaseTime {
		return  0, nil, nil, errors.New("caculateFightInTimeDuration offlineSec < fightBaseTime")
	}

	fightCount := int(durationSec / fightBaseTime)
	ricfg := global.ServerG.GetConfigMgr().GetCfg("Region", p.regionId)
	if ricfg == nil {
		log.Error("Player::caculateFightInTimeDuration getRegionId:%d empty", p.regionId)
		return 0, nil, nil, errors.New("caculateFightInTimeDuration getRegion empty")
	}

	//获取怪物
	rcfg := ricfg.(*global.RegionCfg)
	monArr := strings.Split(rcfg.MonData, "+")

	monNum := rcfg.MonNum
	if monNum > int32(len(monArr)) {
		monNum = int32(len(monArr))
	}

	if monNum == 0 {
		return 0, nil, nil, errors.New("caculateFightInTimeDuration monsterdata empty")
	}

	monIdArr := []int32{}
	for _, monIdStr := range monArr {
		if monId, err := strconv.Atoi(monIdStr); err == nil {
			monIdArr = append(monIdArr, int32(monId))
		}
	}

	randSrc := rand.New(rand.NewSource(time.Now().UnixNano()))

	//总的奖励
	rewardData := &global.RewardData{
		Items: make(map[int32]*global.RewardItem),
	}

	//最多随机20次
	cacCount := 200
	if fightCount < cacCount {
		cacCount = fightCount
	}

	for i := 0; i < cacCount; i++ {
		//生成怪物数据
		copyMonIdArr := monIdArr
		choseMonArr := make([]int32, 0)
		for {
			if len(monIdArr) == 0 || len(choseMonArr) == int(monNum) {
				break
			}

			ri := randSrc.Intn(len(copyMonIdArr))
			choseMonArr = append(choseMonArr, copyMonIdArr[ri])
			copyMonIdArr = append(copyMonIdArr[:ri], copyMonIdArr[ri+1:]...)
		}

		//计算怪物奖励
		for _, monId := range choseMonArr {
			micfg := global.ServerG.GetConfigMgr().GetCfg("Monster", monId)
			if micfg == nil {
				continue
			}

			mcfg := micfg.(*global.MonsterCfg)
			ritems := p.generateFightWinReward(mcfg.DropData, randSrc)
			if ritems != nil {
				for _, item := range ritems.Items {
					if _, ok := rewardData.Items[item.Id]; ok {
						rewardData.Items[item.Id].Num += item.Num
					} else {
						rewardData.Items[item.Id] = item
					}
				}
			}
		}
	}

	per := float32(fightCount) / float32(cacCount)
	mItems := make(map[int32]int32)
	for _, sItem := range rewardData.Items {
		mItems[sItem.Id] = sItem.Num
		if per > 1 {
			mItems[sItem.Id] = int32(float32(mItems[sItem.Id]) * per)
		}

		sItem.Num = mItems[sItem.Id]
	}

	addItems, delInsts, sellItems, _ := p.AddItems(rewardData, false, true)
	mbp := &msg.GSCL_PlayerUpdateBackPack{
		AddItems: addItems,
		DelItems: delInsts,
	}
	p.conn.Send(mbp)

	mSellItems := make(map[int32]int32)
	if sellItems != nil {
		for _, sItem := range sellItems {
			if _, ok := mSellItems[sItem.CfgId]; ok {
				mSellItems[sItem.CfgId] += sItem.Num
			} else {
				mSellItems[sItem.CfgId] = sItem.Num
			}
		}
	}

	return int32(fightCount), mItems, mSellItems, nil
}

//离线奖励
func (p *player) caculateOfflineReward() {
	if !p.IsOnline() {
		return
	}

	offlineSec := p.lastOnlineTime - p.lastOfflineTime

	if offlineSec > int32(global.ONE_DAY_SEC) {
		offlineSec = int32(global.ONE_DAY_SEC)
	}

	fc, addItems, sellItems, err := p.caculateFightInTimeDuration(offlineSec)
	if err != nil {
		return
	}

	//发送离线奖励
	mor := &msg.GSCL_PlayerOfflineReward{
		OfflineSec: offlineSec,
		FightCount: fc,
		Items:      addItems,
		SellItems:  sellItems,
	}

	p.conn.Send(mor)

	return
}

func (p *player) SetLoginType(loginType int)  {
	p.loginType = loginType
}

func (p *player) SetOpenId(openId string)  {
	p.openId = openId
}

func (p *player) SetDeviceType(deviceType string)  {
	p.deviceType = deviceType
}

func (p *player) GetLoginType() int  {
	return p.loginType
}

func (p *player) GetOpenId() string  {
	return p.openId
}

func (p *player) GetDeviceType() string {
	return p.deviceType
}

func (p *player) changeName(recvData []byte) {
	if len(recvData) < 2 {
		return
	}

	nameLen := conf.RdWrEndian.Uint16(recvData)
	if nameLen <= 0 || len(recvData) != 2 + int(nameLen) {
		return
	}

	_, _, _, err := p.AddItem(global.CHANGE_NAME_ITEM_ID, -1, true, false)
	if err != nil {
		return
	}

	name := recvData[2:]
	name = append(name, 0)
	copy(p.name[:global.NAME_MAX_LEN], name)
}

func (p *player) Kick() {
	log.Debug("player:%d kicked", p.dbId)
	p.conn.Kick()
}

func (p *player) quickFight(recvData []byte) {
	//判断次数是否已经用完
	qfCount, _ := p.GetProp(global.Player_Prop_QuickFightCount)
	if qfCount >= p.GetVipEffectValue(global.Vip_Effect_QuickFightCount) {
		return
	}

	//计算消费钻石
	needDiamond := qfCount * 20
	if needDiamond > global.QUICK_FIGHT_MAX_DIAMOND {
		needDiamond = global.QUICK_FIGHT_MAX_DIAMOND
	}

	playerDiamond, _ := p.GetProp(global.Player_Prop_Diamond)
	if playerDiamond < needDiamond {
		return
	}

	fc, addItems, sellItems, err := p.caculateFightInTimeDuration(global.QUICK_FIGHT_DURATION_SEC)
	if err != nil {
		return
	}

	p.SetProp(global.Player_Prop_Diamond, -needDiamond, true)
	p.SetProp(global.Player_Prop_QuickFightCount, 1, true)

	//发送离线奖励
	m := &msg.GSCL_QuickFight{
		FightCount: fc,
		Items:      addItems,
		SellItems:  sellItems,
	}

	p.conn.Send(m)

	return
}

//vip值
func (p *player) GetVipEffectValue(etype int32) int32 {
	vipLevel, _ := p.GetProp(global.Player_Prop_VipLevel)

	cfg := global.ServerG.GetConfigMgr().GetCfg("Vip", vipLevel)
	if cfg == nil {
		log.Error("player::GetVipEffectValue playerid:%d getvipcfg:%d empty", p.dbId, vipLevel)
		return 0
	}

	icfg := cfg.(*global.VipCfg)
	switch etype {
	case global.Vip_Effect_AddExpDrop:
		return icfg.AddExpDrop
	case global.Vip_Effect_AddMoneyDrop:
		return icfg.AddMoneyDrop
	case global.Vip_Effect_AddInstanceCount:
		return icfg.AddInstanceCount
	case global.Vip_Effect_DailyGift:
		return icfg.DailyGiftId
	case global.Vip_Effect_FightTimeSec:
		return icfg.FightTimeSec
	case global.Vip_Effect_FightWaitSec:
		return icfg.FightWaitSec
	case global.Vip_Effect_QuickFightCount:
		return icfg.QuickFightCount
	case global.Vip_Effect_MineRobTime:
		return icfg.MineRobTime
	}

	log.Error("player::GetVipEffectValue playerid:%d effectType:%d error", p.dbId, etype)
	return 0
}


//充值
func (p *player) Recharge(rechargeNum int32) {
	if rechargeNum <= 0 {
		return
	}

	//充值
	p.SetProp(global.Player_Prop_RechargeNum, rechargeNum, true)

	//vip经验
	p.SetProp(global.Player_Prop_VipExp, rechargeNum, true)
}

func fightValTable(pid int32) int32 {
	switch pid {
	case global.Creature_Prop_Two_Attack:return 10
	case global.Creature_Prop_Two_Crit:return 1
	case global.Creature_Prop_Two_AttackSpeed:return 1
	case global.Creature_Prop_Two_Miss:return 1
	case global.Creature_Prop_Two_Blood:return 1
	case global.Creature_Prop_Two_Get:return 1
	case global.Creature_Prop_Two_Defence:return 10
	case global.Creature_Prop_Two_Magic:return 1
	default:
		return 0
	}
}

func (p *player) CaculateFightVal()  {
	fightVal, _ := p.GetProp(global.Player_Prop_FightVal)

	var newFightVal int32
	for i:=global.Creature_Prop_Two_Attack; i <= global.Creature_Prop_Two_Tenacity;i++ {
		v, _ := p.GetProp(i)
		newFightVal += v * fightValTable(int32(i))
	}

	if newFightVal != fightVal {
		p.SetProp(global.Player_Prop_FightVal, newFightVal, false)
		global.ServerG.GetSkeleton().Go(func(){p.playerInfo.SetToRank(p.dbId,newFightVal)}, func(){})
	}
}

func (p *player) vip_reward(recvData []byte) {
	vipRewardTime, _ := p.GetProp(global.Player_Prop_VipReward_Time)

	now := int64(time.Now().Unix())
	if utils.CheckIsSameDayBySec(int64(vipRewardTime), now, 0) {
		return
	}

	giftId := p.GetVipEffectValue(global.Vip_Effect_DailyGift)
	icfg := p.getItemCfgByCfgId(giftId)
	if icfg == nil {
		return
	}

	rewardItem := getRewardFromStr(icfg.(*global.ItemCfg).Data)
	if p.IsBagFullMulti(rewardItem, false) {
		return
	}

	p.SetProp(global.Player_Prop_VipReward_Time, int32(now), false)

	p.AddItems(rewardItem, true, false)
}
