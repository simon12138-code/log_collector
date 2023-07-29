package main

import (
	"context"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/coreos/etcd/clientv3"
	"time"
)

var Client *clientv3.Client
var logConfChan chan string

// 初始化etcd
func initEtcd(addr []string, keyfmt string, timeout time.Duration) (err error) {

	var keys []string
	for _, ip := range ipArrays {
		//keyfmt = /logagent/%s/log_config
		keys = append(keys, fmt.Sprintf(keyfmt, ip))
	}
	//修改配置通道
	logConfChan = make(chan string, 10)
	logs.Debug("etcd watch key:%v timeout:%v", keys, timeout)
	//创建etcd客户端
	Client, err = clientv3.New(clientv3.Config{
		Endpoints:   addr,
		DialTimeout: timeout,
	})
	if err != nil {
		logs.Error("connect failed,err:%v", err)
		return
	}
	logs.Debug("init etcd success")
	//每创建一个对应一个tailMgr
	waitGroup.Add(1)
	for _, key := range keys {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		// 从etcd中获取要收集日志的信息
		resp, err := Client.Get(ctx, key)
		cancel()
		if err != nil {
			logs.Warn("get key %s failed,err:%v", key, err)
			continue
		}
		//所以etcd监视要收集日志的信息，然后在logConf通道中加入对应的信息，然后异步调用才会正常进行
		for _, ev := range resp.Kvs {
			logs.Debug("%q : %q\n", ev.Key, ev.Value)
			logConfChan <- string(ev.Value)
		}
	}
	//前面的是初始化时对应的修改
	go WatchEtcd(keys)
	return
}

func WatchEtcd(keys []string) {
	// 这里用于检测当需要收集的日志信息更改时及时更新
	var watchChans []clientv3.WatchChan
	for _, key := range keys {
		rch := Client.Watch(context.Background(), key)
		watchChans = append(watchChans, rch)
	}

	for {
		//永久循环
		for _, watchC := range watchChans {
			select {
			//发现变化然后更新
			case wresp := <-watchC:
				for _, ev := range wresp.Events {
					logs.Debug("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
					logConfChan <- string(ev.Kv.Value)
				}
			default:

			}
		}
		time.Sleep(time.Second)
	}
	waitGroup.Done()
}

func GetLogConf() chan string {
	return logConfChan
}
