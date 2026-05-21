package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// 模拟区块数据
type Block struct {
	Number       int64
	Hash         string
	Transactions int
}

// TODO: 实现获取最新区块函数
func getLatestBlock() (*Block, error) {
	// 模拟返回最新区块
	// 10% 概率返回错误（模拟网络问题）
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(10) == 0 {
		return nil, errors.New("网络连接失败")
	}

	return &Block{
		Number:       18500000,
		Hash:         "0xabc123def456",
		Transactions: 150,
	}, nil
}

// TODO: 实现获取指定区块函数
func getBlockByNumber(number int64) (*Block, error) {
	// 如果 number <= 0，返回错误
	if number <= 0 {
		return nil, errors.New("区块号必须大于0")
	}

	// 否则返回模拟区块数据
	return &Block{
		Number:       number,
		Hash:         fmt.Sprintf("0x%x", number),
		Transactions: int(number % 200),
	}, nil
}

// TODO: 实现批量查询函数（使用可变参数）
func getBlocks(numbers ...int64) ([]*Block, error) {
	// 查询多个区块
	var blocks []*Block
	for _, num := range numbers {
		block, err := getBlockByNumber(num)
		// 如果任何一个查询失败，返回错误
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func main() {
	// 测试你的函数
	// 测试获取最新区块
	fmt.Println("=== 测试获取最新区块 ===")
	latestBlock, err := getLatestBlock()
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Printf("区块 #%d, Hash: %s, 交易数: %d\n", latestBlock.Number, latestBlock.Hash, latestBlock.Transactions)
	}

	// 测试获取指定区块
	fmt.Println("\n=== 测试获取指定区块 ===")
	block, err := getBlockByNumber(100)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Printf("区块 #%d, Hash: %s, 交易数: %d\n", block.Number, block.Hash, block.Transactions)
	}

	// 测试批量查询
	fmt.Println("\n=== 测试批量查询 ===")
	blocks, err := getBlocks(1, 2, 3, 4, 5)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		for _, b := range blocks {
			fmt.Printf("区块 #%d, Hash: %s, 交易数: %d\n", b.Number, b.Hash, b.Transactions)
		}
	}
}
