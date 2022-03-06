#### 1. **第一部分**

现在你对 Kubernetes 的控制面板的工作机制是否有了深入的了解呢？
是否对如何构建一个优雅的云上应用有了深刻的认识，那么接下来用最近学过的知识把你之前编写的 http 以优雅的方式部署起来吧，你可能需要审视之前代码是否能满足优雅上云的需求。
作业要求：编写 Kubernetes 部署脚本将 httpserver 部署到 Kubernetes 集群，以下是你可以思考的维度。

- 优雅启动
- 优雅终止
- 资源需求和 QoS 保证
- 探活
- 日常运维需求，日志等级
- 配置和代码分离









deployment配置yaml文件

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpserver
  namespace: default
  labels:
    app: httpserver
    version: v0.1
spec:
  replicas: 1
  selector: # 选择器
    matchLabels: # 匹配标签
      app: httpserver
      version: v0.1
  strategy: # 策略
    rollingUpdate: # 滚动更新
      maxSurge: 30% # 最大额外可以存在的副本数，可以为百分比，也可以为整数
      maxUnavailable: 30% # 在更新过程中能够进入不可用状态的 Pod 的最大值，可以为百分比，也可以为整数
    type: RollingUpdate # 滚动更新策略
  template:
    metadata: # 资源的元数据/属性 
      annotations: # 自定义注解列表
        sidecar.istio.io/inject: "false" # 自定义注解名字
      labels: # 设定资源的标签
        app: httpserver
        version: v0.1
    spec: # 资源规范字段
      containers:
      - name: httpserver # 容器的名字   
        image: xumingyu/httpserver:v0.1 # 容器使用的镜像地址   
        imagePullPolicy: IfNotPresent 
        resources: # 资源管理
          limits: # 最大使用
            cpu: 300m # CPU，1核心 = 1000m
            memory: 500Mi # 内存，1G = 1024Mi
          requests:  # 容器运行时，最低资源需求，也就是说最少需要多少资源容器才能正常运行
            cpu: 100m
            memory: 100Mi
        livenessProbe: # pod 内部健康检查的设置
          httpGet: # 通过httpget检查健康，返回200-399之间，则认为容器正常
            path: /healthz # URI地址
            port: 80 # 端口
            scheme: HTTP # 协议
            # host: 127.0.0.1 # 主机地址
          initialDelaySeconds: 30 # 表明第一次检测在容器启动后多长时间后开始
          timeoutSeconds: 5 # 检测的超时时间
          periodSeconds: 30 # 检查间隔时间
          successThreshold: 1 # 成功门槛
          failureThreshold: 5 # 失败门槛，连接失败5次，pod杀掉，重启一个新的pod
        readinessProbe: # Pod 准备服务健康检查设置
          httpGet:
            path: /healthz
            port: 80
            scheme: HTTP
          initialDelaySeconds: 30
          timeoutSeconds: 5
          periodSeconds: 10
          successThreshold: 1
          failureThreshold: 5
        ports:
          - name: http # 名称
            containerPort: 80 # 容器开发对外的端口 
            protocol: TCP # 协议

```







service的yaml文件：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: httpserver-svc
  namespace: default
  labels:
    app: httpserver
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: http
      protocol: TCP
      name: http
  selector:
    app: httpserver

```





配置和日志等级放到ConfigMap中加载：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-configs
  namespace: default
data:
  appsettings: 
   ........配置信息

```



在Pod中药挂载该CM：

```yaml
containers:
  volumeMounts:
  - name: settings
    mountPath: /etc/app.config
    subPath: appsettings


volumes:
  - name: settings
    configMap:
      name: app-configs
```









#### 2. **第二部分**

除了将 httpServer 应用优雅的运行在 Kubernetes 之上，我们还应该考虑如何将服务发布给对内和对外的调用方。
来尝试用 Service, Ingress 将你的服务发布给集群外部的调用方吧。
在第一部分的基础上提供更加完备的部署 spec，包括（不限于）：

- Service
- Ingress

可以考虑的细节

- 如何确保整个应用的高可用。
- 如何通过证书保证 httpServer 的通讯安全。





Service 可以通过NodePort 方式暴露Node端口给集群外部使用，比较快速，默认端口是30000 ~ 33000 的随机端口：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: httpserver-svc
spec:
        #type: ClusterIP
  type: NodePort
  ports:
    - port: 80
      targetPort: 80
      protocol: TCP
  selector:
    app: httpserver




# 效果
root@master01:~# k get svc
NAME             TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
httpserver-svc   NodePort    10.97.25.189   <none>        80:30854/TCP   3d3h

```





Ingress：

```bash
#通过 helm 安装
helm repo add nginx-stable https://helm.nginx.com/stable
helm install ingress-nginx nginx-stable/nginx-ingress --create-namespace --namespace ingress

#安装完成后，ingress 的namespace 中可以看到对应资源
root@master01:~# k get pods -n ingress
NAME                                          READY   STATUS    RESTARTS   AGE
ingress-nginx-nginx-ingress-bcc75ccf5-7cfcp   1/1     Running   0          45h
root@master01:~# k get svc -n ingress
NAME                          TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
ingress-nginx-nginx-ingress   LoadBalancer   10.96.249.132   <pending>     80:32306/TCP,443:30658/TCP   45h
```







```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: httpserver
spec:
  ingressClassName: nginx
  rules:
    - host: www.abc.com
      http:
        paths:
          - backend:
              service:
                name: httpserver-svc
                port:
                  number: 80
            path: /
            pathType: Prefix
```



没有域名则需要在本机器hosts中手动指定域名解析。



https证书通过cert-manager管理：

```bash
#安装
helm repo add jetstack https://charts.jetstack.io
helm repo update
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.crds.yaml
helm install \
    cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --create-namespace \
    --version v1.7.1 \
```

















