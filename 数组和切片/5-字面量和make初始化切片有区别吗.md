#  字面量和make初始化切片有区别吗

有，当我们使用字面量（`[]int{1,2,3}`）创建切片时，会被编译器在编译期间展开成如下所示的代码片段：

```go
var s [3]int
s[0] = 1
s[1] = 2
s[2] = 3
var vauto *[3]int = new([3]int) // 初始化一个数组指针
*vauto = s
slice := vauto[:]
```

汇编输出：

```go
$ go build main.go && go tool objdump ./main | grep "main.go:4" 

493b6610                CMPQ 0x10(R14), SP                      
0f8685000000            JBE 0x108d96f                           
4883ec40                sUBQ $0x40, SP                          
48896c2438              MOVQ BP, 0x38(SP)                       
488d6c2438              LEAQ 0x38(SP), BP                       
488d05e1870000          LEAQ runtime.rodata+34368(SB), AX       
90                      NOPL                                    
e8dbe4f7ff              CALL runtime.newobject(SB)              
48c70001000000          MOVQ $0x1, 0(AX)                        
48c7400802000000        MOVQ $0x2, 0x8(AX)                      
48c7401003000000        MOVQ $0x3, 0x10(AX)                     
440f117c2428            MOVUPS X15, 0x28(SP)                    
bb03000000              MOVL $0x3, BX                           
4889d9                  MOVQ BX, CX                             
e8d1bef7ff              CALL runtime.convTslice(SB)             
488d0dea690000          LEAQ runtime.rodata+26752(SB), CX       
48894c2428              MOVQ CX, 0x28(SP)                       
4889442430              MOVQ AX, 0x30(SP)                       
488b6c2438              MOVQ 0x38(SP), BP                       
4883c440                ADDQ $0x40, SP                          
c3                      RET                                     
e8acddfcff              CALL runtime.morestack_noctxt.abi0(SB)  
e967ffffff              JMP main.main(SB) 
```