package main

import "fmt"

// 类似 C++ using std。 std::xxx


type Animal interface {
	Name() string
	Color() string

	Eat(food interface{})
}

type Cat struct {
	Eye // 等同于 Eye Eye
}

func (c *Cat) Name() string {
	return "Cat"
}

func (c *Cat) Color() string {
	return "yellow"
}

func (c *Cat) Eat(food interface{}) {
	fmt.Println(food)
}

func (c *Cat) Height() int {
	return 0
}

// 接口与实现接口的结构体（类） 的关系是：结构体xx is a 接口

// has-a 组合关系

type Eye struct {
	color string
}


//func main() {
//	var c *Cat
//	//c.Eye.color // 访问成员的成员
//	//c.color // 和上面是一样的
//
//	c = new(Cat)
//	c.Eat(1)
//	c.Eat("foodsss")
//	c.Eat(true)
//}