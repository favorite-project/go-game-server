package http

import (
	"net/http"
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/conf"
	"xianxia/servers/game/game/global"
	"fmt"
	"github.com/name5566/leaf/module"
	"strconv"
	"encoding/json"
	"strings"
)

func init() {

}

var Skeleton  *module.Skeleton

func Start(skeleton *module.Skeleton) {
	Skeleton = skeleton
	registHandlesAndGsLogics()
	go func() {
		err := http.ListenAndServe(conf.Server.HttpServer, nil)
		if err != nil {
			log.Fatal("http::init ListenAndServe: ", err)
		}
	}()
}

func registHandlesAndGsLogics() {
	http.HandleFunc("/hello", HelloServer)

	//http router
	http.HandleFunc("/onlineCount", OnlineCount)

	// http service
	http.HandleFunc("/playerInfo", GetPlayerInfo)


	//chanrpc 回调函数会在gs主逻辑中执行，不存在并发问题
	Skeleton.RegisterChanRPC("GS_OnlineCount", GSHttpOnlineCount)

	//发邮件
	http.HandleFunc("/sendMail", SendMail)
	Skeleton.RegisterChanRPC("GS_SendMail", GsSendMail)

	//玩家添加道具
	http.HandleFunc("/addItem", AddItem)
	Skeleton.RegisterChanRPC("GS_AddItem", GsAddItem)
}

func rpcGsLogic(rpcName string, w http.ResponseWriter, req *http.Request, args ...interface{}) {
	recvChan := make(chan []byte)
	newArgs := []interface{}{w, req, recvChan}
	newArgs = append(newArgs, args...)
	Skeleton.ChanRPCServer.Go(rpcName, newArgs...)
	recvData := <- recvChan
	close(recvChan)
	w.Write(recvData)
}

func gsLogicEnd(msg []byte, args ...interface{}) {
	closeChan := args[2].(chan []byte)
	closeChan <- msg
}

// Test: hello world, the web server
func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello back"))
}

//webserver 在线人数
func OnlineCount(w http.ResponseWriter, req *http.Request) {
	rpcGsLogic("GS_OnlineCount", w, req)
}

func GSHttpOnlineCount(args []interface{}) {
	onlineArr := global.ServerG.GetObjectMgr().GetOnlinePlayer()
	data := fmt.Sprintf("在线人数：%d", len(onlineArr))

	gsLogicEnd([]byte(data), args...)
}

func GetPlayerInfo(w http.ResponseWriter, req *http.Request) {
	u := req.FormValue("uid")
	uid,err := strconv.Atoi(u)
	if err != nil {
		fmt.Println(err)
		return
	}
	player := global.ServerG.GetObjectMgr().GetPlayer(int64(uid))
	fmt.Println("ppp:" , player)
}


//webserver 发送邮件
func SendMail(w http.ResponseWriter, req *http.Request) {
	//在rpc调用gs逻辑之前可以做参数检查
	rpcGsLogic("GS_SendMail", w, req)
}

