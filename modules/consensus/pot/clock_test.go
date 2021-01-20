/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/4/20 10:09 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"testing"
	"time"
)

var testClockUser = testClockUser2

func testClockUser1(t *testing.T, c *Clock, zero time.Time) {
	flag := 0
	for {
		select {
		case curT := <-c.Tick:
			if flag%2 == 0 {
				fmt.Println("pot over, win or lose, make block or wait block ", time.Now().Sub(zero).Milliseconds())
			} else {
				now := time.Now()
				t.Logf("now....: %v\n", now.Sub(zero).Milliseconds())
				err := c.Trigger(&defines.Block{
					Timestamp: now.UnixNano() - TickMs*int64(time.Millisecond),
				})
				if err != nil {
					t.Error(err)
				}
				fmt.Println("decide block, pot start, broadcast proof ", time.Now().Sub(zero).Milliseconds())
			}
			flag++
			t.Logf("tick: %s\n", curT)
		}
	}
}

func testClockUser2(t *testing.T, c *Clock, zero time.Time) {
	now := time.Now()
	var seq = []time.Duration{0}

	for {
		select {
		case curT := <-c.Tick:
			seq = append(seq, curT.Time.Sub(now))
			fmt.Println("seq: ", seq)
		}
	}
}

// 正常情况指的是Clock已经启动，新区块是在Clock控制之下来触发的
// zero.UnixNano() - TickMs * int64(time.Millisecond)
func TestClock_NormalCase(t *testing.T) {
	c := NewClock(true)

	zero := time.Now()
	baseBlock := &defines.Block{
		Timestamp: zero.UnixNano() - TickMs*int64(time.Millisecond),
	}
	fmt.Println("recv a decided block and pot start now")
	t.Logf("start1111: %d\n", time.Now().Sub(zero).Milliseconds())
	// 启动时钟
	c.Start(baseBlock) // 区块1
	t.Logf("start: %d\n", time.Now().Sub(zero).Milliseconds())

	// 启动时钟使用方
	go testClockUser(t, c, zero)

	time.Sleep(3 * time.Second)
}

// StartCase指接收到最新区块，然后以此触发clock，启动clock
// 存在两种情况，一种是启动clock时接下里就要“decide block”，一种是接下来就要“pot over / make block” 再"decide block"
func TestClock_StartCase1(t *testing.T) {
	c := NewClock(true)

	zero := time.Now()
	baseBlock := &defines.Block{
		Timestamp: zero.UnixNano() - 1.5*TickMs*int64(time.Millisecond),
	}
	fmt.Println("recv a decided block")
	t.Logf("start1111: %d\n", time.Now().Sub(zero).Milliseconds())
	// 启动时钟
	c.Start(baseBlock) // 区块1
	t.Logf("start: %d\n", time.Now().Sub(zero).Milliseconds())

	// 启动时钟使用方
	go testClockUser(t, c, zero)

	time.Sleep(3 * time.Second)
}

// StartCase指接收到最新区块，然后以此触发clock，启动clock
// 存在两种情况，一种是启动clock时接下里就要“decide block”，一种是接下来就要“pot over / make block” 再"decide block"
func TestClock_StartCase2(t *testing.T) {
	c := NewClock(true)

	zero := time.Now()
	baseBlock := &defines.Block{
		Timestamp: zero.UnixNano() - 2.5*TickMs*int64(time.Millisecond),
	}
	fmt.Println("recv a decided block")
	t.Logf("start1111: %d\n", time.Now().Sub(zero).Milliseconds())
	// 启动时钟
	c.Start(baseBlock) // 区块1
	t.Logf("start: %d\n", time.Now().Sub(zero).Milliseconds())

	// 启动时钟使用方
	go testClockUser(t, c, zero)

	time.Sleep(3 * time.Second)
}

//////////////////////////// 测试基于loop2的时钟 ///////////////////////////////

func TestClock_disableTimeCorrect(t *testing.T) {
	c := NewClock(false)

	zero := time.Now()
	baseBlock := &defines.Block{
		Timestamp: zero.UnixNano() - 2.5*TickMs*int64(time.Millisecond),
	}
	fmt.Println("recv a decided block")
	t.Logf("start1111: %d\n", time.Now().Sub(zero).Milliseconds())
	// 启动时钟
	c.Start(baseBlock) // 区块1
	t.Logf("start: %d\n", time.Now().Sub(zero).Milliseconds())

	// 启动时钟使用方
	go testClockUser(t, c, zero)

	time.Sleep(3 * time.Second)
}
