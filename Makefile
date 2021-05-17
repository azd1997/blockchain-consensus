# 共识协议类型 pot/pow/pbft/raft
consensus = pot

node:
	# 编译节点二进制
	echo $(consensus)
docker-node:node
	# 编译节点docker镜像

localnet:node
	# 构建局域网
test-all:
	# 执行所有测试函数
test-cluster:
	# 执行集群测试函数
