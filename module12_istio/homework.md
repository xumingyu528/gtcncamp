# 作业要求
> 把我们的 httpserver 服务以 Istio Ingress Gateway 的形式发布出来。以下是你需要考虑的几点：  
> * 如何实现安全保证；  
> * 七层路由规则；  
> * 考虑 open tracing 的接入。  

# 实现步骤
## 安装 Istio
```bash
# 执行官网的下载脚本
curl -L https://istio.io/downloadIstio | sh -
# 解压下载文件
tar zxvf istio-1.13.2-linux-amd64.tar.gz
# 进入istio 目录，执行 install ，profile为demo
cd istio-1.13.2/
./bin/istioctl install --set profile=demo -y
k get pods -n istio-system
```


## 部署httpserver服务，通过istio gateway发布
```bash
# 创建simple namespace，应用simple和istio-specs文件到该ns
kubectl create ns simple
kubectl create -f simple.yaml -n simple
kubectl create -f istio-specs.yaml -n simple
```

### simple和istio文件内容
simple中使用之前的 httpserver 镜像，service和 Pod都使用80端口
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simple
spec:
  replicas: 1
  selector:
    matchLabels:
      app: simple
  template:
    metadata:
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "80"
      labels:
        app: simple
    spec:
      containers:
        - name: simple
          imagePullPolicy: Always
          image: xumingyu/httpserver-metrics:v0.2
          ports:
            - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: simple
spec:
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: 80
  selector:
    app: simple

```

Istio Gateway的配置文件，使用simple.abc.com域名测试，本地要修改hosts文件
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: simple
spec:
  gateways:
    - simple
  hosts:
    - simple.abc.com
  http:
    - match:
        - port: 80
      route:    #目标路由信息，会关联到simple namespace中的 simple service服务
        - destination: 
            host: simple.simple.svc.cluster.local
            port:
              number: 80
---
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: simple
spec:
  selector:
    istio: ingressgateway
  servers:
    - hosts:
        - simple.abc.com
      port:
        name: http-simple
        number: 80
        protocol: HTTP

```

```bash
# 查看相关的服务
k get svc -n istio-system
istio-ingressgateway   LoadBalancer   10.106.160.226

# 访问该svc地址
export INGRESS_IP=10.106.160.226
curl -H "Host: simple.abc.com" $INGRESS_IP/hello -v

```


## ssl 配置

```yaml
# 将ssl相关 key、crt 存入secret，修改istio-specs 配置文件中相关内容

apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: simple
spec:
  gateways:
    - simple
  hosts:
    - simple.abc.com
  http:
    - match:    # 匹配端口改为443
        - port: 443
      route:    #目标路由信息，会关联到simple namespace中的 simple service服务
        - destination: 
            host: simple.simple.svc.cluster.local
            port:
              number: 80
---
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: simple
spec:
  selector:
    istio: ingressgateway
  servers:
    - hosts:
        - simple.abc.com
      port: #修改为443，协议改为HTTPS
        name: http-simple
        number: 443     
        protocol: HTTPS
      tls:  # 增加tls部分，使用刚才增加的 secret 
        mode: SIMPLE
        credentialName: abc-secret

```

```bash
# 给 ns 增加标签，istio 会根据标签内容匹配规则
kubectl label ns simple istio-injection=enabled
# 通过https 访问
curl --resolve simple.abc.com:443:$INGRESS_IP https://simple.abc.com/healthz -v -k

```




## 七层路由规则

新增一份配置文件nginx.yaml，部署nginx，根据路由规则访问不同 后缀 跳转到对应服务上

```bash
# 应用两份yaml到simple中
kubectl apply -f nginx.yaml -n simple
kubectl apply -f istio-specs.yaml -n simple
```
Nginx的yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: 80
  selector:
    app: nginx
```

在原 istio-specs 基础上修改 VirtualService 部分
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: simple
spec:
  gateways:
    - simple
  hosts:
    - simple.abc.com
  http:
    - match:    #匹配规则为uri方式
      - uri: 
          exact: "/simple/hello"    # 精确匹配该uri
      rewrite:  # 改写uri为后端服务的真实uri
        uri: "/hello"
      route:    #目标路由信息，会关联到simple namespace中的 simple service服务
        - destination: 
            host: simple.simple.svc.cluster.local
            port:
              number: 80

    - match:    # nginx 的匹配规则
      - uri:
          prefix: "/nginx"
      rewrite:
          uri: "/"
      route:
        - destination:
            host: nginx.simple.svc.cluster.local
            port:
              number: 80
```

