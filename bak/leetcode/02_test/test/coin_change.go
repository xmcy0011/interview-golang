package main

import "fmt"

func main() {
	coins := []int{1, 2, 5}
	fmt.Println(coinChange(coins, 11))
}

// 零钱兑换
// 状态推导
// F(0) = 0
// F(1) = min(F(1-1), F(1-2), F(1-5)) +1 = min(F(0), MAX, MAX)+1 = 1
// F(2) = min(F(2-1), F(2-2), F(2-5)) +1 = min(F(1), F(0), MAX)+1 = min(1,0,MAX)+1 = 1
// F(3) = min(F(3-1), F(3-2), F(3-5)) +1 = min(F(2), F(1), MAX)+1 = min(1,1,MAX)+1 = 2
// F(n) = min(F(n-1), F(n-2), F(n-5))+1
//
// 状态转移方程式：
// F(n) = 0, n=0
// F(n) = min(c, DP[n-c]+1)
func coinChange(coins []int, amount int) int {
	k := amount + 1 // 模拟无穷大

	// 初始化dp
	dp := make([]int, k)

	// 初始化状态
	dp[0] = 0
	// 把所有面额的硬币组合都设置为无穷大，如果没有找到组合，就返回-1
	for i := 1; i < k; i++ {
		dp[i] = k
	}

	// 从面额1开始，一直到指定的面额，比如11，就循环11次
	for i := 1; i <= amount; i++ {
		for _, coin := range coins {
			// 余下的面额
			restAmount := i - coin
			// 说明，不存在这种组合，直接下一个尝试用下一个面额的硬币
			if restAmount < 0 {
				continue
			}

			// 决策：当前已经用了一枚硬币，所以需要加1
			dp[i] = min(dp[i], dp[restAmount]+1)
		}
	}

	if dp[amount] == k {
		return -1
	}
	return dp[amount]
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
