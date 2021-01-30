package pot

import (
	"container/heap"
)

type Tasks struct {
	DoAt     int       // “时间”
	FuncsPtr *[]func() // 任务函数数组的指针
}

type TasksPQ []int // int存的是任务执行时间

// Len is the number of elements in the collection.
func (tpq *TasksPQ) Len() int {
	return len(*tpq)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (tpq *TasksPQ) Less(i int, j int) bool {
	return (*tpq)[i] < (*tpq)[j]
}

// Swap swaps the elements with indexes i and j.
func (tpq *TasksPQ) Swap(i int, j int) {
	(*tpq)[i], (*tpq)[j] = (*tpq)[j], (*tpq)[i]
}

func (tpq *TasksPQ) Push(x interface{}) {
	*tpq = append(*tpq, x.(int))
}

func (tpq *TasksPQ) Pop() interface{} {
	length := tpq.Len()
	ret := (*tpq)[length-1]
	*tpq = (*tpq)[:length-1]
	return ret
}

func (tpq *TasksPQ) Peek() interface{} {
	length := tpq.Len()
	if length > 0 {
		return (*tpq)[0]
	}
	return nil
}

// Tasker 任务器
type Tasker struct {
	Map map[int]*Tasks // 存储所有的任务列表
	TPQ *TasksPQ       // 任务执行时间的优先队列，执行时早的先出
}

func NewTasker() *Tasker {
	tasker := &Tasker{}
	tasker.TPQ = new(TasksPQ)
	heap.Init(tasker.TPQ)
	tasker.Map = map[int]*Tasks{}
	return tasker
}

// RegisterTask(func, 1, 2)
func (tasker *Tasker) RegisterTask(f func(), doAt ...int) {
	for _, at := range doAt {

		if _, exists := tasker.Map[at]; !exists { // 不存在
			t := &Tasks{
				DoAt:     at,
				FuncsPtr: &([]func(){f}),
			}
			tasker.Map[at] = t
			heap.Push(tasker.TPQ, at)
		} else { // 存在
			t := tasker.Map[at]
			*(t.FuncsPtr) = append(*(t.FuncsPtr), f)
		}

	}
}

// 每次调用时取队列头部的那批任务执行, cur代表当前的时间
func (tasker *Tasker) ExecuteTask(cur int) {
	top := tasker.TPQ.Peek()
	//fmt.Println("top:", top)
	if top == nil {
		return
	}
	at, ok := top.(int)
	if !ok {
		return
	}

	if cur >= at && tasker.Map[at] != nil {
		//fmt.Println(2222)
		funcs := *(tasker.Map[at].FuncsPtr)
		delete(tasker.Map, at)
		heap.Pop(tasker.TPQ)
		// 执行
		for _, f := range funcs {
			//fmt.Println(1111)
			f := f
			f()
		}
		// 递归检查队列下一个元素
		tasker.ExecuteTask(cur)
	}
}
