# map 是线程安全的吗

map 不是线程安全的。

在查找、赋值、遍历、删除的过程中都会检测写标志，一旦发现写标志置位（等于1），则直接 fatal程序崩溃退出。赋值和删除函数在检测完写标志是复位之后，先将写标志位置位，才会进行之后的操作。

以 `go1.20` 为例：

赋值时（`runtime/map.go:mapassign`）：

```go
m["key"] = "value"
```

会设置写标志：

```go
// Set hashWriting after calling t.hasher, since t.hasher may panic,
// in which case we have not actually done a write.
h.flags ^= hashWriting
```

迭代遍历（`runtime/map.go:mapiternext` ）时：

```go
for v := range m {
	//...
}
```

会检测写标志：

```go
if h.flags&hashWriting != 0 {
	fatal("concurrent map iteration and map write")
}
```

如果此时发现写标志被设置，则触发`fatal` ，程序退出。