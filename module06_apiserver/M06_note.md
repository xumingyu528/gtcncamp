[toc]



# API Server 概念
kube-apiserve 是 Kubernetes 最重要的核心组件之一，主要提供以下功能
* 集群管理的REST API接口，包括认证授权、校验及集群状态变更等
* 提供其他模块之间的数据交互和通信的枢纽（其他模块通过 API Server 查询或修改数据，只有 API Server 才能连接操作 etcd


## 访问控制概念
Kubernetes API 的每个请求都会经过多阶段的访问控制之后才会被接受，包括认证、鉴权、准入等

![](./note_images/api-server_processing.png)


### 访问控制细节

* 接收到请求后，http handler 会对请求做处理，包括
  * panic recover（handler 会启动新的goruntine 处理请求，有可能出现 panic）
  * request-timeout，设置请求的超时时间
  * authentication，认证请求合法性
  * audit，对请求进行审计记录
  * impersonation，可以对请求添加一些 header，用于角色确认
  * max-in-flight，限流
  * authorization，鉴权
* kube-aggregator & CRDs，判断数据是否为自定义的对象(CRD)来处理，是的话交给自定义的 API Server 处理，否则默认交给本地的 API Server处理
* resource handler，
  * decoding，将请求的数据（大多为json）反序列化为 kubernetes 对象
  * request conversion & defauting，判断反序列化后的对象具体是什么类型，做 conversion 转换
  * admission，准入，先判断是否有mutating webhook，有就触发，没有就走默认的 validation ，然后判断是否有 validating webhook。
    * validation，会调用对应类型的方法，按照 strategy 操作，判断数据是否符合这些 strategy
  * 通过以上 admission 以后，才会存入 etcd，返回数据给用户


![](./note_images/api-server_processing_details.png)





# 认证机制 Authorization
开启 TLS 时，所有的请求都首先需要认证。 Kubernetes 支持多种认证机制，并支持同时开启多个认证插件（只要有一个认证通过即可）。如果认证通过，则用户的 username 会传入授权模块进行进一步授权验证；如果认证失败则返回 HTTP 401 。

* Insecure port：不开启认证，不建议这样做，kubernetes 集群将不会做任何认证、鉴权操作。  
* secure port：开启后需要经过 Kubernetes 一系列认证流程


## 认证插件
### X509证书
* 在 API Server 启动时配置 --client-ca-file=SOMEFILE。在证书认证时，其 CN 域用作用户名，而组织机构域则用作 group 名。


### 静态 Token 文件
* 在 API Server 启动时配置 --token-auth-file=SOMEFILE。
* 文件类型为 csv 格式，每行至少包括三列：token, username, userid, token, user, uid, "group1, group2, group3"



### 引导 Token
* Bootstrap Token
* 令牌以 Secret 的形式存放在 kube-system 名称空间中


### 静态密码文件
* 在 API Server 启动时配置 --basic-auth-file=SOMEFILE。
* 文件格式为 csv，每行至少包括三列：password，user，uid，后面是可选的 group 名 password, user, uid, "group1, group2, group3"


### ServiceAccount
* Kubernetes 自动生成，并挂载到容器的 /run/secrets/kubernetes.io/serviceaccount 目录中。


### OpenID
* OAuth 2.0 的认证机制


### Webhook 令牌身份认证
对接企业内部的认证系统，大多采用此方式  
* --authentication-token-webhook-config-file 指向一个配置文件，其中描述如何远程访问的 Webhook 服务。
* --authentication-token-webhook-cache-ttl 用来设定身份认证决定的缓存时间，默认为2分钟。

### 匿名请求
* 如果使用 AlwaysAllow 以外的认证模式，则匿名请求默认开启，但可用 --anonymous-auth=false 禁用匿名请求。

### 演示

