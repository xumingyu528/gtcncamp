[toc]

# 微服务架构
## 架构演变


## Sidecar 原理






## Service Mesh






# 服务网格：Istio 架构









# Envoy
## 主流7层代理对比
Envoy、NginX、HAProxy

## 优势

* 性能
  * 在具备大量特性的同时，Envoy提供极高的吞吐量和低尾部延迟差异，而 CPU 和 RAM 消耗相对较少
* 可扩展性
  * Envoy 在 L4 和 L7 都提供了丰富的可插拔过滤器能力，使用户可以轻松添加开源版本中没有的功能。
* API 可配置性
  * Envoy 提供了一组可以通过控制平面服务实现的管理 API。如果控制平面实现所有的 API，则可以使用通用引导配置在整个基础架构上运行 Envoy。所有进一步的配置更改通过管理服务器以无缝方式动态传送，因此 Envoy 从不需要重新启动。这使得 Envoy 成为通用数据平面，当它与一个足够复杂的控制平面结合时，会极大的降低整体运维的复杂性。



## Envoy 线程模式
* Envoy 采用单进程多线程模式：
  * 主线程负责协调；
  * 子线程负责监听过滤和转发。
* 当某连接被监听器接受，那么该连接的全部生命周期会与某线程绑定。
* Envoy 基于非阻塞模式（Epoll）。
* 建议 Envoy 配置的 worker 数量与 Envoy 所在的硬件线程数一致。



## Envoy 配置演示

示例：
```yaml
admin:
  address:
    socket_address: { address: 127.0.0.1, port_value: 9901 }

static_resources:
  listeners:
    - name: listener_0
      address:
        socket_address: { address: 0.0.0.0, port_value: 10000 }
      filter_chains:
        - filters:
            - name: envoy.filters.network.http_connection_manager
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                stat_prefix: ingress_http
                codec_type: AUTO
                route_config:
                  name: local_route
                  virtual_hosts:
                    - name: local_service
                      domains: ["*"]
                      routes:
                        - match: { prefix: "/" }
                          route: { cluster: some_service }
                http_filters:
                  - name: envoy.filters.http.router
  clusters:
    - name: some_service
      connect_timeout: 0.25s
      type: LOGICAL_DNS
      lb_policy: ROUND_ROBIN
      load_assignment:
        cluster_name: some_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: simple
                      port_value: 80
```




## Envoy 架构


## xDS-Envoy 的发现机制



## Envoy 的过滤器模式







# Istio 流量管理

## 流量劫持机制
为用户应用注入 Sidecar
* 自动注入
* 手动注入
  * istioctl kube-inject -f yaml/istio-bookinfo/bookinfo.yaml


注入后的结果
* 注入了 init-container istio-init
  * istio-iptables -p 15001 -z 15006 -u 1337 -m REDIRECT -i * -x -b 9080 -d 15090,15021,15020
* 注入了 sidecar container istio-proxy


## 流量转发规则




## 示例
### 规则配置







# 跟踪采样