# 哪些类型可以作为map的key

只要可比较的类型都可以作为 map 的key，除了 `slice、map、functions` 这3种类型：
- bool
- int，包括有符号和无符号整数
- float32/float64
- string
- 指针
- channel
- interface
- struct
- 只包含上述类型的数组

 [《Go 程序员面试笔试宝典》](https://golang.design/go-questions)：
> 如果是结构体，只有 hash 后的值相等以及字面值相等，才被认为是相同的 key。很多字面值相等的，hash出来的值不一定相等，比如引用。
> 
> 顺便说一句，任何类型都可以作为 value，包括 map 类型