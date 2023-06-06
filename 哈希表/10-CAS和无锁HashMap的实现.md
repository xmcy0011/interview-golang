# CAS和无锁HashMap的实现

## 原子操作

### 定义

> 来源：[原子操作](https://baike.baidu.com/item/%E5%8E%9F%E5%AD%90%E6%93%8D%E4%BD%9C/1880992)
> 所谓原子操作是指不会被线程调度机制打断的操作；这种操作一旦开始，就一直运行到结束，中间不会有任何 context switch [1] （切换到另一个线程）
> 
> 原子操作可以是一个步骤，也可以是多个操作步骤，但是其顺序不可以被打乱，也不可以被切割而只执行其中的一部分。
> 
> 将整个操作视作一个整体是[原子性](https://baike.baidu.com/item/%E5%8E%9F%E5%AD%90%E6%80%A7/7760668?fromModule=lemma_inlink)的[核心特征](https://baike.baidu.com/item/%E6%A0%B8%E5%BF%83%E7%89%B9%E5%BE%81/14082029?fromModule=lemma_inlink)。

### i++的问题

简单看一个示例，下面的代码在并发环境下有什么问题？

```go
i++
```

我们拆开这个语句，实际上是对应了2个计算机指令：

1. 先从内存中读取 i 
2. 把 i 的值设置为 i + 1

如果同时有多个线程执行这个操作，因为 CPU 是抢占式的，当线程 A 执行到第一步时，可能CPU被分配给其他线程运行了 i+1 的操作，然后又切换回来线程A继续执行 i+1，这个时候我们会发现 i 的值回退了！

这就是所谓的执行顺序问题，为了解决这个问题，我们可以使用 互斥锁 确保同时只有一个线程执行这个操作，也就没有争抢问题了。

当然，现在几乎所有的 CPU 都提供了特定指令，只需要一次系统调用，就可以完成上面2件事情（先读取i的指，然后递增n，一个cpu指令完成），比如 X86 下 CAS（compare and swap）操作对应的是 CMPXCHG 汇编指令。

### 有哪些原子操作

维基百科中（[Linearizabilit](https://en.wikipedia.org/wiki/Linearizability) ）列出了以下几种原子操作：
- atomic read-write
- atomic swap (the RDLK instruction in some Burroughs mainframes, and the XCHG x86 instruction)
- [test-and-set](https://en.wikipedia.org/wiki/Test-and-set)
- [fetch-and-add](https://en.wikipedia.org/wiki/Fetch-and-add)
- [compare-and-swap](https://en.wikipedia.org/wiki/Compare-and-swap)
- [load-link/store-conditional](https://en.wikipedia.org/wiki/Load-link/store-conditional)

### Go中的原子操作

在Go语言中，`atomic` 包提供了 atomic read-write 、atomic swap 和 compare-and-swap 的支持：

```go
// /src/sync/atomic/doc.go

// LoadInt32 atomically loads *addr.// Consider using the more ergonomic and less error-prone [Int32.Load] instead.  
func LoadInt32(addr *int32) (val int32)

// StoreInt32 atomically stores val into *addr.// Consider using the more ergonomic and less error-prone [Int32.Store] instead.  
func StoreInt32(addr *int32, val int32)

// AddInt32 atomically adds delta to *addr and returns the new value.// Consider using the more ergonomic and less error-prone [Int32.Add] instead.  
func AddInt32(addr *int32, delta int32) (new int32)

// SwapInt32 atomically stores new into *addr and returns the previous *addr value.// Consider using the more ergonomic and less error-prone [Int32.Swap] instead.  
func SwapInt32(addr *int32, new int32) (old int32)

// CompareAndSwapInt32 executes the compare-and-swap operation for an int32 value.// Consider using the more ergonomic and less error-prone [Int32.CompareAndSwap] instead.  
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)
```

以 `i++` 为例，原子操作中的实现为：

```go
package main  
  
import (  
   "fmt"  
   "sync/atomic"   "time")  
  
func main() {  
   var i int32 = 0  
  
   for j := 0; j < 3; j++ {  
      go func() {  
         for k := 0; k < 10; k++ {  
            atomic.AddInt32(&i, 1)  
         }  
      }()  
   }  
  
   time.Sleep(time.Second)  
   fmt.Print(atomic.LoadInt32(&i))  
}
```

执行后输出：

```bash
30
```

我们看到，没有使用互斥锁，同时启动 3个协程分别进行10次 i++ 操作，最后正确的输出30，并且无论执行多少次，都是输出的30！

## CAS

### 什么是CAS

> 本节来源于：https://coolshell.cn/articles/8239.html

CAS是 `原子操作` 的一种，中文翻译为比较和交换（compare and swap），这个操作用C语言来描述就是下面这个样子：

```c
int compare_and_swap (int* reg, int oldval, int newval)
{
  int old_reg_val = *reg;
  if (old_reg_val == oldval) {
     *reg = newval;
  }
  return old_reg_val;
}
```

意思就是说，看一看内存 *reg 里的值是不是 oldval，如果是的话，则对其赋值 newval。

我们可以看到，old_reg_val 总是返回，于是，我们可以在 compare_and_swap 操作之后对其进行测试，以查看它是否与 oldval相匹配，因为它可能有所不同，这意味着另一个并发线程已成功地竞争到 compare_and_swap 并成功将 reg 值从 oldval 更改为别的值了。

这个操作可以变种为返回bool值的形式（返回 bool值的好处在于，可以调用者知道有没有更新成功）：

```c
bool compare_and_swap (int *addr, int oldval, int newval)
{
  if ( *addr != oldval ) {
      return false;
  }
  *addr = newval;
  return true;
}
```

### ABA问题

所谓ABA（[见维基百科的ABA词条](https://en.wikipedia.org/wiki/ABA_problem)），问题基本是这个样子：

- 进程P1在共享变量中读到值为A
- P1被抢占了，进程P2执行
- P2把共享变量里的值从A改成了B，再改回到A，此时被P1抢占。
- P1回来看到共享变量里的值没有被改变，于是继续执行。

虽然 P1 以为变量值没有改变，继续执行了，但是这个会引发一些潜在的问题。ABA问题最容易发生在 lock free 的算法中的，CAS首当其冲，因为CAS判断的是指针的值。`很明显，值是很容易又变成原样的`。

比如下面的 DeQueue() 函数：

```c++
DeQueue(Q) //出队列，改进版
{
    while(TRUE) {
        //取出头指针，尾指针，和第一个元素的指针
        head = Q->head;
        tail = Q->tail;
        next = head->next;
        // Q->head 指针已移动，重新取 head指针
        if ( head != Q->head ) continue;
        
        // 如果是空队列
        if ( head == tail && next == NULL ) {
            return ERR_EMPTY_QUEUE;
        }
        
        //如果 tail 指针落后了
        if ( head == tail && next == NULL ) {
            CAS(Q->tail, tail, next);
            continue;
        }
        //移动 head 指针成功后，取出数据
        if ( CAS( Q->head, head, next) == TRUE){
            value = next->value;
            break;
        }
    }
    free(head); //释放老的dummy结点
    return value;
}
```

因为我们要让 head 和 tail 分开，所以我们引入了一个 dummy 指针给 head，当我们做CAS的之前，如果head的那块内存被回收并被重用了，而重用的内存又被 EnQueue() 进来了，这会有很大的问题。（内存管理中重用内存基本上是一种很常见的行为）

这个例子你可能没有看懂，维基百科上给了一个活生生的例子——

> 你拿着一个装满钱的手提箱在飞机场，此时过来了一个火辣性感的美女，然后她很暖昧地挑逗着你，并趁你不注意的时候，把用一个一模一样的手提箱和你那装满钱的箱子调了个包，然后就离开了，你看到你的手提箱还在那，于是就提着手提箱去赶飞机去了。

这就是ABA的问题。

### 解决ABA问题

维基百科上给了一个解——使用double-CAS（双保险的CAS），例如在32位系统上，我们要检查64位的内容：
1. 一次用CAS检查双倍长度的值，前半部是值，后半部分是一个计数器。
2. 只有这两个都一样，才算通过检查，要吧赋新的值。并把计数器累加1。

这样一来，ABA发生时，虽然值一样，但是计数器就不一样（但是在32位的系统上，这个计数器会溢出回来又从1开始的，这还是会有ABA的问题）

## 利用CAS实现无锁数据结构

### 无锁队列的实现

请移步：[CollShell-无锁队列的实现](https://coolshell.cn/articles/8239.html)

### 无锁HashMap的实现

请移步：[CollShell-无锁HASHMAP的原理与实现](https://coolshell.cn/articles/9703.html)
