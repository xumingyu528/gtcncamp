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











# 基于 Webhook 的认证服务集成





# 授权机制





# 准入控制
## Mutating


## Validating



## Admission






# 限流方法





# 高可用 API Server







# 运作机制及apimachinery 组件









