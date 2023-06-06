# 如何实现线程安全的map

## 方法

- 方式一：使用互斥锁 `sync.Mutex` + map 实现
- 方式二：使用读写锁 `sync.RWMutex` + map 实现
- 方式三：使用标准库中的 `sync.Map` 实现（基于互斥锁sync.Mutex实现，读写分离适合读多写少的场景），具体参考 [syncMap的实现原理](9-syncMap的实现原理)
- 方式四：使用CAS技术实现无锁 HashMap，具体参考 [CAS和无锁HashMap的实现](10-CAS和无锁HashMap的实现.md)

读写锁可以降低锁的粒度，性能会优于互斥锁的实现，本节简单介绍一下使用读写锁实现线程安全的 map 的关键点。

## 读写锁实现

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

当然，每次遍历都会伴随着 key 的全量复制，在 `读多写少` 的场景下，GC的压力会大大增加，这种场景下更好的方式使用 sync.Map 代替。

sync.Map 本质上也是基于互斥锁实现，通过冗余 read 和 dirty 2个哈希表，实现了读写分离。在读多写少的场景下，遍历时不需要从 dirty 表同步数据，而是直接从缓存的 read 表中读取，故相比于上面全量复制 key 的方式，`GC的压力大大降低`。

但是如果 `写多读少`，for循环中又有耗时操作（比如发送TCP数据包），那么上面全量复制一次 key 的方式会比较好，相比 sync.Map 冗余2个 map，内存会减小一半。

如果 map 的元素较少，且 for 循环中没有耗时操作，那么 RLock 住整个 map 也可以考虑，代码简单，易于维护。

## sync.Map中的遍历

我们简单分析一下标准库 sync.Map 中 `Range` 的源码，来看一下读写分离的实现的细节。

首先来看一下 sync.Map 提供的主要方法：

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
    // 加载 read 哈希表
   if read.amended {  // 有新数据写入到 dirty 哈希表，需要同步
	  m.mu.Lock()     // 加锁并发操作 dirty，确保安全
      read = m.loadReadOnly()  
      // 从 dirty 进行数据同步：整表复制key到 read 哈希表
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
      // 确定该元素是否被删除
      // PS：遍历时，如果read指向的map中某个key被删除，因为map不支持并发读写，所以
      // 删除操作会把 key 对于的 value 设置为 nil
      v, ok := e.load() 
      if !ok {  
         continue  
      }  
      if !f(k, v) {  
         break  
      }  
   }  
}
```

- 先判断 `amended` 标志，如果为 true 代表 dirty 中的数据需要整表同步到 read 表中
- 遍历 read 表，read 是一个原子类型，值是一个指针，指向了一个 readOnly 结构
	- read.m  是一个原生map，类型为： `map[any]*entry` ，因为其不能并发读写，故要从 read 表删除数据需要把 key 对应的 entry 指针设置为 nil，所以上面的 `e.Load` 就是为了检查是否被删除了
	- 遍历时如果有新数据写入，并且其他 routine 也使用 Range 进行遍历，此时 read 的指针被指向一个新的 readOnly 结构，其 m 指向了被整表复制的 dirty map，故当前的遍历不受影响，遍历完成后，整个临时的 map 被回收

我们看到，sync.Map 使用了读写分离的思想，遍历过程中，如果元素被删除，则会直接跳过。如果新元素被添加进来，也无法访问到。