func GsSendMail(args []interface{}) {

	var ret struct {
		Ret int `json:"ret"`
		Msg string `json:"msg"`
	}

	req := args[1].(*http.Request)
	req.ParseForm()

	sysId, err := strconv.Atoi(req.Form.Get("sysId"))
	if err != nil {
		ret.Msg = err.Error()
		ret.Ret = -1
	}

	var mType int
	if ret.Ret == 0 {
		mType, err = strconv.Atoi(req.Form.Get("type"))
		if err != nil {
			ret.Msg = err.Error()
			ret.Ret = -2
		}
	}

	var title string
	if ret.Ret == 0 {
		title = req.Form.Get("title")
		if len(title) == 0 {
			ret.Msg = "标题不能为空"
			ret.Ret = -3
		}
	}

	var content string
	if ret.Ret == 0 {
		content = req.Form.Get("content")
		if len(content) == 0 {
			ret.Msg = "内容不能为空"
			ret.Ret = -4
		}
	}

	items := make(map[int32]int32)
	if ret.Ret == 0 {
		itemsStr := req.Form.Get("items")
		if len(itemsStr) > 0 {
			itemArr := strings.Split(itemsStr, ";")
			for _, itemStr := range itemArr {
				itemData := strings.Split(itemStr, "+")
				if len(itemData) != 2 {
					ret.Msg = "附件格式不对"
					ret.Ret = -5
					break
				}

				id, err := strconv.Atoi(itemData[0])
				if err != nil {
					ret.Msg = "附件格式不对"
					ret.Ret = -5
					break
				}

				num, err := strconv.Atoi(itemData[1])
				if err != nil{
					ret.Msg = "附件格式不对"
					ret.Ret = -5
					break
				}

				items[int32(id)] = int32(num)
			}

		} else {
			items = nil
		}
	}

	var receiverId int
	if ret.Ret == 0 {
		receiverId, err = strconv.Atoi(req.Form.Get("receiver"))
		if err != nil {
			ret.Msg = err.Error()
			ret.Ret = -6
		}
	}

	var expireTime int
	if ret.Ret == 0 {
		expireTime, err = strconv.Atoi(req.Form.Get("expireTime"))
		if err != nil {
			ret.Msg = err.Error()
			ret.Ret = -7
		}
	}

	if ret.Ret == 0 {
		mailMgr := global.ServerG.GetMailMgr()

		if !mailMgr.SendMail(int32(sysId), byte(mType), title, content, items, int64(receiverId), int32(expireTime)) {
			ret.Msg = "发送邮件失败"
			ret.Ret = -8
		}
	}

	j, _ := json.Marshal(ret)

	gsLogicEnd(j, args...)
}

func AddItem(w http.ResponseWriter, req *http.Request) {
	//在rpc调用gs逻辑之前可以做参数检查
	rpcGsLogic("GS_AddItem", w, req)
}

func GsAddItem(args []interface{}) {

	var ret struct {
		Ret int `json:"ret"`
		Msg string `json:"msg"`
	}

	req := args[1].(*http.Request)
	req.ParseForm()

	var Uid int64
	if ret.Ret == 0 {
		uid, err := strconv.Atoi(req.Form.Get("uid"))
		if err != nil {
			ret.Msg = "玩家id格式错误"
			ret.Ret = -1
		}

		Uid = int64(uid)
	}

	var err error
	var propId int
	if ret.Ret == 0 {
		propId, err = strconv.Atoi(req.Form.Get("propId"))
		if err != nil {
			ret.Msg = "属性id格式错误"
			ret.Ret = -2
		}
	}

	var Num int32
	if ret.Ret == 0 {
		num, err := strconv.Atoi(req.Form.Get("num"))
		if err != nil {
			ret.Msg = "属性数量格式错误"
			ret.Ret = -3
		}

		Num = int32(num)
	}

	if ret.Ret == 0 {
		if  propId != global.Player_Prop_Level &&
			propId != global.Player_Prop_Money &&
			propId != global.Player_Prop_Diamond {
			ret.Msg = "只能修改等级 金币和钻石"
			ret.Ret = -4
		}
	}

	if ret.Ret == 0 {
		player := global.ServerG.GetObjectMgr().GetPlayer(Uid)
		if player != nil {
			player.SetProp(propId, Num, true)
		} else {
			var key string
			switch(propId) {
			case global.Player_Prop_Level:
				key = "level"
			case global.Player_Prop_Money:
				key = "money"
			case global.Player_Prop_Diamond:
				key = "diamond"
			}

			conn := global.ServerG.GetDBEngine().Redis.Get()
			if conn == nil {
				ret.Msg = "操作redis错误"
				ret.Ret = -5
			} else {
				_, err := conn.Do("hincrby", fmt.Sprintf("player:%d", Uid), key, Num)
				if err != nil {
					ret.Msg = "操作redis错误"
					ret.Ret = -5
				}
			}
		}
	}

	j, _ := json.Marshal(ret)
	gsLogicEnd(j, args...)
}
