package main

import "fmt"

func main() {
	//coins := []int{5, 3, 1} // 所有硬币面额
	coins := []int{5, 3} // 所有硬币面额
	total := 11          // 找零总面额
	//fmt.Println(coinChangeV1(total, coins))
	fmt.Println(coinChange(total, coins))
}

// 贪心算法
// valueCount: 限制面额数量
func coinChangeV1(total int, coins []int) int {
	rest := total
	count := 0

	// 从大到小遍历所有面值（coins倒序）
	// 因为从最大的开始，相当于一次性花完最大的面额的硬币
	for _, coin := range coins {
		// 计算当前面值最多能用多少个
		num := rest / coin
		// 更新余额
		rest -= num * coin
		// 更新，已用硬币数量
		count += num

		if rest == 0 {
			return count
		}
	}

	return -1
}

// 贪心算法V2版本：解决v1版本过于贪心，无法实现5,3面额的求解问题
// valueCount: 限制面额数量
/*
func coinChange(total int, coins []int, index int) int {
	if index >= len(coins) {
		return -1
	}

	minResult, cur := -1, coins[index]
	// 当前面额能最大使用的数量
	maxCount := total / cur

	// 贪心算法，从最大面值往小面值尝试
	for count := maxCount; count >= 0; count-- {
		// 一次性取最大面额，余额还有多少？
		rest := total - count*cur
		// 刚好，直接返回
		if rest == 0 {
			minResult = min(minResult, count)
			break
		}

		// 否则，剩余面值继续贪心
		restCount := coinChange(rest, coins, index+1)
		if restCount == -1 {
			// 如果当前面值已经为0，返回-1表示尝试失败
			if count == 0 {
				break
			}
			// 否则，更换较小的面额继续求
			continue
		}
	}
	return minResult
}

func coinChangeLoop(total int,coins []int, k int) int{
    minCount := -1
    if k == len(coins){
        return min(minCount,
    }
}
*/

func coinChange(amount int, coins []int) int {
	// 模拟无穷大
	k := amount + 1
	dp := make([]int, k)

	// 初始化状态，第0位置只有
	dp[0] = 0
	for i := 1; i < k; i++ {
		dp[i] = k
	}

	for i := 1; i < k; i++ {
		for _, coin := range coins {
			if i-coin < 0 {
				continue
			}

			// 作出决策
			dp[i] = min(dp[i], dp[i-coin]+1)
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
