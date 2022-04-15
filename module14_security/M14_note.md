[toc]

# 云原生语境下的安全保证




# 容器运行时和 Kubernetes 的安全保
## 以Non-root身份运行容器
在 Dockerfile 中通过 USER 命令切换成非 root 用户。

```dockerfile
FROM ubuntu
RUN user add patrick
USER patrick
```



## 集群的安全性保证




## 存储加密


## 示例
### Security Context


### Pod Security Policies



## 为节点增加 Taint




# 网络策略 NetworkPolicy
## 概念


## 基于Calico的NetworkPolicy










# 零信任架构（ZTA）







# 基于 Istio 的安全保证






# Istio 认证机制的原理与实现





# Istio 鉴权机制的原理与实现





