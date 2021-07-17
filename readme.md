# blockchain-consensus 

- 定位：网络/共识状态机 耦合的完整状态机
- 本地存储允许自定义使用，传入此库

## Makefile用法

### 准备好make环境

- 建议windows用户下载Mingw64，将gnu-make32改名为make并置于环境变量PATH之中，
- 使用git-bash作为命令执行环境

### 配置Makefile变量

### 编译节点二进制

```shell
make node
```

### 构建docker镜像

```shell
make docker-node
```

### 构建localnet

```shell
make localnet
```

## 整体架构

## 共识模块

## 网络模块

## 节点表模块

## 账本模块

## 日志模块

## 数据监控模块

> 使用节点自身日志不便于统计集群状态

因此，监控需要额外做一套，各个节点自己将数据报告，由一个单独的进程或者机器对节点上报的数据进行汇总。有两种汇总方式：
- 方案1：节点将监控数据通过网络协议如websocket上传给一个单独的监控进程monitor，由monitor负责将这些数据持久化到数据库
- 方案2：节点将监控数据直接写到数据库，由可选的监控进程自己读数据库取监控数据

当前亟需监控数据绘图，因此，暂时选择方案2.

## 下一步计划

- 增加其他共识模块实现
- Ledger增加磁盘存储
- 节点与集群支持docker
- 增加控制器Controller，控制若干节点进行某些修改。 Controller作为TCP客户端，连接节点的TCP服务端
