package pot


import(
	"testing"
	"fmt"
)


func TestTasker(t *testing.T) {
	tasker := NewTasker()
	fmt.Println(tasker.TPQ, tasker.Map)
	tasker.RegisterTask(func(){
		fmt.Print("A")
	}, 1, 3)
	fmt.Println(tasker.TPQ, tasker.Map)
	tasker.RegisterTask(func(){
		fmt.Print("B")
	}, 3,4)
	fmt.Println(tasker.TPQ, tasker.Map, 
		tasker.Map[1].FuncsPtr, tasker.Map[3].FuncsPtr, tasker.Map[4].FuncsPtr)

	fmt.Println("====================")

	for t:=0; t<6; t++ {	// 模拟时间前进
		fmt.Println(tasker.TPQ, tasker.Map)
		fmt.Print("Do: ")
		tasker.ExecuteTask(t)
		fmt.Print("\n")
	}
}