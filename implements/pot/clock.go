/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/4/20 4:40 PM
* @Description: 时钟
***********************************************************************/

package pot

import (
	"errors"
	"fmt"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
)

// Moment 时刻
type Moment struct {
	Type MomentType
	Time time.Time
}

func (m Moment) String() string {
	return fmt.Sprintf("%s %s\n", m.Type.String(), m.Time.Format(time.RFC3339Nano))
}

// MomentType 时刻
type MomentType uint8

const (
	// MomentType_PotStart 竞争开始时刻
	MomentType_PotStart MomentType = 0
	// MomentType_PotOver 竞争结束时刻
	MomentType_PotOver MomentType = 1
)

func (mt MomentType) String() string {
	switch mt {
	case MomentType_PotStart:
		return "PotStart"
	case MomentType_PotOver:
		return "PotOver"
	default:
		return "Unknown"
	}
}

// bn -> [recv bn] -> decide bn / trigger clock / broadcast proof
// -> decide bn1 maker / make / wait -> bn1 -> ...

// 两个区块之间的时间被均匀分成两段，每段TickMs

// Clock 时钟
// 使用：
//		bb := baseBlock
// 		c := StartClock(bb)
//		for {
//			select {
//			case <-c.Tick():
//				do something
//			}
//		}
//
// 其实Clock完全可以只使用time.Ticker来做，但问题是由于各个节点使用的基准区块不一定相同
// 构造区块消耗的时间也不容忽视且不相同，所以为了尽量保证所有节点时间上的一致
// 每次都要用新区块来矫正时间
type Clock struct {
	t       *time.Timer
	done    chan struct{}
	trigger chan int64  // unixnano timestamp
	Tick    chan Moment // 对外的tick
	epoch   int64
}

// 接收到网络中最新的区块后，以该区块为起始驱动时钟运行
//func StartClock(baseBlock *defines.Block) *Clock {
//
//	// bb_create -> bb_recv
//	// 这里可能bb(baseBlock)收到时已经过去若干个TickMs了
//	bb_create := baseBlock.Timestamp
//	rest := TickMs - ((time.Now().UnixNano() - int64(bb_create)) >> 6) % TickMs
//
//	c := &Clock{
//		t:       time.NewTimer(time.Duration(rest) * time.Millisecond),
//		done:    make(chan struct{}),
//		trigger: make(chan uint64),
//		tick:    make(chan time.Time),
//	}
//	go c.loop()
//	return c
//}

// StartClock 接收到网络中最新的区块后，以该区块为起始驱动时钟运行
func StartClock(baseBlock *defines.Block) *Clock {
	timer := time.NewTimer(0)
	<-timer.C
	c := &Clock{
		t:       timer,
		done:    make(chan struct{}),
		trigger: make(chan int64),
		Tick:    make(chan Moment),
		epoch:   baseBlock.Index,
	}
	go c.loop()
	if err := c.Trigger(baseBlock); err != nil {
		return nil
	}
	return c
}

// Trigger 用新区块来驱动Clock
// b 是当前用于触发时钟的区块，通常是最新区块
// epoch为当前的纪元，根据epoch可以判断b是否是最新的区块，是则继续，不是则返回错误
// Trigger必须是两个Tick的时间才能调一次
// 调Trigger时传入的block需要是当时“被确认”的新区块
// 正常情况下：
//		clock_tick1: decide bn / pot -> clock_tick2: win/lose
//
// 正常情况下，总是将最新区块传入
func (c *Clock) Trigger(b *defines.Block) error {
	//if b == nil {
	//	c.trigger <- 0	// 特殊情况，说明当前轮重新开始了
	//} else {
	//	c.trigger <- b.Timestamp
	//}

	if b.Timestamp >= time.Now().UnixNano() {
		return errors.New("fatal block timestamp")
	}
	c.trigger <- b.Timestamp
	return nil
}

// loop 时钟运转循环
// 每确定一个新区块，需要驱动出两个时刻信号(PotOver(n) PotStart(n+1))
func (c *Clock) loop() {

	unit := int64(TickMs * time.Millisecond)
	divisor := 2 * unit

	for {
		select {
		case <-c.done:
			return
		case bt := <-c.trigger:
			delta := divisor - (time.Now().UnixNano()-int64(bt))%divisor
			fmt.Printf("delta: %d\n", delta/1e6)
			//fmt.Println("now: ", time.Now())
			var phase1, phase2 time.Duration
			if delta < unit {
				phase1 = time.Duration(delta)
				phase2 = time.Duration(unit)
				fmt.Println(phase1.Milliseconds(), phase2.Milliseconds())
				//fmt.Println("1111", time.Now())
				c.t.Reset(phase1)
				//fmt.Println("2222", time.Now())
				c.Tick <- Moment{
					Type: MomentType_PotOver,
					Time: <-c.t.C,
				} // 对外传递1次Tick
				//fmt.Println("3333", time.Now())
				c.t.Reset(phase2)
				c.Tick <- Moment{
					Type: MomentType_PotStart,
					Time: <-c.t.C,
				} // 对外传递1次Tick
				//fmt.Println("4444", time.Now())
			} else {
				phase1 = time.Duration(delta - unit)
				c.t.Reset(phase1)
				c.Tick <- Moment{
					Type: MomentType_PotStart,
					Time: <-c.t.C,
				} // 对外传递1次Tick
			}
		}
	}
}

// Close 关闭时钟
func (c *Clock) Close() {
	close(c.done)
	c.t.Stop()
}

// 每次都用新区块时间去校正，这种做法存在一个问题：
// 处理这个校正时，会有一定的处理时间，会使得原本利用delta > unit判断
// 接下来是1次tick还是两次tick时，出现问题，因为处理时延是不确定的，虽然都在10us级

// makeblock -> trigger / decide block / trigger
// 这个过程比TickMs会稍微多一些
