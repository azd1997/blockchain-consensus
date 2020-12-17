/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/14/20 11:32 AM
* @Description: 测试用的最简单的区块链，内存中维护数据
***********************************************************************/

package test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/azd1997/blockchain-consensus/utils/log"

	"github.com/azd1997/blockchain-consensus/defines"
)

const (
	Module_Bcl = "BCL"
	//initialBlocksLength = 10
)

var (
	// 这个错误用于提示上层
	ErrWrongChain = errors.New("wrong blocks in chain")
)

// BlockSegment 区块分段
type BlockSegment struct {
	start, end int64
	blocks *[]*defines.Block
}

func NewBlockChain(id string) *BlockChain {

	logger := log.NewLogger(Module_Bcl, id)
	if logger == nil {
		return nil
	}

	firstSeg := make([]*defines.Block, 0)

	return &BlockChain{
		id:id,
		chain: []*BlockSegment{
			&BlockSegment{	// 创建第1个分段
				start:  0,
				end:    0,
				blocks: &firstSeg,
			},
		},
		blocks:        &firstSeg,
		indexes:       map[string]int64{},
		discontinuous: map[int64]*defines.Block{},
		txpool:        map[string]*defines.Transaction{},
		txin:          make(chan *defines.Transaction, 100),
		Logger:logger,
	}
}

// BlockChain 渐进式地增加区块，不允许制造空洞
type BlockChain struct {
	id string

	// 该标志标记是否添加过最新区块
	addnew bool

	chain []*BlockSegment
	//blocks map[string]*defines.Block

	// [0,1,2,3]
	// [10,11]	// 10是启动时收到的最新的区块
	// [20, 21]
	// 一旦产生了多个分段，也就是链还未连续，由于只信任每个分段的第1个区块(通过多数法选出的)
	// 所以所有空缺的补全都必须从可信任的区块开始倒序填充

	// maxIndex记录的是本地曾经达到的最大高度，在RLB阶段完成之后，maxIndex其实就是网络区块链的最大索引了
	// 当本地区块链遇到错误需要删除一部分区块分段时，maxIndex并不回缩
	maxIndex int64

	// blocks是chain的第一个分段。使用blocks的前提是chain只有1个分段
	blocks  *[]*defines.Block
	indexes map[string]int64 // <hash, index>

	// 每次遇到不能直接追加在blocks末尾的时候先加入到discon。 每次加入discon的时候都检查下blocks末尾是否可以从discon取区块
	discontinuous map[int64]*defines.Block // 不连续的区块

	txpool map[string]*defines.Transaction

	txin chan *defines.Transaction

	// 自己的进度是否存在空洞？空洞情况
	// “不重叠区间” 按left升序
	//holes [][2]int64 // 空洞，还欠缺的区块区间 [2]uint64{left, right}，[left, right]这个区间的区块还没有同步到


	*log.Logger
}

// Display 展示进度
func (bc *BlockChain) Display() string {
	display := fmt.Sprintf("\nBlockChain(%s)(maxIndex=%d):\n", bc.id, bc.maxIndex)
	for i:=0; i<len(bc.chain); i++ {
		seg := bc.chain[i]
		segStr := fmt.Sprintf("Segment%d<%d-%d>:\t", i+1, seg.start, seg.end)
		for j:=0; j<len(*seg.blocks); j++ {
			if j < len(*seg.blocks) - 1 {
				segStr += fmt.Sprintf("%s ——> ", (*seg.blocks)[j].ShortName())
			} else {
				segStr += fmt.Sprintf("%s\n", (*seg.blocks)[j].ShortName())
			}
		}
		display += segStr
	}
	display += "\n"
	return display
}

func (bc *BlockChain) ID() string {
	return bc.id
}

func (bc *BlockChain) Init() error {
	bc.Debug("BlockChain: Init. start collect tx loop")

	go bc.collectTxLoop()

	return nil
}

func (bc *BlockChain) Inited() bool {
	return true
}

// GetMaxIndex maxIndex不一定代表本地最高区块的索引，因此这里使用最新分段记录的值
func (bc *BlockChain) GetMaxIndex() int64 {
	segnum := len(bc.chain)
	lastSeg := bc.chain[segnum-1]
	localMaxIndex := lastSeg.end
	return localMaxIndex
}

