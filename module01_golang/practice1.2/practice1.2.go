package main

/*
课后练习 1.2
基于 Channel 编写一个简单的单线程生产者消费者模型：

队列：
队列长度 10，队列元素类型为 int
生产者：
每 1 秒往队列中放入一个类型为 int 的元素，队列满时生产者可以阻塞
消费者：
每一秒从队列中获取一个元素并打印，队列为空时消费者阻塞
*/

import (
	"fmt"
	"time"
)

func producer(ch chan int) {
	for i := 0; ; i++ {
		ch <- i
		fmt.Printf("Produce data: %d\n", i)
		time.Sleep(1 * time.Second)
	}
}

func consumer(ch chan int) {
	for {
		j := <-ch
		fmt.Printf("Consume data: %d\n", j)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	ch := make(chan int, 10)

	//测试阻塞注释下面两行代码
	go producer(ch)
	consumer(ch)

	//单独执行生产者，10条数据后没有接收则阻塞
	//producer(ch)

	//单独执行消费者，没有数据生产会阻塞
	//consumer(ch)
}
