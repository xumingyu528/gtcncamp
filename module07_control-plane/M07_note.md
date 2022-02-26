[toc]









# kube-scheduler

kube-scheduler 负责分配调度Pod 到集群内的节点上，它监听 kube-apiserver，查询还未分配的 Node 的 Pod，然后根据调度策略为这些 Pod 分配节点（更新 Pod 的 NodeName 字段）。  

调度器需要充分考虑诸多的因素：

* 公平调度
* 资源高效利用
* QoS
* affinity 和 anti-affinity
* 数据本地化（data locality）
* 内部负载干扰（inter-workload interfacerence ）
* deadlines





## 调度器策略

kube-scheduler 调度分为两个阶段，predicate 和 priority：

* predicate：过滤不符合条件的节点
* priority：优先级排序，选择优先级最高的节点



### Predicate策略

* PodFitsHostPorts：检查是否有 Host Ports 冲突
* PodFitsPorts：同 PodFitsHostPorts
* PodFitsResources：检查 Node 的资源是否充足，包括允许的 Pod 数量、CPU、内存、GPU个数以及其它的OpaqueIntResources
* HostName：检查 pod.Spec.NodeName 是否与候选节点一致
* MatchNodeSelector：检查候选节点的 pod.Spec.NodeSelector 是否匹配
* NoVolumeZoneConflict：检查 volume zone 是否冲突



### Priorities策略

* SelectorSpreadPriority：优先减少节点上属于同一个 Service 或 Replication Controller 的 Pod 数量
* InterPodAffinityPriority：优先将 Pod 调度到相同的拓扑上（如同一个节点、Rack、Zone等）
* LeastRequestedPriority：优先调度到请求资源少的节点上
* BalancedResourceAllocation：优先平衡各节点的资源使用
* NodePreferAvoidPodsPriority：alpha.kubernetes.io/preferAvoidPods 字段判断，权重为10000， 避免其他优先级策略的影响
* NodeAffinityPriority：优先调度到匹配 NodeAffinity 的节点上
* TaintTolerationPriority：优先调度到匹配 TaintToleration 的节点上
* ServiceSpreadingPriority：尽量将同一个 service 的 Pod 分布到不同节点上，已经被 SelectorSpreadPriority 替代（默认未使用）
* EqualPriority：将所有节点的优先级设置为1（默认未使用）
* ImageLocalityPriority：尽量将使用大镜像的容器调度到已经下拉了该镜像的节点上（默认未使用）
* MostRequestedPriority：尽量调度到已经使用过的 Node 上，特别适用于 cluster-autoscaler（默认未使用）





### 调度器实践

#### 资源需求

##### CPU资源需求



##### 内存资源需求



##### Init Container 资源需求



##### 把Pod调度到指定Node上





### Affinity 亲和度

#### NodeAffinity







#### PodAffinity









### Taints 污点 和 Tolerations 容忍度





## 多租户 Kubernetes 集群-计算资源隔离





# Controller Manager

## 控制器的工作流程





## 通用Controller

* Job Controller：处理job
* Pod AutoScaler：处理Pod 的自动缩容/扩容
* RelicaSet：
* Service Controller：
* ServiceAccount Controller：
* StatefulSet Controller：
* Volume Controller：
* Resource quota Controller：
* Namespace
* Replication
* Node
* Daemon
* Deployment
* Endpoint
* Garbage Collector：处理级联删除，比如删除 deployment 的同事删除 replicaset 以及 Pod
* CronJob





## Cloud Controller Manager







# kubelet组件

## kubelet 架构





## kubelet 管理 Pod 的核心流程





## kubelet 职责

每个节点上都运行一个 kubelet 服务进程，默认监听 10250 端口

* 接收并执行 master 发来的指令
* 管理 Pod 及 Pod 中的容器
* 每个kubelet 进程会在 API Server 上注册节点自身信息，定期向 master 节点汇报节点的资源使用情况，并通过 cAdvisor 监控节点和容器的资源





### 节点管理

节点管理主要是节点自注册和节点状态更新：

* kubelet 可以通过设置启动参数 --register-node 来确定是否向 API Server 注册自己
* 如果kubelet 没有选择自注册模式，则需要用户自己配置 Node 资源信息，同时需要告知 kubelet 集群上的 API Server 的位置
* kubelet 在启动时通过 API Server 注册节点信息，并定时向 API Server 发送节点新消息，API Server 在接收到新消息后，将信息写入 etcd





### Pod 管理

获取 Pod 清单：

* 文件：启动参数 --config 指定的配置目录下的文件（默认 /etc/Kubernetes/manifests/），该文件每20秒将重新检查一次（可配置）
* HTTP endpoint（URL）：启动参数 --manifest-url 设置。每 20秒检查一次这个端点（可配置）
* API Server：通过 API Server 监听 etcd 目录，同步 Pod 清单



### Pod启动流程

















# 容器运行时接口 CRI





# 容器网络接口 CNI







# 容器存储接口 CSI







# Rook 的工作原理













