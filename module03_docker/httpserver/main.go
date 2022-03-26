package main

/*
上周作业
编写一个 HTTP 服务器，大家视个人不同情况决定完成到哪个环节，但尽量把 1 都做完：

1. 接收客户端 request，并将 request 中带的 header 写入 response header
2. 读取当前系统的环境变量中的 VERSION 配置，并写入 response header
3. Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
4. 当访问 localhost/healthz 时，应返回 200


本周作业
构建本地镜像
编写 Dockerfile 将练习 2.2 编写的 httpserver 容器化
将镜像推送至 docker 官方镜像仓库
通过 docker 命令本地启动 httpserver
通过 nsenter 进入容器查看 IP 配置

*/

import (
	"fmt"
	"net/http"
	"runtime"
)

func main() {
	fmt.Println("Starting Server ... ")
	http.HandleFunc("/", simpleServer)
	http.HandleFunc("/healthz", healthyCheck)
	//初始化一个listener监听80端口
	err := http.ListenAndServe("0.0.0.0:80", nil)
	if err != nil {
		fmt.Println("Error listening ", err.Error())
		return
	}

	//优雅终止
	// go func() {
	// 	sigint := make(chan os.Signal, 1)
	// 	signal.Notify(sigint, os.Interrupt)
	// 	signal.Notify(sigint, syscall.SIGTERM)

	// 	<-sigint

	// 	if err := http.

	// }()

}

func simpleServer(w http.ResponseWriter, req *http.Request) {
	//作业1，获取request的Header信息，写入到response header中

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

	//作业2，获取go version信息,写入Header中
	w.Header().Set("Go-Version", runtime.Version())

	//作业3， Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
	http_code := 200
	w.WriteHeader(http_code)
	fmt.Printf("client info { IP: %s, HTTP_CODE: %d }\n", req.RemoteAddr, http_code)

	//返回给client页面内容
	fmt.Fprint(w, "hello \n")

}

//作业4，健康检查返回200
func healthyCheck(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	fmt.Fprintf(w, "StatusCode: %d\n", 200)
}
