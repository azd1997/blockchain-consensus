package calculator

import (
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/measure/common"
)

var _ Calculator = &BasicCalculator{}

func NewBasicCalculator(mdChan chan<-common.MeasureData) *BasicCalculator {
	if mdChan == nil {return nil}
	bc := &BasicCalculator{
		mdChan: mdChan,
		blockVoteMap: make(map[string]*BlockVote),
		decidedBlocks: make([]*BlockInfo, common.DefaultRangeBlockNum),	// WARN: 这个长度不要小于2，计算MeasureData的方法没有考虑到这个边界情况
	}
	// 必需确保bc.decidedBlocks所有元素都已初始化！！！
	for i:=0; i<len(bc.decidedBlocks); i++ {
		bc.decidedBlocks[i] = &BlockInfo{}
	}
	return bc
}

type BasicCalculator struct {
	decidedBlocks []*BlockInfo		// 已经确定的历史区块数量，其长度待决定，需与echarts中展示长度一致

	blockVoteMap map[string]*BlockVote	// 收集当前待确定的区块的投票
	blockVoteIndex int64	// 投票的区块索引或者说高度
	//lastDecidedBlock *defines.Block	// 上一个decidedBlock

	// 注意：这里所有的totalxxx都是自monitor启动开始计算的，而不一定是区块链集群启动开始计算

	totalTxGenNum int64	// 生成的交易数
	lastTotalTxGenNum int64	// 上一次decide区块时交易的总生成数

	totalBlockNum int64
	totalTxOutNum int64

	lastRangeTxOutNum int64
	lastRangeTxGenNum int64

	lastRangeBlockDuration int64
	lastNonnilBlockInRangeNum int64	// decidedBlocks非空的区块数量

	lastRangeTotalTxConfirmation int64

	mdChan chan<- common.MeasureData	// 将计算结果写入该chan
}

func (c *BasicCalculator) AddBlock(b *defines.Block) {
	fmt.Printf("AddBlock: %s\n", b.Key())
	fmt.Printf("================ c.blockVoteIndex=%d, b,Index=%d\n", c.blockVoteIndex, b.Index)
	if b == nil {return}
	// 检查b.Index和c.blockVoteIndex关系
	if c.blockVoteIndex == 0 {	// 这是第一次准备收集区块
		c.blockVoteIndex = b.Index
	}
	if b.Index == c.blockVoteIndex + 1 {	// 这一轮收集结束了，将blockVoteMap整理下
		// 获取本轮整个集群确定的区块，并且重置blockVoteMap，blockVoteIndex++
		decidedBlock := c.decideCurBlock()
		if decidedBlock == nil {return}		// 这种情况不会出现
		fmt.Printf("Decide Block: %s\n", decidedBlock.Key())
		// 根据decidedBlock计算MeasureData
		md := c.calculate(decidedBlock)
		// 将md写入mdChan
		if c.mdChan == nil {return}
		c.mdChan <- *md
		fmt.Printf("Send MeasureData: %v\n", md)
		// 将新区块加入到重置后的blockVoteMap
		c.blockVoteMap[b.Key()] = &BlockVote{B:b, Votes: 1}
	} else if b.Index == c.blockVoteIndex {	// 说明新加进来的区块是本轮decidedBlock的候选者
		k := b.Key()
		if blockVote, ok := c.blockVoteMap[k]; ok {
			blockVote.Votes++
		} else {
			c.blockVoteMap[k] = &BlockVote{B: b, Votes: 1}
		}
	} else {
		panic(fmt.Sprintf("c.blockVoteIndex=%d, b,Index=%d", c.blockVoteIndex, b.Index))
	}	// 其他情况不会出现，忽略
}

func (c *BasicCalculator) AddTx(tx *defines.Transaction) {
	fmt.Printf("AddTx: %s\n", tx.Key())
	c.totalTxGenNum++
}

func (c *BasicCalculator) SetMDChan(mdChan chan<- common.MeasureData) {
	c.mdChan = mdChan
}

