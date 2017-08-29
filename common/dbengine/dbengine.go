package dbengine

import (
	"github.com/garyburd/redigo/redis"
	"github.com/name5566/leaf/module"
	"time"
)

type CDBRet struct {
	OpType  int
	Sink    IDBSink
	DBId 	int64
	Err     error
	Content interface{}
}

type IDBSink interface {
	OnRet(ret *CDBRet)
}

type CDBEngine struct {
	Skeleton *module.Skeleton
	Redis    *redis.Pool
}

var DBEngine *CDBEngine

func init() {
	DBEngine = &CDBEngine{}
}

func (db *CDBEngine) Start(skeleton *module.Skeleton, dbAddr string) {
	db.Redis = &redis.Pool{
		MaxIdle:     100,
		MaxActive:   100,
		IdleTimeout: 60 * 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", dbAddr)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	db.Skeleton = skeleton
}

func (db *CDBEngine) Close() {
	if db.Redis != nil {
		db.Redis.Close()
	}
}

func (db *CDBEngine) Request(sink IDBSink, opType int, dbId int64, cmd string, args ...interface{}) {
	dbRet := &CDBRet{}
	f := func() {
		conn := db.Redis.Get()
		defer conn.Close()
		data, err := conn.Do(cmd, args...)
		dbRet.OpType = opType
		dbRet.DBId = dbId
		dbRet.Err = err
		dbRet.Content = data
		dbRet.Sink = sink
	}

	cb := func() {
		if sink != nil {
			sink.OnRet(dbRet)
		}
	}

	db.Skeleton.Go(f, cb)
}

func (db *CDBEngine) RedisExec(cmd string, args ...interface{} )(reply interface{}, err error) {
	conn := db.Redis.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

func (db *CDBEngine) GetUniqueID() (id int, err error) {
	conn := db.Redis.Get()
	defer conn.Close()
	return redis.Int(conn.Do("INCRBY", "uniqueId", 1))
}
