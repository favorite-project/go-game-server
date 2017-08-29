package log

import (
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
	llog "github.com/name5566/leaf/log"
	"encoding/json"
)

const Redis_Log_Key = "gs_redis_log_key"

type LogLine struct {
	Device string
	Data string
	FileName string
}

type LogAllInfo struct {
	LogLine *LogLine
	ServerId int
	Platform string
	Time string
}

type Log struct {
	WriteChanNum 		int
	WriteGoRouterNum	int
	Platform string
	ServerId int
	writeChan chan *LogLine
	closeChan chan bool
	redis   			*redis.Pool
	wg               sync.WaitGroup
}

func(log *Log) Start(dbAddr string) {
	log.redis = &redis.Pool{
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

	log.writeChan = make(chan *LogLine, log.WriteChanNum)
	log.closeChan = make(chan bool)

	//开协程往redis里写
	for i:= 0; i < log.WriteGoRouterNum;i++ {
		log.wg.Add(1)
		go log.write()
	}
}

func(log *Log) Close() {
	for i :=0;i< log.WriteGoRouterNum;i++ {
		log.closeChan <- true
	}
	log.wg.Wait()

	if log.redis != nil {
		log.redis.Close()
	}
}

func (log *Log) genLogLine(line *LogLine) (string,error) {
	v, err := json.Marshal(&LogAllInfo{
		LogLine:line,
		ServerId:log.ServerId,
		Platform:log.Platform,
		Time:time.Now().Format("2006-01-02 15:04:05"),
	})

	return string(v), err
}

func (log *Log) write() {
	defer log.wg.Done()
	for {
		select {
			case data := <-log.writeChan:
				rcon := log.redis.Get()
				if rcon == nil {
					llog.Error("Log::write log.redis.Get nil data:", data)
				} else {
					lineStr, err := log.genLogLine(data)
					if err != nil {
						llog.Error("Log::write log.genLogLin err:%s, data:", err, data)
					} else {
						rcon.Do("lpush", Redis_Log_Key, lineStr)
					}
				}
			case <- log.closeChan:
				return
		}
	}
}

func(log *Log) Info(fileName string, device string, data string) {
	if len(log.writeChan) == log.WriteChanNum {
		llog.Error("Log::Info writeChan Full, data:%s", data)
		return
	}

	item := &LogLine {
		Device:device,
		Data:data,
		FileName: fileName,
	}

	log.writeChan <- item
}