// calculate 计算
// 注意所有Range的计算是包含：
// decidedBlocks(20) | decidedBlock(1)
// 总共21个
// 而 lastRangexxx数值其实是记录的上面这个总和 - decidedBlocks[0]
func (c *BasicCalculator) calculate(decidedBlock *defines.Block) *common.MeasureData {
	if decidedBlock == nil {return nil}

	totalTxGenNum := c.totalTxGenNum	// 备份，因为在接下来执行过程中可能c.totalTxGenNum会发生增加情况

	// 计算curBlockInfo
	curBlockInfo := c.calculateStep1_curBlockInfo(decidedBlock, totalTxGenNum)

	// 计算MeasureData。先填充一部分Num相关的数值
	md := &common.MeasureData{}
	c.calculateStep2_fillNums(md, curBlockInfo, totalTxGenNum)

	// 计算RangeTxOutNum并更新c.lastRangeTxOutNum
	c.calculateStep3_RangeTxOutNum(md, curBlockInfo)
	// 计算RangeTxGenNum并更新c.lastRangeTxGenNum
	c.calculateStep4_RangeTxGenNum(md, curBlockInfo)

	// 计算RangeBlockDuration并更新c.lastRangeBlockDuration
	c.calculateStep5_CurAndRangeAndRangeAverageBlockDuration(md, curBlockInfo)
	// 计算CurAverageTxThroughput和RangeAverageTxThroughput
	c.calculateStep6_CurAverageAndRangeAverageTxThroughput(md, curBlockInfo)
	// 计算CurAverageTxConfirmation和RangeAverageTxConfirmation
	c.calculateStep7_CurAverageAndRangeAverageTxConfirmation(md, curBlockInfo)
	// 计算
	c.calculateStep8_CurAndRangeTxOutInRatio(md)

	// 最后将decidedBlocks更新
	c.decidedBlocks = append(c.decidedBlocks[1:], curBlockInfo)

	return md
}

// 计算curBlockInfo并更新c.lastTotalTxGenNum
func (c *BasicCalculator) calculateStep1_curBlockInfo(decidedBlock *defines.Block, totalTxGenNum int64) *BlockInfo {
	totalTxConfirmation := int64(0)
	for _, tx := range decidedBlock.Txs {
		totalTxConfirmation += (decidedBlock.Timestamp - tx.Timestamp)
	}
	curBlockInfo := &BlockInfo{
		B:decidedBlock,
		BlockDuration: 0,
		TxOutNum: int64(len(decidedBlock.Txs)),
		TxGenNum: totalTxGenNum - c.lastTotalTxGenNum,
		TotalTxConfirmation: totalTxConfirmation,
	}

	lastBlockInfo := c.decidedBlocks[len(c.decidedBlocks)-1]	// 取最后一个区块
	if lastBlockInfo != nil && lastBlockInfo.B != nil {
		curBlockInfo.BlockDuration = curBlockInfo.B.Timestamp - lastBlockInfo.B.Timestamp
	}

	// 更新c.lastTotalTxGenNum
	c.lastTotalTxGenNum = totalTxGenNum

	return curBlockInfo
}

func (c *BasicCalculator) calculateStep2_fillNums(md *common.MeasureData, curBlockInfo *BlockInfo, totalTxGenNum int64) {
	md.BlockTime = curBlockInfo.B.Timestamp		// 当前区块构造时间
	md.BlockKey = curBlockInfo.B.Key()

	// 总区块数及其更新
	md.TotalBlockNum = c.totalBlockNum + 1	// 总区块数
	c.totalBlockNum++

	md.TotalTxGenNum = totalTxGenNum			// 总交易生成数

	// 总交易输出数及其更新
	md.TotalTxOutNum = c.totalTxOutNum + curBlockInfo.TxOutNum		// 总交易确认数
	c.totalTxOutNum = md.TotalTxOutNum

	md.CurTxGenNum = curBlockInfo.TxGenNum			// 当前轮交易生成数
	md.CurTxOutNum = curBlockInfo.TxOutNum			// 当前轮交易确认数
}

// 计算Range相关的数值时一定要注意为nil的情况

