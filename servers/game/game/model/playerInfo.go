package model

import (
	"github.com/garyburd/redigo/redis"
	"xianxia/common/dbengine"
	"fmt"
	"time"
	"xianxia/servers/game/utils"
	"xianxia/servers/game/game/global/db"
	"encoding/json"
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/game/global"
	"errors"
)


type PlayerInfo struct{
	DB *dbengine.CDBEngine
}

func resloveTimesKey(uid int64) string{
	return fmt.Sprintf("userResloves:%d", uid)
}

func playerEquipCfgKey(uid int64) string {
	return fmt.Sprintf("equipCreateCfgId:%d", uid)
}

func (u *PlayerInfo) SetPlayerEquipCreateCfgId(uid int64,equipCfgId int32) (ret string, err error) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	remainTime := utils.GetTodayEndUnix() - time.Now().Unix()
	return redis.String(conn.Do("SET", playerEquipCfgKey(uid), equipCfgId, "EX", remainTime))
}

func (u *PlayerInfo) GetEquipCreateCfgIdCache(uid int64) (id int32, err error) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	ret, err := redis.Int(conn.Do("GET", playerEquipCfgKey(uid)))
	if err !=nil {
		return
	}
	return int32(ret),nil

}

// 获取用户已锻造次数
func (u *PlayerInfo) GetPlayerResloveTimes(uid int64) (id int, err error) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	return redis.Int(conn.Do("GET", resloveTimesKey(uid)))
}

// 增加锻造次数
func (u *PlayerInfo) IncrByPlayerResloveTimes(uid int64) (times int, err error){
	conn := u.DB.Redis.Get()
	defer conn.Close()
	redis.Int(conn.Do("EXPIRE", resloveTimesKey(uid), utils.GetTodayEndUnix() - time.Now().Unix()))
	return redis.Int(conn.Do("INCR", resloveTimesKey(uid)))
}

func getCDKeyKey(key string) string {
	return fmt.Sprintf("cdkey:%s", key)
}

func (u *PlayerInfo) GetCDKey(key string) (*db.DB_CDKey_Info, error) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	ret, err := redis.String(conn.Do("GET", getCDKeyKey(key)))
	if err !=nil {
		return nil, err
	}

	data := &db.DB_CDKey_Info{}
	err = json.Unmarshal([]byte(ret), data)
	return data, err
}

func (u *PlayerInfo) RemoveCDKey(key string) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	_, err := conn.Do("del", getCDKeyKey(key))
	if err !=nil {
		log.Error("PlayerModel RemoveCDKey:%s error:", key, err)
	}
}


func getFightValRankKey() string{
	return fmt.Sprintf("fightRank")
}
func (u *PlayerInfo) SetToRank(uid int64, fightVal int32) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	_, err := conn.Do("ZADD", getFightValRankKey(), fightVal, uid)
	if err != nil {
		fmt.Println("插入集合失败!!!")
	}
	return
}

func getPlayerChallengeLogKey(uid int64) string {
	return fmt.Sprintf("challengeMax:%d", uid)
}

func getChallengeRankKey(challenge_id int32) string {
	return fmt.Sprintf("challengeRank:%d", challenge_id)
}

func (u *PlayerInfo) GetPlayerChallengeInfo(uid int64,ch_id int32) (cnt int,err error){
	conn := u.DB.Redis.Get()
	defer conn.Close()
	cnt, err = redis.Int(conn.Do("ZSCORE", getChallengeRankKey(ch_id), uid))
	if err != nil  && err != redis.ErrNil{
		return
	}

	// 查不到数据,说明今天没挑战
	if err == redis.ErrNil {
		//fmt.Println("查不到数据!,ZSCORE ", getChallengeRankKey(ch_id), " ",)
		err = nil
		cnt = 0
	}
	return
}

func (u *PlayerInfo) GetPlayerChallengeMaxCnt(uid int64, challenge_id int32) int {
	conn := u.DB.Redis.Get()
	defer conn.Close()

	old_cnt,err := redis.Int(conn.Do("HGET", getPlayerChallengeLogKey(uid), challenge_id))
	if err != nil {
		return 0
	}

	return old_cnt
}