* 静态 Token 认证
  ```bash
  # 先准备好 static-token 文件
  mkdir -p /etc/kubernetes/auth
  # static-token 内容填写：xmy-token xmy 999 "gp1, gp2, gp3"

  # 搭建好kubernetes集群后，在/etc/kubernetes/manifests/下有kube-apiserver.yaml文件，操作前将文件备份
  cp /etc/kubernetes/manifests/kube-apiserver.yaml ~/kube-apiserver.yaml

  # 编辑该文件，增加 static token 认证配置
  vim /etc/kubernetes/manifests/kube-apiserver.yaml

  ...
  # 指定token 文件位置
  - --token-auth-file=/etc/kubernetes/auth/static-token
  ...
  # API Server 的container 配置部分，增加mountPath 挂载token 文件，来自下面的hostpath 挂载卷auth-files 
  - mountPath: /etc/kubernetes/auth
      name: auth-files
      readOnly: true
  ...
  # volumes 中增加hostPath 相关配置
  - hostPath:
      path: /etc/kubernetes/auth
      type: DirectoryOrCreate
  name: auth-files


  # 修改完成后，相关组件会重启加载配置文件
  # 使用手动请求
  # 格式：curl https://${这里填写api server 的地址和端口}/api/v1/namespace/default -H "Authorization: Bearer xmy-token" -k  
  # 下面是我机器中的请求及返回内容，由于没有权限访问，因此返回403，但是说明已经通过了认证步骤，到达鉴权步骤
  root@master01:~# curl https://10.0.12.2:6443/api/v1/namespace/default -H "Authorization: Bearer xmy-token" -k
  {
    "kind": "Status",
    "apiVersion": "v1",
    "metadata": {},
    "status": "Failure",
    "message": "namespace \"default\" is forbidden: User \"xmy\" cannot get resource \"namespace\" in API group \"\" at the cluster scope",
    "reason": "Forbidden",
    "details": {
      "name": "default",
      "kind": "namespace"
    },
    "code": 403
  }

  ```
* X509认证
  ```bash
  # 生成证书
  openssl genrsa -out myuser.key 2048
  # 根据key文件输出csr
  openssl req -new -key myuser.key -out myuser.csr
  # 将csr文件以base64形式输出，替换至下面的csr对象(CertificateSigningRequest)中
  cat myuser.csr | base64 |tr -d "\n"

  # 替换下方request中内容为上一步的输出内容
  cat << EOF | kubectl apply -f -
  apiVersion: certificates.k8s.io/v1
  kind: CertificateSigningRequest
  metadata:
    name: myuser
  spec:
    request: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJQ3lqQ0NBYklDQVFBd2dZUXhDekFKQmdOVkJBWVRBa05PTVJJd0VBWURWUVFJREFsSGRXRnVaM3BvYjNVeApFVEFQQmdOVkJBY01DRk5vWlc1NmFHVnVNUTh3RFFZRFZRUUtEQVp0ZVhWelpYSXhEekFOQmdOVkJBc01CbTE1CmRYTmxjakVQTUEwR0ExVUVBd3dHYlhsMWMyVnlNUnN3R1FZSktvWklodmNOQVFrQkZneHRlWFZ6WlhKQVp5NWoKYjIwd2dnRWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUURVaXZJaGwwR2tIS1J6eXlpbgpPUERnVlBTSkRQSVdUb1Y3azhIMEJVRkJ2aHE4ZlIwcmd4UWVpMHdZTlVYNGdKaitiY3dBbWp2VFJTamdtWXRGCmhraTBkM0FHTng3Y0RzVXBoNi9BVk52NkRFOEEwclQ0N0RoYkVGZmVyYkRobmhqYWtacWlwbWtPKzZBbEFEb1AKQm1aYUxhRWNqdnhINGNKQU9OdG52Ny9SUzI5RmxqVzhtbW9acktHZXdJMTYwWUhXSlc4ZlpTRkJnV2ljQmlMSgpMN0VJN1dDcUZ3MGlGeW5TOHZlV0FsRmd2OHVzbDh5RzZzSk41Q2xxTGJvTHBGKzZJN2kwUFEwTDVzMmU3TXpZCko3dW0raEovUmZYZDN5cmZmeXdsTy83bmRYVExUUHRia0tYaElMeDdDM3FYVnBESHZiNThhbWtJYlRXMlNDQkoKejhzSEFnTUJBQUdnQURBTkJna3Foa2lHOXcwQkFRc0ZBQU9DQVFFQXlMQXYyUUVqUmY5QU01djdtRE56UnhTMgpGVFdHaDRSQ3dOaHRZMFJpQXpGQkczZVVhN0F2eHA4TEowSHl3akYycjdteVZZVlVqQ01rdzZZeDEzL2JnemJ5ClJKLzNnMzN0Rkp2NmtvTjJiZCsxZnZZcjhSNEVuMURPcGllOHZxMHRWbnNmTWpwTENReWtTUHd5aHBnZUNsRm4KS3VoVWZuM1FMUU1EcFlMUGdCbWl4bnpPbmFYd2ZMTGlrOVFXQWkwZitpdjB4K2x2OTkraHFvQjgySGYzdzhadQo2KzdWVnREUFg3aGxSSTBvK21Jb2lFRnNHVGJpRmZNR0Y3eUpqTTBuNkxGNkp4RHNDazkvMzM3djhvVGoybXNpClZkdHkwSzltZncvMm1FM1lKY0pydloxbkx5VzExclNiT1NVS3JFTEsvaWJqSU4wRERNd3N2UmI4c0hCZzNnPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg==
    signerName: kubernetes.io/kube-apiserver-client
    expirationSeconds: 86400  # one day

    usages:
    - client auth
  EOF
  ```
  ```bash
  # 查看kubernetes集群的csr对象
  k get csr
  # 通过kubernetes签发证书给用户myuser
  kubectl certificate approve myuser
  # 将myuser的csr中.status.certificate部分导出为crt
  k get csr myuser -o jsonpath='{.status.certificate}'|base64 -d > myuser.crt

  # 配置myuser和对应的key、crt 到kube的config中
  k config set-credentials myuser --client-key=myuser.key --client-certificate=myuser.crt --embed-certs=true
  # 查看用户目录.kube下的config文件
  cat ~/.kube/config 
  # 输出内容的 users 部分包括kubernetes-admin的和myuser用户的信息：
  - name: myuser
  user:
    client-certificate-data: 
      # 证书内容。。。
    client-key-data: 
      # 证书key。。。

  # 通过myuser获取pods信息，会提示无权限访问defaultnamespace，但是说明证书已经被API Server认证通过
  k get po --user myuser
  Error from server (Forbidden): pods is forbidden: User "myuser" cannot list resource "pods" in API group "" in the namespace "default"

  # 创建一个角色，和rolebinding，将myuser绑定授权访问pods资源
  k create role developer --verb=create --verb=get --verb=list --verb=update --verb=delete --resource=pods
  k create rolebinding developer-binding-myuser --role=developer --user=myuser

  # 再次通过myuser 获取 pod 资源可以获取到
  k get po --user myuser
  NAME                                           READY   STATUS    RESTARTS      AGE
  httpserver-65cbb484d-bdbz7                     1/1     Running   0             9d
  jenkins-0                                      1/1     Running   1 (10d ago)   15d


  ```

