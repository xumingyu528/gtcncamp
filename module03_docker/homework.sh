#!/bin/bash

#Dockerfile 和 httpserver 在本脚本文件同级目录下
#构件本地镜像，使用练习2.2中的httpserver
docker build -t xumingyu/httpserver:v0.1 .

#将镜像推送至 docker 官方镜像仓库
docker push xumingyu/httpserver:v0.1

#通过 docker 命令本地启动 httpserver
docker run -d -p 8090:80  --name myhttpserver xumingyu/httpserver:v0.1

#通过nsenter进入容器查看IP配置
pid=`lsns -t net -n|grep httpserver|awk '{print $4}'`
nsenter -t $pid -n ip addr