func (bc *BlockChain) GetBlockByIndex(index int64) (*defines.Block, error) {
	// 检查所有分段，有则返回
	for i:=0; i<len(bc.chain); i++ {
		start, end := bc.chain[i].start, bc.chain[i].end
		if start > 0 && end > 0 && index >= start && index <= end {
			return (*(bc.chain[i].blocks))[index - start], nil
		}
	}
	return nil, fmt.Errorf("bc is discontinuous now and block(%d) is missing, try again later", index)
}

func (bc *BlockChain) GetBlocksByRange(start, count int64) ([]*defines.Block, error) {
	bc.Debugf("BlockChain: GetBlocksByRange: start=%d, count=%d\n", start, count)

	//if bc.Discontinuous() {
	//	return nil, errors.New("bc is discontinuous now, try again later")
	//}
	//// bc连续时，blocks才代表区块链


	resL, resR := int64(0), int64(0)
	left, right := int64(1), int64(len(*bc.blocks)-1)
	if start >= left && start <= right { // 正常情况
		if count == 0 { // 表示数量上的模糊查询
			//res = append(res, (*bc.blocks)[start:]...)
			resL, resR = start, right
		} else if count > 0 {
			if start+count-1 < right {
				//res = append(res, (*bc.blocks)[start:start+count]...)
				resL, resR = start, start + count - 1
			} else {
				//res = append(res, (*bc.blocks)[start:]...)
				resL, resR = start, right
			}
		} else {
			return nil, errors.New("negative count")
		}
	} else if start < 0 && (-start) <= right { // start反向有效
		r := right + 1 + start
		if count == 0 { // 表示数量上的模糊查询
			//res = append(res, (*bc.blocks)[:r+1]...)
			resL, resR = start, r
		} else if count > 0 {
			if r-count+1 > 0 {
				//res = append(res, (*bc.blocks)[r-count+1:r+1]...)
				resL, resR = r-count+1, r
			} else {
				//res = append(res, (*bc.blocks)[:r+1]...)
				resL, resR = start, r
			}
		} else {
			return nil, errors.New("negative count")
		}
	} else {
		return nil, errors.New("invalid start")
	}

	bc.Debugf("resL=%d, resR=%d", resL, resR)

	res := make([]*defines.Block, resR - resL + 1)
	existErr := false
	for i:=resL; i<=resR; i++ {
		b, err := bc.GetBlockByIndex(i)
		if err != nil {
			existErr = true
		}
		res[i-resL] = b
	}

	if existErr {
		return res, errors.New("bc is discontinuous now and some blocks are missing, try again later")
	}
	return res, nil
}

func (bc *BlockChain) GetBlocksByHashes(hashes [][]byte) ([]*defines.Block, error) {
	bc.Debugf("BlockChain: GetBlocksByHashes: hashes=%v\n", hashes)

	//if bc.Discontinuous() {
	//	return nil, errors.New("bc is discontinuous now, try again later")
	//}
	//// bc连续时，blocks才代表区块链

	res := make([]*defines.Block, len(hashes))
	existErr := false
	for i := 0; i < len(hashes); i++ {
		//k := fmt.Sprintf("%x", hashes[i])
		//idx := bc.indexes[k]
		//b := *((*bc.blocks)[idx])
		//res[i] = &b

		b, err := bc.GetBlockByHash(hashes[i])
		if err != nil {
			existErr = true
		}
		res[i] = b
	}

	if existErr {
		return nil, errors.New("bc is discontinuous now and some blocks are missing, try again later")
	}
	return res, nil
}

func (bc *BlockChain) GetBlockByHash(hash []byte) (*defines.Block, error) {
	bc.Debugf("BlockChain: GetBlocksByHash: hash=%v\n", hash)

	//if bc.Discontinuous() {
	//	return nil, errors.New("bc is discontinuous now, try again later")
	//}
	//// bc连续时，blocks才代表区块链

	k := fmt.Sprintf("%x", hash)
	idx := bc.indexes[k]
	//b := *((*bc.blocks)[idx])

	b, err := bc.GetBlockByIndex(idx)
	if err != nil {
		return nil, fmt.Errorf("bc is discontinuous now and block(%d, %s) is missing, try again later", idx, k)
	}
	return b, nil
}