```bash
# 访问不同的url测试
curl -H "Host: simple.abc.com" $INGRESS_IP/simple/hello
curl -H "Host: simple.abc.com" $INGRESS_IP/nginx
```


## open tracing 接入
### 部署jaeger
详细yaml内容放到最后，用了孟老师给出的yaml

### 后端service 服务
同样httpserver 没有调用其他服务，使用孟老师repo中的service实验。  
yaml内容如下，三个类似
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: service0
spec:
  replicas: 1
  selector:
    matchLabels:
      app: service0
  template:
    metadata:
      labels:
        app: service0
    spec:
      containers:
        - name: service0
          imagePullPolicy: Always
          image: cncamp/service0:v1.0
          ports:
            - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: service0
spec:
  ports:
    - name: http-service0
      port: 80
      protocol: TCP
      targetPort: 80
  selector:
    app: service0
```

### Gateway的yaml
istio-specs 的内容，只将service0发布出去
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: service0
spec:
  gateways:
    - service0
  hosts:
    - '*'
  http:
  - match:
      - uri:
          exact: /service0
    route:
      - destination:
          host: service0
          port:
            number: 80
---
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: service0
spec:
  selector:
    istio: ingressgateway
  servers:
    - hosts:
        - '*'
      port:
        name: http-service0
        number: 80
        protocol: HTTP
```


### 测试
访问url测试，jaeger 的dashboard 中可以看到，后端服务间的调用情况





### jeager的 yaml文件

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jaeger
  namespace: istio-system
  labels:
    app: jaeger
spec:
  selector:
    matchLabels:
      app: jaeger
  template:
    metadata:
      labels:
        app: jaeger
      annotations:
        sidecar.istio.io/inject: "false"
        prometheus.io/scrape: "true"
        prometheus.io/port: "14269"
    spec:
      containers:
        - name: jaeger
          image: "docker.io/jaegertracing/all-in-one:1.23"
          env:
            - name: BADGER_EPHEMERAL
              value: "false"
            - name: SPAN_STORAGE_TYPE
              value: "badger"
            - name: BADGER_DIRECTORY_VALUE
              value: "/badger/data"
            - name: BADGER_DIRECTORY_KEY
              value: "/badger/key"
            - name: COLLECTOR_ZIPKIN_HOST_PORT
              value: ":9411"
            - name: MEMORY_MAX_TRACES
              value: "50000"
            - name: QUERY_BASE_PATH
              value: /jaeger
          livenessProbe:
            httpGet:
              path: /
              port: 14269
          readinessProbe:
            httpGet:
              path: /
              port: 14269
          volumeMounts:
            - name: data
              mountPath: /badger
          resources:
            requests:
              cpu: 10m
      volumes:
        - name: data
          emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: tracing
  namespace: istio-system
  labels:
    app: jaeger
spec:
  type: ClusterIP
  ports:
    - name: http-query
      port: 80
      protocol: TCP
      targetPort: 16686
    # Note: Change port name if you add '--query.grpc.tls.enabled=true'
    - name: grpc-query
      port: 16685
      protocol: TCP
      targetPort: 16685
  selector:
    app: jaeger
---
# Jaeger implements the Zipkin API. To support swapping out the tracing backend, we use a Service named Zipkin.
apiVersion: v1
kind: Service
metadata:
  labels:
    name: zipkin
  name: zipkin
  namespace: istio-system
spec:
  ports:
    - port: 9411
      targetPort: 9411
      name: http-query
  selector:
    app: jaeger
---
apiVersion: v1
kind: Service
metadata:
  name: jaeger-collector
  namespace: istio-system
  labels:
    app: jaeger
spec:
  type: ClusterIP
  ports:
    - name: jaeger-collector-http
      port: 14268
      targetPort: 14268
      protocol: TCP
    - name: jaeger-collector-grpc
      port: 14250
      targetPort: 14250
      protocol: TCP
    - port: 9411
      targetPort: 9411
      name: http-zipkin
  selector:
    app: jaeger
```