// 计算RangeTxOutNum并更新c.lastRangeTxOutNum
// RangeTxOutNum 取的是total( decidedBlocks[1:] ) + decidedBlock
// c.lastRangeTxOutNum 表示的就是total( decidedBlocks[1:] )
func (c *BasicCalculator) calculateStep3_RangeTxOutNum(md *common.MeasureData, curBlockInfo *BlockInfo) {
	md.RangeTxOutNum = c.lastRangeTxOutNum + curBlockInfo.TxOutNum
	// 更新c.lastRangeTxOutNum
	head := int64(0)
	if len(c.decidedBlocks) >1 && c.decidedBlocks[1] != nil {
		head = c.decidedBlocks[1].TxOutNum
	}
	c.lastRangeTxOutNum += (curBlockInfo.TxOutNum - head)
}

// 计算RangeTxGenNum并更新c.lastRangeTxGenNum
// RangeTxGenNum 取的是total( decidedBlocks[1:] ) + decidedBlock
// c.lastRangeTxGenNum 表示的就是total( decidedBlocks[1:] )
func (c *BasicCalculator) calculateStep4_RangeTxGenNum(md *common.MeasureData, curBlockInfo *BlockInfo) {
	md.RangeTxGenNum = c.lastRangeTxGenNum + curBlockInfo.TxGenNum
	// 更新c.lastRangeTxGenNum
	head := int64(0)
	if len(c.decidedBlocks) >1 && c.decidedBlocks[1] != nil {
		head = c.decidedBlocks[1].TxGenNum
	}
	c.lastRangeTxGenNum += (curBlockInfo.TxGenNum - head)
}

// 计算CurBlockDuration和RangeBlockDuration以及RangeAverageBlockDuration并更新c.lastRangeBlockDuration和c.lastNonnilBlockInRangeNum
// RangeBlockDuration 取的是total( decidedBlocks[1:] ) + decidedBlock
// c.lastRangeBlockDuration 表示的就是total( decidedBlocks[1:] )
func (c *BasicCalculator) calculateStep5_CurAndRangeAndRangeAverageBlockDuration(md *common.MeasureData, curBlockInfo *BlockInfo) {
	// 计算md.RangeB(实际Range中所统计的区块数)
	if c.lastNonnilBlockInRangeNum < int64(len(c.decidedBlocks)) {
		md.RangeBlockNum = c.lastNonnilBlockInRangeNum + 1	// 因为要算上新加进来的这个区块
	} else {
		md.RangeBlockNum = int64(len(c.decidedBlocks))
	}

	// 根据lastBlock是否为空处理
	lastBlockInfo := c.decidedBlocks[len(c.decidedBlocks)-1]	// 取最后一个区块
	if lastBlockInfo.B == nil {
		md.CurBlockDuration = 0
		md.RangeBlockDuration = 0
		md.RangeAverageBlockDuration = 0
	} else {
		md.CurBlockDuration = curBlockInfo.B.Timestamp - lastBlockInfo.B.Timestamp
		md.RangeBlockDuration = c.lastRangeBlockDuration + md.CurBlockDuration
		// 这个地方要注意第一个区块的BlockDuration=0这个情况
		if c.lastNonnilBlockInRangeNum < int64(len(c.decidedBlocks)) {
			md.RangeAverageBlockDuration = md.RangeBlockDuration / (md.RangeBlockNum - 1)	// 这里不用担心除数为0
		} else {
			md.RangeAverageBlockDuration = md.RangeBlockDuration / md.RangeBlockNum
		}
	}

	// 更新c.lastNonnilBlockInRangeNum，
	if c.lastNonnilBlockInRangeNum < int64(len(c.decidedBlocks)) {
		c.lastNonnilBlockInRangeNum++
	}

	// 更新c.lastRangeBlockDuration
	head := int64(0)
	if len(c.decidedBlocks) >1 {
		head = c.decidedBlocks[1].BlockDuration
	}
	c.lastRangeBlockDuration += (curBlockInfo.BlockDuration - head)
}

