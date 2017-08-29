package main

import (
	"os"
	"os/signal"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"xianxia/common/log"
	"encoding/json"
)

func writeLog() {
	conn, err := redis.Dial("tcp", "139.199.11.14:8888")
	if err != nil {
		fmt.Println("writeLog redis.Dial error:", err)
		return
	}

	for {
		reply, err := conn.Do("brpop", log.Redis_Log_Key, 0)
		data, err := redis.Values(reply, err)
		if err != nil || len(data) != 2 {
			fmt.Println("writeLog redis.Values or datalen != 2", err)
			continue
		}

		ldata, err := redis.String(data[1], err)
		if err != nil  {
			fmt.Println("writeLog redis.String ", err)
			continue
		}

		logInfo := &log.LogAllInfo{}
		err = json.Unmarshal([]byte(ldata), logInfo)
		if err != nil {
			fmt.Println("writeLog json.Unmarshal",err)
			conn.Close()
			continue
		}

		fmt.Println(logInfo)
	}
}

func main() {

	go writeLog()

	//close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	fmt.Println("LogServer closed by", sig)
}