// AddNewBlock 将最新的区块的添加到区块链中
// 调用方如果Add返回错误ErrWrongChan，那么重新请求最新区块
func (bc *BlockChain) AddNewBlock(nb *defines.Block) error {
	bc.Debugf("BlockChain: AddNewBlock: block=%v\n", nb)

	if nb.Index <= bc.maxIndex {
		return nil
	} else if nb.Index == bc.maxIndex + 1 {		// 刚好是下一个区块
		bc.maxIndex = nb.Index	// 更新maxIndex
		segnum := len(bc.chain)
		if segnum == 0 {
			bc.Panic("blockchain should have at least one segment!")
		}
		latestSeg := bc.chain[segnum-1]
		if latestSeg.start == 0 {
			latestSeg.start = 1
		}
		latestSeg.end++
		(*latestSeg.blocks) = append((*latestSeg.blocks), nb)	// 添加到最新的分段末尾

	} else {	// nb.Index > bc.maxIndex + 1
		// 创建新的分段
		bc.maxIndex = nb.Index	// 更新maxIndex
		newSeg := &BlockSegment{
			start:  nb.Index,
			end:    nb.Index,
			blocks: &([]*defines.Block{nb}),
		}
		bc.chain = append(bc.chain, newSeg)		// 将新区块放在新分段中，新分段添加进chain
	}

	// 添加哈希键的索引
	bc.indexes[nb.Key()] = nb.Index

	// 将addnew位置true
	bc.addnew = true

	// 每次添加新区块之后，都检查下是否可以填补空白
	if err := bc.checkDiscontinuous(); err != nil {
		return err
	}

	return nil
}

func (bc *BlockChain) AddBlock(block *defines.Block) error {
	bc.Debugf("BlockChain: AddBlock: block=%v\n", block)

	if block.Index == bc.maxIndex + 1 {
		return bc.AddNewBlock(block)
	}

	// 否则的话，直接加到discon中并检查能否填空
	bc.discontinuous[block.Index] = block
	return bc.checkDiscontinuous()


	//maxIndex := bc.GetMaxIndex()
	//if block.Index == maxIndex+1 {
	//	bc.blocks = append(bc.blocks, block)
	//	bc.indexes[block.Key()] = block.Index
	//
	//	// 当区块链连续时，还需要注意将交易池中已经被使用掉的交易给删除掉
	//	if len(bc.discontinuous) == 0 {
	//		bc.cleanTxPool(block)
	//	}
	//
	//} else if block.Index > maxIndex+1 {
	//	bc.discontinuous[block.Index] = block
	//	bc.checkDiscontinuous()
	//}

	//return nil
}

// 检查discon，如果有可以追加的区块，那么递归的取出
// 约定 每找到一个可以插到一个"可信任的区块"时，将该区块插到"可信任的区块"所在的分段前面
func (bc *BlockChain) checkDiscontinuous() error {

	for i:=len(bc.chain)-1; i>0; i++ {	// 0号分段不需要检查

		selfSeg, prevSeg := bc.chain[i], bc.chain[i-1]

		firstBlock := (*(selfSeg.blocks))[0]	// 该分段上绝对可信的区块(本分段第一个区块)
		prevLastBlock := (*(prevSeg.blocks))[prevSeg.end - prevSeg.start]	// 上一分段的最后一个区块

		prevIndex := firstBlock.Index - 1
		for bc.discontinuous[prevIndex] != nil &&
			bytes.Equal(bc.discontinuous[prevIndex].SelfHash, firstBlock.PrevHash) {

			// 如果前一个区块存在，需要先更新本分段，再检查是否能与前一分段连起来
			prevBlock := bc.discontinuous[prevIndex]

			// 先将该区块加到本分段
			selfSeg.start--
			newblocks := append([]*defines.Block{prevBlock}, *(selfSeg.blocks)...)
			selfSeg.blocks = &newblocks
			bc.chain[i] = selfSeg
			// 检查更新后的本分段能否与前一个分段合并
			if prevBlock.Index == prevLastBlock.Index + 1 {
				// 满足条件，可以合并
				if bytes.Equal(prevBlock.PrevHash, prevLastBlock.SelfHash) {
					prevSeg.end = selfSeg.end
					newblocks := append(*(prevSeg.blocks), *(selfSeg.blocks)...)
					prevSeg.blocks = &newblocks
					bc.chain[i-1] = prevSeg
					bc.chain = bc.chain[:i]
				} else {	// 不满足条件，发现该区块与前一个不连贯，说明这一段出现了问题，丢弃
					bc.chain = bc.chain[:i]
					// 返回错误，上层再往上抛出，让外边重新
					return ErrWrongChain
				}
			}

			prevIndex--
		}

		if bc.discontinuous[prevIndex] != nil &&
			bytes.Equal(bc.discontinuous[prevIndex].SelfHash, firstBlock.PrevHash) {

			bc.chain = bc.chain[:i]
			// 返回错误，上层再往上抛出，让外边重新
			return ErrWrongChain
		}

	}
	return nil
}