// 计算CurAverageTxThroughput和RangeAverageTxThroughput
// RangeAverageTxThroughput 取的是 包括本轮区块在内的n个区块总交易数 / 总Duration
func (c *BasicCalculator) calculateStep6_CurAverageAndRangeAverageTxThroughput(md *common.MeasureData, curBlockInfo *BlockInfo) {
	// 根据lastBlock是否为空处理（为空的话，一部分数值需要填为0）
	lastBlockInfo := c.decidedBlocks[len(c.decidedBlocks)-1]	// 取最后一个区块
	if lastBlockInfo.B == nil {
		md.CurAverageTxThroughput = 0
		md.RangeAverageTxThroughput = 0
	} else {
		md.CurAverageTxThroughput = float64(curBlockInfo.TxOutNum) / float64(md.CurBlockDuration)	// 这里也不用担心除数为0
		// 由于牵扯到RangeBlockDuration，如果实际RangeBlockNum范围的第一个区块其BlockDuration为0，应排除
		// 假设len(c.decidedBlocks)=5， md.RangeBlockNum（含新区块）（1，2，3，4，5）
		// [b1, b2, b3, b4, b5] b6		------   md.RangeBlockNum=5
		//      head
		// [-, -, b1, b2, b3] b4		------   md.RangeBlockNum=4
		//      head
		head := c.decidedBlocks[len(c.decidedBlocks) - int(md.RangeBlockNum) + 1]
		if head.BlockDuration > 0 {
			md.RangeAverageTxThroughput = float64(md.RangeTxOutNum) / float64(md.RangeBlockDuration)
		} else {	// 去掉head
			rangeTxOutNum := md.RangeTxOutNum - head.TxOutNum
			md.RangeAverageTxThroughput = float64(rangeTxOutNum) / float64(md.RangeBlockDuration)
		}

	}
}

// 计算CurAverageTxConfirmation和RangeAverageTxConfirmation
// RangeAverageTxConfirmation 取的是 包括本轮区块在内的n个区块总交易数 / 总交易数
func (c *BasicCalculator) calculateStep7_CurAverageAndRangeAverageTxConfirmation(md *common.MeasureData, curBlockInfo *BlockInfo) {
	if curBlockInfo.TxOutNum > 0 {
		md.CurAverageTxConfirmation = curBlockInfo.TotalTxConfirmation / curBlockInfo.TxOutNum
	} else {
		md.CurAverageTxConfirmation = 0		// 0表示异常
	}

	if md.RangeTxOutNum > 0 {
		md.RangeAverageTxConfirmation = (c.lastRangeTotalTxConfirmation + curBlockInfo.TotalTxConfirmation) / md.RangeTxOutNum
	} else {
		md.RangeAverageTxConfirmation = 0
	}

	// 更新c.lastRangeTotalTxConfirmation
	head := int64(0)
	if len(c.decidedBlocks) > 1 {
		head = c.decidedBlocks[1].TotalTxConfirmation
	}
	c.lastRangeTotalTxConfirmation += (curBlockInfo.TotalTxConfirmation - head)
}

// 计算CurTxOutInRatio和RangeTxOutInRatio
// RangeAverageTxConfirmation 取的是 包括本轮区块在内的n个区块总交易数 / 总交易数
func (c *BasicCalculator) calculateStep8_CurAndRangeTxOutInRatio(md *common.MeasureData) {
	md.CurTxOutInRatio = 0
	if md.CurTxGenNum > 0 {
		md.CurTxOutInRatio = float64(md.CurTxOutNum) / float64(md.CurTxGenNum)
	}

	md.RangeTxOutInRatio = 0
	if md.RangeTxGenNum > 0 {
		md.RangeTxOutInRatio = float64(md.RangeTxOutNum) / float64(md.RangeTxGenNum)
	}
}

// 确定出当前轮的区块，并将blockVoteMap重置，blockVoteIndex++
func (c *BasicCalculator) decideCurBlock() *defines.Block {
	maxBlockVote := &BlockVote{Votes: 0}
	for _, blockVote := range c.blockVoteMap {
		if blockVote.Votes > maxBlockVote.Votes {
			maxBlockVote = blockVote
		}
	}

	// 清除blockVoteMap
	c.blockVoteMap = make(map[string]*BlockVote)
	c.blockVoteIndex++

	return maxBlockVote.B
}