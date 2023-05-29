# slice底层的实现原理

切片（slice）通过数组指针、长度（len）和容量（cap）3个字段的设计，实现了类似动态数组的功能。但是它本身并非动态数组，而是通过内部指针引用底层数组，所以在赋值、函数传参时，不会涉及到底层数组的数据拷贝。因为复制 `slice` 的代价很小（1个指针，2个int变量），通常在函数传参时参数类型使用T而不是 \*T（指针）：

```go
type slice struct {  
   array unsafe.Pointer  
   len   int  
   cap   int  
}
```

切片为我们封装了快速访问底层数组的能力，我们可以使用 `索引下标` 访问或更新底层数组中元素的值，切片会自动计算底层数组的地址偏移：

```go
var arr = [3]int{0, 1, 2}  
s := arr[1:]  // 使用arr底层数组
s[0] = 3      // 底层数组的偏移地址 = arr[1]，切片进行了转换
fmt.Println(arr)  // 0 3 2
fmt.Println(s)    // 3 2
```

也可以使用 `索引区间` 访问数组或者其他切片中的某一部分数据，因为是基于数组指针的操作，所以不会有内存的拷贝：

```go
var arr = [5]int{0, 1, 2, 3, 4}  
s1 := arr[:3] // 0,1,2 基于数组创建切片  
s2 := s1[:1]  // 0,1   基于切片创建切片(reslice)  
  
fmt.Println((*reflect.SliceHeader)(unsafe.Pointer(&s1))) // &{824634957824 3 5}  
fmt.Println((*reflect.SliceHeader)(unsafe.Pointer(&s2))) // &{824634957824 1 5}
```

当追加元素底层数组容量不足时，切片还会自动 `创建新的底层数组`，实现动态扩容的功能：

```go
var s1 = []int{1, 2} // 注意：不指定长度是创建的切片  
fmt.Println((*reflect.SliceHeader)(unsafe.Pointer(&s1))) // &{824634400800 2 2}

s2 := append(s1, 3, 4)  
fmt.Println((*reflect.SliceHeader)(unsafe.Pointer(&s2))) // &{824634466336 4 4}
```