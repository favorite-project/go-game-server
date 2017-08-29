package dbengine

/*
import (
	"github.com/name5566/leaf/module"
	"github.com/garyburd/redigo/redis"
	"errors"
)

type DBSet struct {
	Addr string
	MaxIdle int
	MaxActive int
}

type CDBEngineMgr struct {
	DBS map[string]*CDBEngine
}

func (mgr *CDBEngineMgr) Start(skeleton *module.Skeleton, dbAddrs map[string]*DBSet) bool {
	if skeleton == nil || dbAddrs == nil || len(dbAddrs) == 0 {
		return false
	}

	if mgr.DBS == nil {
		mgr.DBS = make(map[string]*CDBEngine)
	}

	for k,v := range dbAddrs {
		db := &CDBEngine{}
		if !db.Start(skeleton, v.Addr, v.MaxIdle, v.MaxActive) {
			return false
		}

		mgr.DBS[k] = db
	}

	return true
}

func(mgr *CDBEngineMgr) Close() bool {
	if mgr.DBS != nil {
		for _, db := range mgr.DBS {
			if !db.Close() {
				return false
			}
		}

		mgr.DBS = nil
	}

	return true
}

func (mgr *CDBEngineMgr) Request(sink IDBSink, db string, opType int, dbId int64, cmd string, args ...interface{}) {
	if mgr.DBS == nil {
		return
	}

	if _, ok := mgr.DBS[db]; !ok {
		return
	}

	mgr.DBS[db].Request(sink, opType, dbId, cmd, args)
}

func (mgr *CDBEngineMgr) GetDBEngine(db string) *CDBEngine {
	if mgr.DBS == nil {
		return nil
	}

	if _, ok := mgr.DBS[db]; !ok {
		return nil
	}

	return mgr.DBS[db]
}

func (mgr *CDBEngineMgr) GetRedis(db string) *redis.Pool {
	if mgr.DBS == nil {
		return nil
	}

	if _, ok := mgr.DBS[db]; !ok {
		return nil
	}

	return mgr.DBS[db].Redis
}

//要记得conn.Close，不然不会释放端口
func (mgr *CDBEngineMgr) GetRedisConn(db string) redis.Conn {
	if mgr.DBS == nil {
		return nil
	}

	if _, ok := mgr.DBS[db]; !ok {
		return nil
	}

	return mgr.DBS[db].Redis.Get()
}

func (mgr *CDBEngineMgr) RedisExec(db string, cmd string, args ...interface{}) (reply interface{}, err error) {
	conn := mgr.GetRedisConn(db)
	if conn == nil {
		return nil, errors.New("get redis.conn nil")
	}

	defer conn.Close()
	return conn.Do(cmd, args...)
}
*/