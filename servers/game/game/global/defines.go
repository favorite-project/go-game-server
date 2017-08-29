package global

const NAME_MAX_LEN = 256
const PIC_MAX_LEN = 256

//服务器启动状态
const (
	Server_State_Closing  = 1 + iota //正在关闭
	Server_State_Closed              //关闭成功
	Server_State_Starting            //正在启动中
	Server_State_Started             //启动成功
)

const (
	ONE_DAY_SEC = 24 * 60 * 60
)

const CHANGE_NAME_ITEM_ID = 60003

const HEARTBEAT_SEC = 10 //心跳包秒数

const SYSTEM_DBID = int64(0)