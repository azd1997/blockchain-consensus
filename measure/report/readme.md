# 监控数据上报

## 集群整体监控需求

1. 瞬时交易吞吐量 `Tr = NumberOf(tx) / Tb`
2. 平均交易吞吐量 `Tra = Total(tx) / 总运行时间`
3. 单区块交易确认时延`Td`
4. 平均交易确认时延`Tda`   
5. 单区块区块间隔`Tba`
6. 单区块区块间隔`Tba`

## 单个节点

- 获得新区块时上报“我已获得新区块”
    - 所有节点执行该项，则上层（监控器层面）可感知集群区块产生间隔`Tb`以及交易吞吐量`Tr`
    - 交易确认时延可以从区块中解析出交易列表，交易确认时延=区块构造时间-交易构造时间

## report报告端与监控端

监控端目前有monitor和perf95两种，这两种对于reporter来讲都称作collector，收集器
monitor收集数据后计算并交给echartsRunner
perf95收集数据、计算后调整测试参数，直至达到测试目的

monitor ( collector -> calculator ) ---> echarts    : EchartsMonitor
                                    ---> parameter  : Perf95