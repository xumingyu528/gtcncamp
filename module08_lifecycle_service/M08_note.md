[toc]





# Pod 的生命周期





## Pod 状态机





## Qos & Quota

* Guaranteed
* Burstable
* BestEffort







# 服务发现

## 服务发布

* 需要把服务发布至集群内部或者外部，服务的不同类型：
  * ClusterIP（Headless）
  * NodePort
  * LoadBalancer
  * ExternalName
* 证书管理和七层负载均衡的需求
* 需要gRPC 负载均衡如何做？
* DNS需求
* 与上下游服务的关系



## 服务发布的挑战

### kube-dns

* DNS TTL 问题



### Service

* ClusterIP 只能对内
* kube-proxy 支持的 iptables/ipvs 规模有限
* IPVS 的性能和生产化问题
* kube-proxy 的 drift 问题
* 频繁的 Pod 变动（spec change，failover，crashloop）导致 LB 频繁变更
* 对外发布的 Service 需要与企业 ELB 集成
* 不支持 gRPC
* 不支持自定义 DNS 和高级路由功能



### Ingress

* Spec 的成熟度？



其它可选方案？







### 跨地域部署

* 需要多少实例？
* 如何控制失败域，部署在几个地区，AZ，集群？
* 如何进行精细的流量控制
* 如何做按照地域的顺序更新？
* 如何回滚？



## 微服务架构的挑战















# Service 对象

* Service Selector
* Ports







## Service 类型

* clusterIP
* nodePort
* LoadBalancer























## Endpoint 对象





















# kube-proxy 组件







## Netfilter框架







## iptables





## kube-proxy 工作原理

* watch api-server，监听到与节点或Pod相关的IP映射
* 调用iptables、ipvs等配置规则，实现功能







# DNS 原理和实践



## CoreDNS

CoreDNS 包含一个内存态 DNS，以及与其他 controller 类似的控制器。  

CoreDNS 的实现原理是，控制器监听 Service 和 Endpoint 的变化并配置 DNS，客户端 Pod 在进行域名解析时，从 CoreDNS 中查询服务对应的地址记录。



## 不同类型服务的 DNS 记录

* 普通 Service
* Headless Service
* ExternalName Service



## Kubernetes 中的 DNS 解析

* Kubernetes Pod 有一个与 DNS 策略相关的属性 DNSPolicy，默认值是 ClusterFirst

* Pod 启动后的 /etc/resolv.conf 会被改写，所有的地址解析优先发送至 CoreDNS

  * ```bash
    $ cat /etc/resolv.conf
    search ns1.svc.cluster.local svc.cluster.local cluster.local
    nameserver 192.168.0.100
    options ndots:3		//与第一条配合，域名的长度在几个以内会匹配search 中的后缀，以实现短域名访问服务
    ```

* 当 Pod 启动时，同一Namespace 的所有 Service 都会以环境变量的形式设置到容器内





## 自定义 DNSPolicy







# Ingress 对象

## 四层负载和七层负载





## Service 中的 Ingress 的对比





## Ingress 工作原理

* Ingress
  * Ingress 是一层代理
  * 负责根据 hostname 和 path 将流量转发到不同的服务上，使得一个负载均衡器用于多个后台应用
  * Kubernetes Ingress Spec 是转发规则的集合
* Ingress Controller
  * 确保实际状态（Actual）与期望状态（Desired）一直的 Control Loop
  * Ingress Controller 确保
  * 负载均衡配置
  * 边缘路由配置
  * DNS 配置







# 案例：通过 Ingress 和 Service 完成一个网络拓扑