* Token -- ServerAccount
  ```bash
  # Token 存放在secret中
  # k get secret 可以查看secret
  # k get secret secret名 -oyaml  可以输出secret中的数据，token字段为base64加密
  # 解码token
  echo token加密字段 | base64 -d 
  # 输出的内容为token，可以在请求API Server时，在头部加入该token，API Server会识别对应的用户
  # 如下示例，和静态token访问方式一致，将token部分替换为上面输出的内容
  curl https://10.0.12.2:6443/api/v1/namespace/default -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjQyaDhlTnowMlZkQmxNUFlaTjZKT0REYnJtdVpxcEJQenYtWDhOVE1zSW8ifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImplbmtpbnMtdG9rZW4tbnAyaDIiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiamVua2lucyIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjZmNzIzM2IyLWEyNjktNDJhNS05ZDIxLTljNTdmZjBhYmI3ZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmplbmtpbnMifQ.KXCa-K8PYdEvRttF17_EwfgZQdAV1ICkBFZ0OxmxUfujuNb9o3mMSAeqA37lY52yvX_zG-GSfM6LS6uWtNKInWXdkTrJ5kzMKbWmb6dXgM2mVvIn4WxMfMwVUaLD-VSQ5e9Og2fJKxHVc7Yza8zLdsAj_9Kmgo38tFe4kl2FDTQr3vjefoaAdoZxXafca_HF9hiUPij6E_2OBxtZAjlUv-vHZp-Oxt7l9F3_iXntw4pOK8scXc1rSy6f-K7gm5Buw2-t7MvGi2We0vqMHcVoSYGABEyuZ76iPTGO8yaBVfuHy1mtUxXep6OZF0KFuAmSAGB-L0bRpdHlcRopu0Qbew" -k

  ```





