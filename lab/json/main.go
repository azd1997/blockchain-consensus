package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// 测试结构体与map在json时有无区别
func main() {
	// 结构体s
	s := struct{
		A int `json:"a"`
		B string `json:"b"`
	}{A:5, B:"xxx"}

	// 将s转为map
	sm := map[string]interface{}{}
	sm["a"] = s.A 
	sm["b"] = s.B

	// 比较json.Marshal之后的区别
	data1, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data1))
	data2, err := json.Marshal(sm)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data2))

	fmt.Println(bytes.Equal(data1, data2))
}