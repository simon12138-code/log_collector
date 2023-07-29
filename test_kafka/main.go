package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

var Consumer sarama.Consumer

func main() {
	var err error
	Consumer, err = sarama.NewConsumer([]string{"192.168.112.100:9092"}, nil)
	if err != nil {
		fmt.Printf("fail to start consumer,err:%v\n", err)
		return
	}
	log.Print(Consumer)
	topics := []string{"log"}
	//topics := []string{"nsq"}
	//topic := "nsq"
	for _, topic := range topics {
		ctx, _ := context.WithCancel(context.Background())
		go testTopic(topic, ctx)
	}
	select {}
}
func testTopic(topic string, ctx context.Context) {
	partitionList, err := Consumer.Partitions(topic)
	fmt.Println(partitionList)
	if err != nil {
		fmt.Printf("fail to start consumer partition,err:%v\n", err)
		return
	}
	for partition := range partitionList {
		//  遍历所有的分区，并且针对每一个分区建立对应的消费者
		pc, err := Consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("fail to start consumer for partition %d,err:%v\n", partition, err)
			return
		}
		defer pc.AsyncClose()
		go testGetMsg(pc, ctx)
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}
func testGetMsg(partitionConsumer sarama.PartitionConsumer, ctx context.Context) {
	for msg := range partitionConsumer.Messages() {
		fmt.Printf("Partition:%d Offset:%v Key:%v Value:%v\n", msg.Partition, msg.Offset, msg.Key, string(msg.Value))
		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}
