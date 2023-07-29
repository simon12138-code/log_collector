package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type logConfig struct {
	Topic    string `json:"topic"`
	LogPath  string `json:"log_path"`
	Service  string `json:"service"`
	SendRate int    `json:"send_rate"`
}

func main() {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect failed, err:", err)
		return
	}
	//cfg := logConfig{Topic: "log", LogPath: "/root/dockerfile/go/go/go_workspace/kitex-examples-main/bizdemo_copy/easy_note/log/output.202307281459.log", Service: "user", SendRate: 10000}
	//cfglist := []logConfig{cfg}
	//jsonBytes, err := json.Marshal(cfglist)

	fmt.Println("connect succ")
	defer cli.Close()
	//设置1秒超时，访问etcd有超时控制
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//_, err = cli.Put(ctx, "/logagent/192.168.112.100/log_config", string(jsonBytes))
	//操作etcd
	user := []string{"log"}
	userBytes, err := json.Marshal(user)
	_, err = cli.Put(ctx, "/logtransfer/192.168.112.100/log_config", string(userBytes))
	//操作完毕，取消etcd
	cancel()
	if err != nil {
		fmt.Println("put failed, err:", err)
		return
	}
	//取值，设置超时为1秒
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, "/logtransfer/192.168.112.100/log_config")
	cancel()
	if err != nil {
		fmt.Println("get failed, err:", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
}