# 基于 Webhook 的认证服务集成

## 构件符合 Kubernetes 规范的认证服务
需要依照 Kubernetes 规范，构件认证服务，用来认证 tokenreview request。  
构件认证服务。
* 认证服务需要满足如下 Kubernetes 的规范：
  * URL：https://authn.example.com/authenticate
  * Method: POST
  * Input:
  * Output:



## 配置 apiserver
可以使任何认证系统：
* 但在用户认证完成后，生成代表用户身份的token；
* 该 token 通常是有失效时间的；
* 用户获取该 token 以后，将 token 配置进 kubeconfig。

修改 apiserver 设置，开启认证服务，apiserver 保证将所有收到的请求中的 token 信息，发给认证服务进行验证。
* --authentication-token-webhook-config-file， 改文件描述如何访问认证服务。
* --authentication-token-webhook-cache-ttl ，默认为2分钟。
配置文件需要 mount 进 Pod。  
配置文件中的服务器地址需要指向 authService。



### 参考链接
[https://github.com/appscode/guard](https://github.com/appscode/guard)


## 认证演示（待完成）
### 开发认证服务





### 认证服务工作流程










## 生产系统中遇到的陷阱（ebay）
>
>  基于 Keystone 的认证插件导致 Keystone 故障且无法恢复。  
> Keystone 是企业的关键服务。  
> Kubernetes 以 Keystone 作为认证插件。  
> Keystone 在出现故障后会抛出 401 错误。  
> Kubernetes 发现 401 错误后会尝试重新认证。  
> 大多数 controller 都有指数级 back off，重试间隔越来越慢。  
> 但 gophercloud 针对过期 token 会一直 retry。  
> 大量的 request 积压在 Keystone 导致服务无法恢复。  
> Kubernetes 成为压死企业认证服务的最后一根稻草。  
>
> 解决方案？
> * Circuit break
> * Rate limit







# 授权机制 authentication
授权主要是用于对集群资源的访问控制，通过检查请求包含的相关属性值，与相对应的访问策略相比较，API 请求必须满足某些策略才能被处理。  
跟认证类似，Kubernetes 也支持多种授权机制，并支持同时开启多个授权插件（只要有一个验证通过即可）。  
如果授权成功，则用户的请求会发送到准入控制模块作进一步请求验证；对鱼授权失败的请求则返回 HTTP 403。  


Kubernetes 授权需要处理以下的请求属性：
* user，group，extra
* API、请求方法（如 get、post、update、patch和delete）和请求路径（如/api）
* 请求资源和子资源
* Namespace
* API Group

目前 Kubernetes 支持以下授权插件：
* ABAC 
* RBAC
* Webhook
* Node

## RBAC vs ABAC
ABAC（Attribute Based Access Control）  
在 Kubernetes 中实现比较难于管理和理解，而且需要 Master 所在节点的 SSH 和文件系统权限，要使得对授权的变更成功生效，还需要重新启动 API Server。  
RBAC（Role Base Access Control）   
可以利用 kubectl 或者 Kubernetes API 直接进行配置。RBAC可以授权给用户，让用户有权进行授权管理，这样就可以无需接触节点，直接进行授权管理。RBAC在 Kubernetes 中被映射为 API 资源和操作。 

### RBAC 老图
* 将不同对象和操作绑定为Permissions，每一个Role设定不同Permission
* 每个 User 分配一个或多个 Role

![](./note_images/RBAC_old.png)

### RBAC 新解
Kubernetes 基于 RBAC 设计的一套用户角色管理机制
![](./note_images/RBAC_new.png)


## Role与ClusterRole
Role 是一系列权限的集合，Role 只能用来给某个特定 namespace 中的资源做鉴权，对多 namespace 和集群级的资源或者是非资源类的API使用 ClusterRole  

```yaml
# Role 示例
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: default
  name: pod-reader
rules:
- apiGroups: [""]   # ""表示核心API group
  resources: ["pods"]
  verbs: ["get","watch","list"]
```

```yaml
# ClusterRole 示例
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:   # ClusterRole 没有namespace，是全局范围的
  name: secret-reader
rules:
- apiGroups: [""] 
  resources: ["secrets"]
  verbs: ["get","watch","list"]
```

```yaml
# binding
# 配置是允许 dave 读取 development 的namespace下的secrets 
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: read-secrets
  namespace: development 
subjects:
- kind: User
  name: dave
  apiGroup: rbac.authorization.k8s.io
roleRef:  #引用 ClusterRole
  kind: ClusterRole
  name: secret-reader
  apiGroup: rbac.authorization.k8s.io

```

### 账户/组的管理
**角色绑定（Role Binding）** 是将角色中定义的权限赋予一个或者一组用户。  
它包含若干 **主体**（用户、组或服务账户）的列表和对这些主体所获得的角色的引用。  
组的概念：
* 当与外部认证系统对接时，用户信息（UserInfo）可包含 Group 信息，授权可针对用户群组。
* 当对 ServiceAccount 授权时，Group 代表某个namespace 下的所有ServiceAccount



```yaml
# 针对群组授权
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:   
  name: read-secrets-global
subjects:
- kind: Group
  name: manager # 区分大小写
  apiGroup: rbac.authorization.k8s.io
roleRef:  #引用 ClusterRole
  kind: ClusterRole
  name: secret-reader
  apiGroup: rbac.authorization.k8s.io
```



```yaml
# 对service account 授权
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:   
  name: read-secrets-global
subjects:
- kind: serviceaccount  # 也可以针对sa对应的group 授权，下面的name 只写到namespace一级
  name: system:serviceaccount:qa
  apiGroup: rbac.authorization.k8s.io
roleRef:  #引用 ClusterRole
  kind: ClusterRole
  name: secret-reader
  apiGroup: rbac.authorization.k8s.io
```

## 规划系统角色

User 
* 管理员
  * 根据不同 case 场景配置
* 普通用户
  * 是否有该用户创建的 namespace 下的所有 object 的操作权限
  * 对其它用户的namespace 资源是否可读，是否可写？

SystemAccount
* SystemAccount
* 用户可以创建自定义的 ServiceAccount
* Default ServiceAccount 通常需要给定权限以后才能对 api server 做写操作

## 实现方案
针对 namespace 创建admin用户  
在 cluster 创建时，创建自定义的 role ，比如 namespace-creator。 
namespace-creator role 定义用户可操作的对象和对应的读写操作。  

创建自定义的 namespace admission controller：  
* 当 namespace 创建请求被处理时，获取当前用户信息并 annotate 到 namespace


创建 RBAC controller：
* Watch namespace 的创建事件
* 获取当前 namespace 的创建者信息
* 在当前 namespace 创建 rolebinding 对象，并将 namespace-creator 角色和用户绑定







# 准入控制
## 应用场景  
配额管理
* 原因： 资源有限，限定某个用户的资源数量

方案
* 预定义每个 namespace 的 ResourceQuota，并把 spec 保存为 configmap；
  * 用户可以创建多少个 Pod
    * BestEffortPod
    * QoSPod
  * 用户可以创建多少个 service
  * 用户可以创建多少个 ingress
  * 用户可以创建多少个 service VIP
* 创建 ResourceQuota Controller
  * 监控 namespace 创建事件，当 namespace 创建时，在该 namespace 创建对应的 ResourceQuota 对象




## 准入控制插件
* AlwaysAdmit：
* AlwaysPullImages：
* DenyEscalatingExec：
* ImagePolicyWebhook：
* ServiceAccount：
* SecurityContextDeny：
* ResourceQuota：
* LimitRanger：
* InitialResources：
* NamespaceLifecycle：
* DefaultStorageClass：
* DefaultTolerationSeconds：
* PodSecurityPolicy：
* NodeRestriction：

```bash
# 查看apiserver 帮助中可以从中检索admission关键字
kubectl exec -it kube-apiserver-master01 -n kube-system -- kube-apiserver -h

# 其中有 --disable-admission-plugins 和 --enable-admission-plugins 代表了默认开启和关闭的插件列表

```

## 准入控制插件开发



## Mutating


## Validating



## Admission






# 限流方法

## 计数器固定窗口算法
原理就是对一段固定时间窗口内的请求进行计数，如果请求数超过了阈值，则舍弃该请求；  
如果没有达到设定的阈值，则接收该请求，且计数加1.  
当时间窗口结束时，重置计数器为0.  

## 计数器滑动窗口算法


## 漏斗算法


## 令牌桶算法



## APIServer 中的限流

## 传统限流方法的局限性
* 粗粒度
* 单队列
* 不公平
* 无优先级


## API Priority and Fairness 
* APF 以耕细粒度的方式对请求进行分类和隔离
* 它还引入了空间有限的排队机制，因此在非常短暂的突发情况下，API 服务器不会拒绝任何请求
* 通过使用公平排队技术从队列中分发请求，这样一个行为不佳的控制器就不会饿死其他控制器（即使优先级相同）
* APF的核心
  * 多等级
  * 多队列

* APF 的实现依赖两个非常重要的资源 FlowSchema，PriorityLevelConfiguration
* APF 对请求进行更细粒度的分类，每一个请求分类对应一个FlowSchema（FS）
* FS 内的请求又会根据 distinguisher 进一步划分为不同的 Flow
* FS 会设置一个优先级（Priority Level，PL），不同优先级的并发资源是隔离的。所以不同优先级的资源不会相互排挤。特定优先级的请求可以被高优处理。


* 一个 PL 可以对应多个 FS，PL 中维护了一个 QueueSet，用于缓存不能及时处理的请求，请求不会因为超出 PL 的并发限制而被丢弃
* FS 中的每个 Flow 通过 shuffle sharding 算法从 QueueSet  选取特定的 queues 缓存请求
* 每次从 QueueSet 中请求执行时，会先应用 fair queuing 算法 QueueSet 中选中一个 queue ，然后从这个queue 中选取出 oldest 请求执行。所以即使同一个 PL 内的请求，也不会出现一个 Flow 内的请求一直占用资源的不公平现象。

### 概念
* 传入的请求通过 FlowSchema 按照其属性分类，并分配优先级
* 每个优先级维护自定义的并发限制，加强了隔离度，这样不同优先级的请求，就不会相互饿死
* 在同一个优先级内，公平排队算法可以防止来自不同 flow 的请求相互饿死
* 该算法将请求排队，通过配对机制，防止在平均负载较低时，通信量突增而导致请求失败

```bash
# kubernetes 默认的flowschema
root@master01:~# k get flowschema
NAME                           PRIORITYLEVEL     MATCHINGPRECEDENCE   DISTINGUISHERMETHOD   AGE   MISSINGPL
exempt                         exempt            1                    <none>                46d   False
probes                         exempt            2                    <none>                46d   False
system-leader-election         leader-election   100                  ByUser                46d   False
workload-leader-election       leader-election   200                  ByUser                46d   False
system-node-high               node-high         400                  ByUser                46d   False
system-nodes                   system            500                  ByUser                46d   False
kube-controller-manager        workload-high     800                  ByNamespace           46d   False
kube-scheduler                 workload-high     800                  ByNamespace           46d   False
kube-system-service-accounts   workload-high     900                  ByNamespace           46d   False
service-accounts               workload-low      9000                 ByUser                46d   False
global-default                 global-default    9900                 ByUser                46d   False
catch-all                      catch-all         10000                ByUser                46d   False
root@master01:~# 
```



### 排队
* 即使在同一优先级内，也可能存在大量不同的流量源
* 在过载的情况下，防止一个请求流饿死其他流失非常有价值的（尤其是在一个较为常见的场景中，一个有故障的客户端会疯狂地向 kube-apiserver 发送请求，理想情况下，这个有故障的客户端不应对其他客户端产生太大的影响）。
* 公平排队算法在处理具有相同优先级的请求时，实现了上述场景

* 每个请求都会被分配到某个流中，该流由对应的 FlowSchema 的名字加上一个流区分项（FlowDistinguisher）来标识
* 这里的流区分项可以是发出请求的用户、目标资源的名称空间或什么都不是
* 系统尝试为不同流中具有相同优先级的请求赋予近似相等的权重
* 将请求划分到流中之后，APF 功能将请求分配到队列中
* 分配时使用一种称为混洗分片（Shuffle-Sharding）的技术。该技术可以相对有效地利用队列隔离低强度流与高强度流
* 排队算法的细节可针对每个优先等级进行调整，并允许管理员在内存占用、公平性（当总流量超标时，各个独立的流将都会取得进展）、突发流量的容忍度以及排队引发的额外延迟之间进行权衡



### 豁免请求
exempt


### 默认配置
* system
* leader-election
* workload-high
* workload-low
* global-default



### PriorityLevelConfiguration
一个 PriorityLevelConfiguration 标识单个隔离类型   
每个 PriorityLevelConfigurations 对未完成的请求数有各自的限制，对排队中的请求数也有限制

```yaml
apiVersion: flowcontrol.apiserver.k8s.io/v1beta1
kind: PriorityLevelConfiguration
metadata:
  name: global-default
spec:
  limited:
    assuredConcurrencyShares: 20  #允许的并发请求数
    limitResponse:
      queuing:
        handSize: 6 #shuffle sharding 的配置，每个 flowschema+distinguisher 的请求会被 enqueue 到多少个队列
        queueLengthLimit: 50  # 每个队列中的对象数量
        queues: 128 #当前PriorityLevel 的队列总数
      type: Queue
    type: Limited
```


### FlowSchema
```yaml
apiVersion: flowcontrol.apiserver.k8s.io/v1beta1
kind: FlowSchema
metadata:
  name: kube-scheduler  # FlowSchema名
spec:
  distinguisherMethod:
    type: ByNamespace # Distinguisher 和 FlowSchema名一起确定一个flow
  matchingPrecedence: 800 # 规则优先级
  priorityLevelConfiguration: # 对应的队列优先级
    name: workload-high
  rules:
  - resourceRules:
    - resources:  # 对应的资源和请求类型
      - '*'
      verbs:
      - '*'
    subjects:
    - kind: User
      user:
        name: system:kube-scheduler
```


### 调试
* /debug/api_priority_and_fairness/dump_priority_levels   所有优先级及其当前状态
  * 用法：`k get --raw /debug/api_priority_and_fairness/dump_priority_levels`
* /debug/api_priority_and_fairness/dump_queues    所有队列及其当前状态的列表
* /debug/api_priority_and_fairness/dump_requests    当前正在队列中等待的所有请求的列表






# 高可用 API Server

## RateLimit
限制速率，防止API Server 被压垮

## 设置合适的缓存大小
客户端请求 ListOption 中没有设置 ResourceVersion，API Server 会直接从 etcd 拉取最新数据。  
客户端尽量避免此操作，应该在 ListOption 中设置 resourceVersion 为0，API Server 则会去缓存中读取数据。  

## 客户端尽量使用长链接
避免全量从API Server 获取资源

## 如何访问 API Server
对于外部用户，永远只通过 LoadBalancer 访问  


## 搭建多租户的 Kubernetes 集群
授信  
* 认证
  * 禁止匿名访问，只允许可信用户做操作
* 授权
  * 基于授信的操作，防止多用户之间互相影响，比如普通用户删除 Kubernetes 核心服务，或者 A 用户删除或修改 B 用户的应用

隔离
* 可见性隔离
  * 用户只关心自己的应用，无需看到其它用户的服务和部署
* 资源隔离
  * 有些关键项目对资源需求较高，需要专有设备，不与其他人共享
* 应用访问隔离
  * 用户创建的服务，按照既定规则允许其它用户访问

资源管理
* Quota 管理
  * 谁能用多少资源

## 规划系统角色

User
* 管理员
* 普通用户

SystemAccount
* SystemAccount
* ServiceAccount
* Default ServiceAccount


# API Server 运作机制及apimachinery 组件
## 回顾 GKV
* Group
* Kind
* Version
  * 所有版本都是从 V1alpha1 开始，需要向前兼容
  * Internel version：API Server 存放到etcd之前会转换成 Internel version
  * External version：面向 Kubernetes 集群外部
  * conversion 版本转换：External 与 Internel 之间需要转换

## 代码解读
### Group
Group 一般在 core/register.go 中定义 SchemeGroupVersion，
addKnownTypes 将对象加入Group中  

types.go 定义了每个对象
* List 对象，例如Pod 、Service
* 单一对象将数据结构
  * TypeMeta
  * ObjectMeta
  * Spec
  * Status


### 代码生成
* Tag
* code generate
* etcd storage
  * 一般是 storage.go 中定义，包含创建、更新、删除等策略


### subresource
内嵌在 Kubernetes 对象中，有独立的操作逻辑的属性集合，如 podstatus


### 注册 APIGroup
* 定义 Storage
* 定义对象的 StorageMap
* 将对象注册至 API Server（挂载 handler）

### 代码生成方法
* deepcopy-gen
* client-gen
* informer-gen
* lister-gen
* conversion-gen




