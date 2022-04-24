

# 课后练习 4.2
* 启动一个 Envoy Deployment。
* 要求 Envoy 的启动配置从外部的配置文件 Mount 进 Pod。
* 进入 Pod 查看 Envoy 进程和配置。
* 更改配置的监听端口并测试访问入口的变化。
* 通过非级联删除的方法逐个删除对象。

## Envoy Deployment 配置
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: envoy
  name: envoy
spec:
  replicas: 1
  selector:
    matchLabels:
      run: envoy
  template:
    metadata:
      labels:
        run: envoy
    spec:
      containers:
      - image: envoyproxy/envoy-dev
        name: envoy
        volumeMounts:
        - name: envoy-config
          mountPath: "/etc/envoy"
          readOnly: true
      volumes:
      - name: envoy-config
        configMap:
          name: envoy-config
```


创建 configmap，将下面的 envoy.yaml 导入到 cm 中  
`kubectl create cm envoy-config --from-file=envoy.yaml`

## Envoy 的配置文件


官方提供的默认配置导出到 envoy.yaml ，代理到 www.envoyproxy.io
```yaml
admin:
  address:
    socket_address:
      protocol: TCP
      address: 0.0.0.0
      port_value: 9901
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        protocol: TCP
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          scheme_header_transformation:
            scheme_to_overwrite: https
          stat_prefix: ingress_http
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  host_rewrite_literal: www.envoyproxy.io
                  cluster: service_envoyproxy_io
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
  clusters:
  - name: service_envoyproxy_io
    connect_timeout: 30s
    type: LOGICAL_DNS
    # Comment out the following line to test on v6 networks
    dns_lookup_family: V4_ONLY
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: service_envoyproxy_io
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: www.envoyproxy.io
                port_value: 443
    transport_socket:
      name: envoy.transport_sockets.tls
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
        sni: www.envoyproxy.io


```

## 进入 Pod 查看进程和配置文件
```bash
# 以交互式方式进入 Pod 
kubectl exec -it envoy-6958c489d9-bbtvt sh
# 查看进程
# ps -ef|grep envoy
envoy          1       0  0 02:38 ?        00:00:05 envoy -c /etc/envoy/envoy.yaml
# 查看配置文件，与 configmap 中一致
# cat /etc/envoy/envoy.yaml     
admin:
  address:
    socket_address:
      protocol: TCP
... 省略
```

## 修改配置文件
直接修改 configmap 中监听端口  
`kubectl edit cm envoy-config`    
将默认监听的端口 10000 和管理端口 9901 修改，过一会再查看 Pod 中文件会生效  
通过新端口访问不行时，可以在管理页面手动重载生效，无需重启 envoy 进程 

## 非级联删除
delete 时增加参数 `--cascade=orphan`
```bash
kubectl delete deploy envoy --cascade=orphan
kubectl delete rs envoy-6958c489d9   --cascade=orphan
# 删除完 deployment 和 replicaset 后，pod 仍然在运行
kubectl get pods

```