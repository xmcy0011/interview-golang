package main

import "fmt"

type MyNum struct {
	num int
}

// exampleUpdateSlice 更新slice元素值示例
func exampleUpdateSlice() {
	// 错误：v是元素的拷贝，对齐的更改不会影响切片中的元素
	data := []int{1, 2, 3}
	for _, v := range data {
		v *= 10 // original item is not changed
	}
	fmt.Println("data:", data) // [1 2 3]

	// 正确：使用索引更新元素值
	for i := range data {
		data[i] *= 10
	}
	fmt.Println("data:", data) // [10 20 30]

	// 正确：v拷贝的是指针，对其的更改会影响结构体字段的值
	nums := []*MyNum{{1}, {2}, {3}}
	for _, v := range nums {
		v.num *= 10
	}

	fmt.Println(nums[0], nums[1], nums[2]) //prints &{10} &{20} &{30}
}

func main() {
	exampleUpdateSlice()
}
