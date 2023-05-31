# 如何实现线程安全的map

- 方式一：使用读写锁 `sync.RWMutex` 实现
- 方式二：使用标准库中的 `sync.Map` 实现

使用读写锁实现的关键点在于遍历，遍历时如果 lock 整个map，则锁的粒度等于循环的次数，此时可以考虑空间换时间，通过复制 key 的方式，每次检查一下即可。

假设现在要遍历用户列表：

```go
mu := sync.RWMutex{}  
users := make(map[string]interface{})
```

最差的方式，就是 lock 住整个遍历过程：

```go
// Bad：如果 for 循环中的代码处理耗时1ms，则整个锁被持有 1ms * 元素个数  
defer mu.RUnlock()  
mu.RLock()  
for range users {  
   // you logic  
}
```

此时，如果 `for` 循环中有耗时操作，可考虑延迟 lock 方式降低锁持有的时间，防止锁饥饿：

```go
// Good：利用空间换时间，此时锁的粒度很细，不会随着 users 变长而变长  
var keys []string  

// 快速复制key
mu.RLock()  
keys = make([]string, len(users))  
for v := range users {  
   keys = append(keys, v)  
}  
mu.RUnlock()  
  
for _, v := range keys {  
   // 检测该元素是否在遍历时被其他 routine 删除
   mu.RLock()  
   _, ok := users[v]  
   mu.Unlock()  
  
   if ok {  
      // your logic  
   }  
}
```

这个思想启发自标准库 sync.Map 的实现，先看一下它提供的接口：

```go
type mapInterface interface {  
   Load(any) (any, bool)  
   Store(key, value any)  
   LoadOrStore(key, value any) (actual any, loaded bool)  
   LoadAndDelete(key any) (value any, loaded bool)  
   Delete(any)  
   Swap(key, value any) (previous any, loaded bool)  
   CompareAndSwap(key, old, new any) (swapped bool)  
   CompareAndDelete(key, old any) (deleted bool)  
   Range(func(key, value any) (shouldContinue bool))  
}
```

其中 `Range` 是用来进行遍历的，其中关键代码如下：

```go
func (m *Map) Range(f func(key, value any) bool) {  
    // 复制 key
   if read.amended {  
      read = m.loadReadOnly()  
      if read.amended {  
         read = readOnly{m: m.dirty}  
         m.read.Store(&read)  
         m.dirty = nil  
         m.misses = 0  
      }  
      m.mu.Unlock()  
   }  

   // 遍历 key 
   for k, e := range read.m {  
      v, ok := e.load() // 查看该元素是否存在，使用了读写锁
      if !ok {  
         continue  
      }  
      if !f(k, v) {  
         break  
      }  
   }  
}
```

我们看到它的思想类似，这样无论 `f` 回调耗时多久，遍历时都不会导致整个map被锁住，其他 routine 无法写入了。