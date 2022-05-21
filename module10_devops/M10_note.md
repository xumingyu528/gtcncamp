[toc]

# 镜像仓库

## 简介

镜像仓库（Docker Registry）负责存储、管理和分发镜像。  

镜像仓库管理多个 Repository，Repository 通过命名来区分。每个 Repository 包含一个或多个镜像，镜像通过镜像名称和标签（tag）来区分。  

客户端拉取镜像时，要指定三要素：

* 镜像仓库：要从哪个镜像仓库拉取镜像，通常通过 DNS 或 IP 地址来确定一个镜像仓库，如 hub.docker.com
* Repository：组织名，如 apache
* 镜像名称+标签：如 nginx:latest



**镜像仓库遵循 OCI 的 Distribution Spec**

![](./note_images/oci_distribution_spec.png)

### 数据和块文件

镜像由元数据和块文件两部分组成，镜像仓库的核心职能就是管理这两项数据。  

* 元数据
  * 用于描述一个镜像的核心信息，包含镜像的镜像仓库、repository、标签、校验码、文件层、镜像构建描述等信息。
  * 通过这些信息可以从抽象层面完整描述一个镜像：它是如何构建出来的、运行过什么构建命令、构建的每一个文件层的校验码、打的标签、镜像的校验码等
* 块文件（blob）
  * 块文件是组成镜像的联合文件层的实体，每一个块文件是一个文件层，内部包含对应文件层的变更。



![](./note_images/metadata_blob.png)

### 公有和私有镜像仓库

公有镜像仓库优势

* 开放：任何开发者都可以上传，分享镜像到公有仓库中
* 便捷：开发者可以非常方便的搜索、拉取其它开发者镜像，避免重复构建
* 免运维：开发者只需要关注应用开发，不必关心镜像仓库的更新、升级、维护
* 成本低：企业或开发者不需要购买硬件、解决方案来搭建镜像仓库，也不需要团队维护



私有镜像仓库优势

* 隐私性：企业的代码和数据都是企业的私有财产，不允许随意共享到公共平台
* 敏感性：企业的镜像会包含一些敏感信息，例如密钥信息、令牌信息等
* 网络连通性：企业网络结构多种多样，并非所有环境都可以访问互联网
* 安全性：在企业环境中，若使用一些含有漏洞的依赖包，则会引入安全隐患



## Harbor

quar 是 CoreOS 以前开发的私有镜像仓库软件。  

目前主流的是 Harbor，VMware 开源的企业级镜像仓库，已经是 CNCF 的毕业项目。它拥有完整的仓库管理、镜像管理、基于角色的权限控制、镜像安全扫描集成、镜像签名等。  

![](./note_images/harbor.png)

* Harbor 核心服务：提供 Harbor 的核心管理服务 API，包括仓库管理、认证管理、授权管理、配置管理、项目管理、配额管理、签名管理、副本管理等
* Harbor Portal：Harbor 的 Web 界面
* Registry：负责接收客户端的 pull/push 请求，其核心为 Docker/Distribution
* Replication controller（副本控制器）：Harbor 可以以主从模式部署镜像仓库，副本控制器将镜像从主镜像服务分发到从镜像服务
* Log Collector（日志收集器）：收集各模块日志
* 垃圾回收控制器：回收日常操作中删除镜像记录后遗留在块存储中的孤立块文件



### Harbor 架构

![](./note_images/harbor_architecture.png)





### Harbor 安装









### Harbor 高可用架构

![](./note_images/harbor_ha_architecture.png)





### Harbor 用户管理

![](./note_images/harbor_user_management.png)





### 垃圾回收 Garbage Collection





### 本地镜像加速 Dragonfly

Dragonfly 是一款基于 P2P 的智能镜像和文件分发工具。  

旨在提高文件传输的效率和速率，最大限度地利用网络带宽，尤其在分发大量数据时。  

* 应用分发
* 缓存分发
* 日志分发
* 镜像分发



#### 优势





#### 镜像下载流程





# 镜像安全







# 基于 Kubernetes 的 DevOps





# 自动化流水线
## 基于GitHub Action



## 基于 Jenkins








## 基于声明式 API ：Tekton





# 持续交付 Argo CD

Argo CD 是用于 Kubernetes 的声明性 GitOps 连续交付工具。  

* 应用程序定义，配置和环境应为声明性的，并受版本控制。
* 应用程序部署和生命周期管理应该是自动化的，可审核的且易理解的



## 架构





## 安装









# 日志收集分析









## Loki















# 监控系统



## Prometheus

























