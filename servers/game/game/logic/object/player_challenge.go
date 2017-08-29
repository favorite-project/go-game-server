package object

import (
	"xianxia/servers/game/game/global"
	"github.com/name5566/leaf/log"
	"fmt"
	"strings"
	"strconv"
	"xianxia/servers/game/game/errorx"
)

// 获取用户挑战信息
func (p *player) GetPlayerChallengeInfo(challenge_id int32) (player_chanllenge_cnt int32,err error) {
	cnt,err := p.playerInfo.GetPlayerChallengeInfo(p.dbId,challenge_id)
	if err != nil {
		fmt.Println(err)
		return
	}
	if cnt > 0 {
		player_chanllenge_cnt = int32(cnt)
	} else {
		player_chanllenge_cnt = 0
	}
	fmt.Println("挑战ID", challenge_id, "用户当前挑战(所在)层数:", player_chanllenge_cnt)
	return
}

const Challenge_Mon_Num = 3

// 根据用户挑战层数,获取对应怪物, 怪物直接默认出三只
func (p *player) GetChallengeMonsters(challenge_id int32,player_chanllenge_cnt int32) ([]global.Monster, bool) {
	monster_csv := global.ServerG.GetConfigMgr().GetCsv("ChallengeMonster")
	if monster_csv  == nil {
		log.Error("monster_csv GetCsv Config nil")
		return nil,false
	}

	//monsterCfg := []*global.ChallengeMonster{}
	monObjArr := []global.Monster{}
	for i := 0; i < monster_csv.NumRecord(); i++ {
		epicfg := monster_csv.Record(i)
		if epicfg == nil {
			continue
		}

		epCfg := epicfg.(*global.ChallengeMonster)

		if epCfg.Cid == challenge_id && epCfg.Lid == player_chanllenge_cnt {
			mcfg := global.ServerG.GetConfigMgr().GetCfg("Monster", epCfg.Mid)
			if mcfg == nil {
				log.Error("Player::GetChallengeMonsters monId:%d empty", epCfg.Mid)
				return nil, false
			}

			for i := 0; i < Challenge_Mon_Num;i++ {
				monObjArr = append(monObjArr,ObjectMgr.CreateMonster(mcfg.(*global.MonsterCfg)))
			}

			break
		}
	}

	if len(monObjArr) == 0 {
		return nil, false
	}

	return monObjArr,false
}

func (p *player) GetChallengeMonCfg(challenge_id int32,player_chanllenge_cnt int32)  *global.ChallengeMonster {
	monster_csv := global.ServerG.GetConfigMgr().GetCsv("ChallengeMonster")
	if monster_csv  == nil {
		log.Error("GetChallengeRewardDropData monster_csv GetCsv Config nil")
		return nil
	}

	for i := 0; i < monster_csv.NumRecord(); i++ {
		epicfg := monster_csv.Record(i)
		if epicfg == nil {
			continue
		}

		epCfg := epicfg.(*global.ChallengeMonster)

		if epCfg.Cid == challenge_id && epCfg.Lid == player_chanllenge_cnt {
			return epCfg
		}
	}

	return nil
}

// 挑战战斗结束接口
func (p *player) ChallengeEnd(uid int64,challenge_id,challenge_cnt int32) {
	err := p.playerInfo.SetPlayerChallenge(uid,challenge_id,challenge_cnt)
	fmt.Println("挑战结束:",err, "层数:",challenge_cnt)
}

// 扫荡
func (p *player) QuickChallenge(challenge_id int32) (itemId int32,val int32,err error){
	maxCnt := p.playerInfo.GetPlayerChallengeMaxCnt(p.dbId,challenge_id)
	if maxCnt <= 1 {
		fmt.Println("该挑战你的历史挑战层数是:", maxCnt, "扫荡个JB")
		err = errorx.WRONG_PARAMETER
		return
	}

	// 判断当前层数是否在最大层:
	nowCnt,err := p.playerInfo.GetPlayerChallengeInfo(p.dbId,challenge_id)
	if err != nil {
		return
	}

	if nowCnt == maxCnt {
		fmt.Println("最大层了！！,当前所在:", nowCnt, "最大层数:", maxCnt)
		err = errorx.WRONG_PARAMETER
		return
	}

	cfg := global.ServerG.GetConfigMgr().GetCfg("Challenge", challenge_id)
	if cfg == nil {
		fmt.Println("找不到挑战配置:" , challenge_id)
		err = errorx.CSV_ROW_NOT_FOUND
		return
	}

	config, _ := cfg.(*global.Challenge)
	reArr := strings.Split(config.QuickReward, "+")
	itemCfgId,_:= strconv.Atoi(reArr[0])
	itemId = int32(itemCfgId)
	n,_ := strconv.Atoi(reArr[1])
	num := n * maxCnt
	val = int32(num)
	p.AddItem(itemId,val,true,false)
	p.playerInfo.SetPlayerChallengeRank(p.dbId,challenge_id,maxCnt)
	return
}