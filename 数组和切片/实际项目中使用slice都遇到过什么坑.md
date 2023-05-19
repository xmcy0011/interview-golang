# 实际项目中使用slice都遇到过什么坑

1. 设置合理的初始化大小，避免不必要的扩容
2. for range遍历切片时，如果要更改元素的值，请注意value是元素的拷贝，当切片中元素的类型是值类型时，对它的更改不会影响切片中的元素

```go
import "fmt"

type MyNum struct {
	num int
}

func main() {
	data := []int{1, 2, 3}
	// 错误：v是元素的拷贝，int是值类型，对其的更改不会影响切片中的元素
	for _, v := range data { 
		v *= 10                 // original item is not changed
	}
	fmt.Println("data:", data)  // [1 2 3]

    // 正确：使用索引更新元素值
	for i := range data { 
		data[i] *= 10
	}
	fmt.Println("data:", data)  // [10 20 30]

    // 正确：v拷贝的是指针，v.num 和 nums[i].num 指向的是同一块内存，故赋值操作有效
    nums := []*MyNum{{1}, {2}, {3}}
    for _, v := range nums {
        v.num *= 10
    }

    fmt.Println(nums[0], nums[1], nums[2]) // &{10} &{20} &{30}
}
```

3. 使用 `copy()` 避免切片污染

```go
var arr = [5]int{0, 1, 2, 3, 4}  
s1 := arr[:3] // 0,1,2 基于数组创建切片  
s2 := s1[:1]  // 0,1   基于切片创建切片(reslice)  
s2[0] = 9  
fmt.Println(s1,s2)  // [9 1 2] [9]

var arr = [5]int{0, 1, 2, 3, 4}  
var s3 = make([]int, 1)  
copy(s3, arr[:1])  
s3[0] = 10  
fmt.Println(arr, s3) // [0, 1, 2, 3, 4] [10]
```

4. slice的内存泄漏问题（[来源：Go 语言高性能编程-大量内存得不到释放](https://geektutu.com/post/hpg-slice.html#3-1-%E5%A4%A7%E9%87%8F%E5%86%85%E5%AD%98%E5%BE%97%E4%B8%8D%E5%88%B0%E9%87%8A%E6%94%BE)）

在已有切片的基础上进行切片（`reslice`），不会创建新的底层数组。因为原来的底层数组没有发生变化，内存会一直占用，直到没有变量引用该数组。因此很可能出现这么一种情况，原切片由大量的元素构成，但是我们在原切片的基础上切片，虽然只使用了很小一段，但底层数组在内存中仍然占据了大量空间，得不到释放。比较推荐的做法，使用 `copy` 替代 `re-slice`。

```go
// bad
func lastNumsBySlice(origin []int) []int {  
	return origin[len(origin)-2:]  
}  

// good
func lastNumsByCopy(origin []int) []int {  
	result := make([]int, 2)  
	copy(result, origin[len(origin)-2:])  
	return result  
}
```
