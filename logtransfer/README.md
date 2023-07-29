# log_transfer
配合logagent消费日志保存到es中

目录结构
```shell
.
├── conf
│ └── app.conf #配置文件
├── config #热配置
│   ├── config.go 
│   └── config_notify.go
├── es.go #esclient
├── etcd.go #etcd监听器
├── go.mod
├── go.sum
├── ip.go #ip扫描
├── kafka.go #kafka消费者
├── logs
│   ├── transfer.2023-07-29.001.log
│   └── transfer.log
└── main.go

```