// 收到最新区块，将本地的交易池进行清理
func (bc *BlockChain) cleanTxPool(newb *defines.Block) {
	for _, usedtx := range newb.Txs {
		k := usedtx.Key()
		if bc.txpool[k] != nil {
			delete(bc.txpool, k)
		}
	}
}

//func (bc *BlockChain) appendBlockToBlocks(b *defines.Block) {
//	bc.blocks[genesis.Index] = genesis
//	bc.indexes[genesis.Key()] = genesis.Index
//}

func (bc *BlockChain) CreateTheWorld() (genesis *defines.Block, err error) {
	bc.Debug("BlockChain: CreateTheWorld")

	if len(*bc.blocks) > 0 { // 说明已经有区块
		return nil, errors.New("non-empty blockchain")
	}

	genesis, err = defines.NewBlockAndSign(1, bc.id, nil, nil)
	if err != nil {
		return nil, err
	}
	*bc.blocks = append(*bc.blocks, genesis)
	bc.indexes[genesis.Key()] = genesis.Index
	bc.maxIndex = 1

	seg0 := bc.chain[0]
	seg0.start, seg0.end = 1, 1

	return genesis, nil
}

func (bc *BlockChain) GenNextBlock() (*defines.Block, error) {
	bc.Debugf("BlockChain: GenNextBlock: next=%d\n", bc.GetMaxIndex()+1)

	if bc.Discontinuous() {
		return nil, errors.New("discontinuous blockchain, fill first")
	}

	// 收集交易列表
	txsmp := map[string]*defines.Transaction{}
	for k := range bc.txpool {
		if txsmp[k] == nil {
			txsmp[k] = bc.txpool[k]
		}
	}
	txs := make([]*defines.Transaction, len(txsmp))
	idx := 0
	for k := range txsmp {
		txs[idx] = txsmp[k]
		idx++
	}

	maxIndex := bc.GetMaxIndex()
	nextb, err := defines.NewBlockAndSign(maxIndex+1, bc.id, (*bc.blocks)[maxIndex].SelfHash, txs)
	if err != nil {
		return nil, err
	}

	return nextb, nil
}

func (bc *BlockChain) TxInChan() chan *defines.Transaction {
	return bc.txin
}

func (bc *BlockChain) collectTxLoop() {
	for tx := range bc.txin {
		bc.Debugf("collectTxLoop: recv a tx: %s\n", tx.Key())
		bc.addTx(tx)
	}
}

func (bc *BlockChain) addTx(tx *defines.Transaction) error {
	bc.Debugf("BlockChain: AddTransaction: tx=%v\n", tx)

	k := tx.Key()

	// 如果交易没被使用过，并且有效，就可以进行接下来的步骤
	// TODO

	// 添加到ubtxp
	if bc.txpool[k] == nil {
		bc.txpool[k] = tx
	}

	return nil
}

func (bc *BlockChain) Discontinuous() bool {
	return len(bc.discontinuous) > 0 && len(bc.chain) == 1
}

// 获取最新的区块(这里的最新的指的是网络最新)
func (bc *BlockChain) GetLatestBlock() *defines.Block {
	if !bc.addnew || bc.maxIndex > bc.GetMaxIndex() {
		return nil
	}

	segnum := len(bc.chain)
	lastSeg := bc.chain[segnum-1]
	localMaxIndexBlock := (*lastSeg.blocks)[lastSeg.end - lastSeg.start]
	return localMaxIndexBlock
}