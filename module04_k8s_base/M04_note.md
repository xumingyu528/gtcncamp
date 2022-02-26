[toc]

# Kubernetes架构基础

## Borg





## 主要组件









# 核心对象

## Node

* Node是Pod 真正运行的主机，可以是物理机，也可以是虚拟机
* 为了管理Pod，每个 Node 节点上至少要运行 container runtime（比如 Docker 或者Rkt ）、Kubelet 和 kube-proxy服务。



## Namespace

* Namespace 是对一组资源和对象的抽象集合，比如可以用来讲系统内部的对象划分为不同的项目组或用户组。

* 常见的 pods，services，replication controllers 和 deployments 等都是属于某一个Namespace 的（默认是default），而 Node，persistentVolumes 等则不属于任何 Namespace。





## Pod

* Pod 是一组紧密关联的容器集合，它们共享 PID、IPC、Network和UTS namespace，是kubernetes调度的基本单位。
* Pod 的设计理念是支持多个容器在一个Pod  中共享网络和文件系统，可以通过进程间通信和文件共享这种简单高效的方式组合完成服务
* 同一个 Pod 中的不同容器可共享资源
  * 共享网络 Namespace
  * 可通过挂在存储卷共享存储
  * 共享 Security Context



### 健康检查





### 存储





## ConfigMap











## Secret









## Service













## Replica Set









## 部署集（Deployment）









## 有状态服务集（Stateful Set）









## 任务（Job）







## 后台支撑服务集（DaemonSet）















## 存储 PV 和 PVC



















## CRD - CustomResourceDefinition















# 集群安装









