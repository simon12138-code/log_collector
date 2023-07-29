# log-collector

## 目录结构
```shell
    .
├── logagent-master
│   ├── conf
│   │   └── app.conf
│   ├── config
│   │   ├── config.go
│   │   └── config_notify.go
│   ├── config_log.go
│   ├── data.go
│   ├── etcd.go
│   ├── go.mod
│   ├── go.sum
│   ├── ip.go
│   ├── kafka.go
│   ├── limit.go
│   ├── logs
│   │   ├── logcollect.2023-07-29.001.log
│   │   └── logcollect.log
│   ├── main.go
│   ├── README.md
│   └── server.go
├── logtransfer
│   ├── conf
│   │   └── app.conf
│   ├── config
│   │   ├── config.go
│   │   └── config_notify.go
│   ├── es.go
│   ├── etcd.go
│   ├── go.mod
│   ├── go.sum
│   ├── ip.go
│   ├── kafka.go
│   ├── logs
│   │   ├── transfer.2023-07-29.001.log
│   │   └── transfer.log
│   ├── main.go
│   └── README.md
├── README.md
├── test_etcd
│   ├── go.mod
│   ├── go.sum
│   └── main.go
└── test_kafka
    ├── go.mod
    ├── go.sum
    └── main.go
```
### log-agent
收集日志文件并且生产到kafka中等待消费
### log-transfer
消费kafka中的日志并保存到es中
### test_etcd
发出收集日志指令
### test_kafka
消费对应日志并进行检测