package main

import (
	"fmt"
	"gtcncamp/metrics"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	metrics.Register()

	http.HandleFunc("/", simpleServer)
	http.HandleFunc("/healthz", healthyCheck)
	http.Handle("/metrics", promhttp.Handler())

	//初始化一个listener监听80端口
	err := http.ListenAndServe("0.0.0.0:80", nil)
	if err != nil {
		fmt.Println("Error listening ", err.Error())
		return
	}

	//优雅终止

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	signal.Notify(sigint, syscall.SIGTERM)

	<-sigint

	println("server stoped!")

}

func simpleServer(w http.ResponseWriter, req *http.Request) {

	timer := metrics.NewTimer()
	defer timer.ObserveTotal()

	randSecond := rand.Intn(2000)
	println(randSecond)
	time.Sleep(time.Millisecond * time.Duration(randSecond))

	//遍历request的Header
	for key, value := range req.Header {
		//获取到的value为string数组，还需要遍历拿到所有的值赋给tempValue
		var tempValue string
		for _, i := range value {
			tempValue += i
		}
		//将key和tempValue写入response的Header中，为了便于区分在request的header前加了"client_"
		w.Header().Set("client_"+key, tempValue)
	}

	//获取go version信息,写入Header中
	w.Header().Set("Go-Version", runtime.Version())

	//Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
	http_code := 200
	w.WriteHeader(http_code)
	fmt.Printf("client info { IP: %s, HTTP_CODE: %d }\n", req.RemoteAddr, http_code)

	//返回给client页面内容
	fmt.Fprint(w, "hello \n")

}

//健康检查返回200
func healthyCheck(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	fmt.Fprintf(w, "StatusCode: %d\n", 200)
}
