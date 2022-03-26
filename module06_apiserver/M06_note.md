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





# 认证机制
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







# 授权机制
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





## 规划系统角色
















# 准入控制
## Mutating


## Validating



## Admission






# 限流方法





# 高可用 API Server







# 运作机制及apimachinery 组件