func (u *PlayerInfo) SetPlayerChallenge(uid int64,challenge_id,challenge_cnt int32) (err error){
	conn := u.DB.Redis.Get()
	defer conn.Close()

	// 记录最大层数
	old_cnt,err := redis.Int(conn.Do("HGET", getPlayerChallengeLogKey(uid), challenge_id))
	if err != nil && err != redis.ErrNil {
		return
	}

	if err == redis.ErrNil {
		old_cnt = 0
	}

	// 新记录大于旧记录
	if  challenge_cnt >int32(old_cnt) {
		conn.Do("HSET", getPlayerChallengeLogKey(uid), challenge_id, challenge_cnt)
	}

	// 写排行榜
	_, err = conn.Do("ZADD", getChallengeRankKey(challenge_id), challenge_cnt, uid)
	fmt.Println("写入排行榜,uid:", uid, "cid:",challenge_id, "层数:",challenge_cnt)
	// 12d点过期
	conn.Do("EXPIRE", getChallengeRankKey(challenge_id), utils.GetTodayEndUnix())
	return
}
func (u *PlayerInfo) SetPlayerChallengeRank(uid int64,challenge_id int32,challenge_cnt int) (err error) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	_, err = conn.Do("ZADD", getChallengeRankKey(challenge_id), challenge_cnt, uid)
	return
}

//矿场数据
func getMineKey(uid int64) string{
	return fmt.Sprintf("mine:%d", uid)
}

func (u *PlayerInfo) GetPlayerMineInfo(uid int64) (*global.PlayerMineInfo, error) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	data := &global.PlayerMineInfo{
		Mines:make([]*global.PlayerMine, 0),
		LastUpdateTime:int32(time.Now().Unix()),
	}

	ret, err := redis.String(conn.Do("GET", getMineKey(uid)))
	if err != nil {
		if err == redis.ErrNil {
			return data, nil
		} else {
			return nil, err
		}
	}

	err = json.Unmarshal([]byte(ret), data)
	return data, err
}

func (u *PlayerInfo) SetPlayerMineInfo(uid int64,mineInfo *global.PlayerMineInfo) (error) {
	if mineInfo == nil {
		return errors.New("SetPlayerMineInfo::PlayerMineInfo nil")
	}

	conn := u.DB.Redis.Get()
	defer conn.Close()
	j, err := json.Marshal(mineInfo)
	_, err = conn.Do("SET", getMineKey(uid), j)
	return err
}

func genMineRobTimeKey(uid int64, robUid int64, cfgId int32) string {
	return fmt.Sprintf("mine_rob:%d_%d_%d", robUid, cfgId, uid)
}
func (u *PlayerInfo) GetMineRobTimeKey(uid int64, robUid int64, cfgId int32) bool {
	conn := u.DB.Redis.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("Get", genMineRobTimeKey(uid, robUid, cfgId)))
	if err == redis.ErrNil {
		return false
	}

	return true
}

func (u *PlayerInfo) SetMineRobTimeKey(uid int64, robUid int64, cfgId int32, sec int32) {
	conn := u.DB.Redis.Get()
	defer conn.Close()

	expireTime := int32(time.Now().Unix()) + sec
	conn.Do("SETEX", genMineRobTimeKey(uid, robUid, cfgId), sec, expireTime)
}


//签到数据
func getSignKey(uid int64) string{
	return fmt.Sprintf("sign:%d", uid)
}

func (u *PlayerInfo) GetPlayerSignInfo(uid int64) (*global.PlayerSignInfo, error) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	data := &global.PlayerSignInfo{
		RewardStates: make([]int32, 0),
	}

	ret, err := redis.String(conn.Do("GET", getSignKey(uid)))
	if err != nil {
		if err == redis.ErrNil {
			return data, nil
		} else {
			return nil, err
		}
	}

	err = json.Unmarshal([]byte(ret), data)
	return data, err
}


func (u *PlayerInfo) SetPlayerSignInfo(uid int64,signInfo *global.PlayerSignInfo) (error) {
	if signInfo == nil {
		return errors.New("SetPlayerSignInfo::PlayerSignInfo nil")
	}

	conn := u.DB.Redis.Get()
	defer conn.Close()
	j, err := json.Marshal(signInfo)
	_, err = conn.Do("SET", getSignKey(uid), j)
	return err
}


//登陆数据
func getLoginActKey(uid int64) string{
	return fmt.Sprintf("login_act:%d", uid)
}

func (u *PlayerInfo) GetPlayerLoginActInfo(uid int64) (*global.PlayerLoginInfo, error) {
	conn := u.DB.Redis.Get()
	defer conn.Close()
	data := &global.PlayerLoginInfo{
		RewardStates: make([]int32, 0),
	}

	ret, err := redis.String(conn.Do("GET", getLoginActKey(uid)))
	if err != nil {
		if err == redis.ErrNil {
			return data, nil
		} else {
			return nil, err
		}
	}

	err = json.Unmarshal([]byte(ret), data)
	return data, err
}


func (u *PlayerInfo) SetPlayerLoginActInfo(uid int64,loginInfo *global.PlayerLoginInfo) (error) {
	if loginInfo == nil {
		return errors.New("SetPlayerLoginActInfo::PlayerLoginInfo nil")
	}

	conn := u.DB.Redis.Get()
	defer conn.Close()
	j, err := json.Marshal(loginInfo)
	_, err = conn.Do("SET", getLoginActKey(uid), j)
	return err
}