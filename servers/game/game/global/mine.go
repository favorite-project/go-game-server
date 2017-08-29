package global

const Mine_Msg_Max_Len = 10

const (
	Mine_Msg_Type_Rob = 1+iota //抢劫消息
	Mine_Msg_Type_Work			//打工消息
)

type PlayerMineMsg struct {
	Type int 		`json:"type"`
	Num int32  		`json:"num"`
	PlayerId int64  `json:"playerId"`
	WorkId int32	`json:"workId"`
	Time int32	`json:"time"`
}

type PlayerMine struct {
	Id 			int32		`json:"id"`//id
	CfgId 		int32		`json:"cfgId"`//配置id
	LastCalcTime int32		`json:"lastCalcTime"`//上次计算收益时间
	Works map[int32]int32  `json:"works"`//工作次数
	RobNum int32 			`json:"robNum"`//被抢数量
	Msgs []*PlayerMineMsg	`json:"msgs"`
}

type PlayerMineInfo struct {
	Mines []*PlayerMine   `json:"mines"`
	LastUpdateTime int32 `json:"lastUpdateTime"`
}

const Mine_Pool_Key  = "mine_pool_list"

type MinePoolNode struct {
	PlayerId int64 `json:"playerId"`//
	PlayerName string `json:"playerName"`
	CfgId int32  `json:"cfgId"`//
}
