package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/hpcloud/tail"
	"strings"
	"sync"
)

type TailMgr struct {
	//因为我们的agent可能是读取多个日志文件，这里通过存储为一个map
	tailObjMap map[string]*TailObj
	lock       sync.Mutex
}

type TailObj struct {
	//这里是每个读取日志文件的对象
	tail     *tail.Tail
	secLimit *SecondLimit
	offset   int64 //记录当前位置
	//filename string
	logConf  logConfig
	exitChan chan bool
}

var tailMgr *TailMgr
var waitGroup sync.WaitGroup

func NewTailMgr() *TailMgr {
	tailMgr = &TailMgr{
		tailObjMap: make(map[string]*TailObj, 16),
	}
	return tailMgr
}

// 提供单个文件的更新变化跟踪类对象
func (t *TailMgr) AddLogFile(conf logConfig) (err error) {
	//上锁
	t.lock.Lock()
	defer t.lock.Unlock()
	//添加文件路径
	_, ok := t.tailObjMap[conf.LogPath]
	//尝试访问map，如果存在则说明重复
	if ok {
		err = fmt.Errorf("duplicate filename:%s\n", conf.LogPath)
		return
	}
	//创建tail跟随文件的修改状况
	tail, err := tail.TailFile(conf.LogPath, tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	})

	//创建一个跟随文件的结构体
	tailobj := &TailObj{
		//secLimit控制发送频率
		secLimit: NewSecondLimit(int32(conf.SendRate)),
		logConf:  conf,
		offset:   0,
		tail:     tail,
		exitChan: make(chan bool, 1),
	}

	t.tailObjMap[conf.LogPath] = tailobj
	logs.Info("map [%s]", t.tailObjMap)
	go tailobj.readLog()
	return
}

// 更新配置
func (t *TailMgr) reloadConfig(logConfArr []logConfig) (err error) {
	//遍历所有的更新配置
	for _, conf := range logConfArr {
		tailObj, ok := t.tailObjMap[conf.LogPath]
		//不存在就报错
		if !ok {
			logs.Debug("conf:%v -- tailobj:%v", conf, tailObj)
			err = t.AddLogFile(conf)
			if err != nil {
				logs.Error("add log file failed,err:%v", err)
				continue
			}
			continue
		}
		//更新配置项
		tailObj.logConf = conf
		t.tailObjMap[conf.LogPath] = tailObj
		//打印新的配置像
		logs.Info(t.tailObjMap)
	}
	// 处理删除的日志收集配置，方法是对每一个现存的tail在新的配置中寻找，如果没有则证明不存在，不存在就终止该tail对象跟进一个空文件（不需要跟随的文件）
	for key, tailObj := range t.tailObjMap {
		var found = false
		for _, newValue := range logConfArr {
			if key == newValue.LogPath {
				found = true
				break
			}
		}
		if found == false {
			logs.Warn("log path:%s is remove", key)
			tailObj.exitChan <- true
			delete(t.tailObjMap, key)
		}
	}

	return
}

// 开启管道等待对端发送新的log配置然后更新配置
func (t *TailMgr) Process() {
	//获取管道（指针复制）
	logChan := GetLogConf()
	//遍历管道时，如果没有接受值，则阻塞，所以当通道被创建后一定处理完所有事宜才会关闭
	for conf := range logChan {
		logs.Debug("log conf :%v", conf)
		var logConfArr []logConfig
		//将conf string反序列化到config对象上
		err := json.Unmarshal([]byte(conf), &logConfArr)
		if err != nil {
			logs.Error("unmarshal failed,err:%v conf:%v", err, conf)
			continue
		}
		logs.Debug("unmarshal succ conf:%v", logConfArr)
		//更新配置
		err = t.reloadConfig(logConfArr)
		if err != nil {
			logs.Error("realod config from etcd failed err:%v", err)
			continue
		}
		logs.Debug("reaload from etcd success,config:%v", logConfArr)
	}
}

// 读取日志并且发送到kafka消费者端
func (t *TailObj) readLog() {
	//读取每行日志内容
	for line := range t.tail.Lines {
		if line.Err != nil {
			logs.Error("read line failed,err:%v", line.Err)
			continue
		}
		//去掉多余换行符
		str := strings.TrimSpace(line.Text)
		if len(str) == 0 || str[0] == '\n' {
			continue
		}
		//发送给卡夫卡生产者
		kafkaSender.addMessage(line.Text, t.logConf.Topic)
		t.secLimit.Add(1)
		//可能是多携程操作，所以内部操作包含原子性，该方法可以限制发送速率
		t.secLimit.Wait()
		//每次发完后非阻塞等待退出信号
		select {
		case <-t.exitChan:
			logs.Warn("tail obj is exited,config:%v", t.logConf)
			return
		default:
		}

	}
	waitGroup.Done()
}

func RunServer() {
	//创建跟随组
	tailMgr = NewTailMgr()
	//接受信号然后发送消息
	tailMgr.Process()
	//异步执行等待
	waitGroup.Wait()
}
