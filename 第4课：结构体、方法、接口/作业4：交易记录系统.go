package main

import (
	"fmt"
	"time"
)

//核心功能：
//
//BaseTransaction - 基础交易结构体，包含所有交易的通用字段（哈希、发送方、接收方、金额、时间戳）
//EthTransaction - 以太坊交易，嵌入基础结构并添加 Gas 相关字段
//TONTransaction - TON 交易，嵌入基础结构并添加转发费用
//TransactionManager - 交易管理器，统一管理所有类型的交易
//实现的方法：
//
//GetFee() - 计算不同链的手续费（ETH 用 Gas，TON 用 ForwardFee）
//GetInfo() - 格式化输出交易详情
//AddTransaction() - 添加交易
//GetTotalFees() - 统计总手续费
//GetTransactionCount() - 获取交易数量
//测试用例：
//
//创建了 2 笔 ETH 交易和 2 笔 TON 交易
//展示了添加、查询、统计等完整功能

// Transaction 接口定义了所有交易类型必须实现的方法
type Transaction interface {
	GetFee() float64 // 获取交易手续费
	GetInfo() string // 获取交易详细信息
}

// TODO: 定义 BaseTransaction 基础结构体
type BaseTransaction struct {
	Hash      string    // 交易哈希值，唯一标识一笔交易
	From      string    // 发送方地址
	To        string    // 接收方地址
	Amount    float64   // 交易金额
	Timestamp time.Time // 交易时间戳
}

// TODO: 定义 EthTransaction（嵌入 BaseTransaction）
type EthTransaction struct {
	BaseTransaction         // 嵌入基础交易结构体，继承其所有字段
	GasUsed         uint64  // 实际消耗的 Gas 数量
	GasPrice        float64 // Gas 价格（单位：Gwei）
}

// TODO: 实现方法
// 1. GetFee() float64 - 计算手续费
// GetFee 计算以太坊交易的手续费
// 手续费 = GasUsed * GasPrice
func (e *EthTransaction) GetFee() float64 {
	return float64(e.GasUsed) * e.GasPrice
}

// 2. GetInfo() string - 获取交易信息
// GetInfo 返回以太坊交易的详细信息字符串
func (e *EthTransaction) GetInfo() string {
	return fmt.Sprintf(
		"[ETH交易]\n"+
			"  哈希: %s\n"+
			"  从: %s\n"+
			"  到: %s\n"+
			"  金额: %.4f ETH\n"+
			"  Gas使用: %d\n"+
			"  Gas价格: %.4f Gwei\n"+
			"  手续费: %.6f ETH\n"+
			"  时间: %s",
		e.Hash, e.From, e.To, e.Amount,
		e.GasUsed, e.GasPrice, e.GetFee(),
		e.Timestamp.Format("2006-01-02 15:04:05"),
	)
}

// TODO: 定义 TONTransaction（嵌入 BaseTransaction）
type TONTransaction struct {
	BaseTransaction         // 嵌入基础交易结构体
	ForwardFee      float64 // TON 网络的转发费用
}

// GetFee 计算 TON 交易的手续费
// TON 的手续费就是 ForwardFee
func (t *TONTransaction) GetFee() float64 {
	return t.ForwardFee
}

// GetInfo 返回 TON 交易的详细信息字符串
func (t *TONTransaction) GetInfo() string {
	return fmt.Sprintf(
		"[TON交易]\n"+
			"  哈希: %s\n"+
			"  从: %s\n"+
			"  到: %s\n"+
			"  金额: %.4f TON\n"+
			"  转发费: %.6f TON\n"+
			"  时间: %s",
		t.Hash, t.From, t.To, t.Amount,
		t.ForwardFee,
		t.Timestamp.Format("2006-01-02 15:04:05"),
	)
}

// TODO: 实现 TransactionManager
type TransactionManager struct {
	transactions []Transaction // 存储所有交易的切片，使用 Transaction 接口类型
}

// NewTransactionManager 创建一个新的交易管理器实例
func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		transactions: make([]Transaction, 0), // 初始化空的交易切片
	}
}

// TODO: 实现方法
// 1. AddTransaction(tx interface{})
// AddTransaction 添加一笔新交易到管理器中
// 参数 tx 必须实现 Transaction 接口
func (tm *TransactionManager) AddTransaction(tx Transaction) {
	tm.transactions = append(tm.transactions, tx) // 将交易追加到切片末尾
	fmt.Printf("✓ 交易已添加 (哈希: %s)\n\n", getHash(tx))
}

// getHash 是一个辅助函数，用于从不同类型的交易中提取哈希值
func getHash(tx Transaction) string {
	switch t := tx.(type) {
	case *EthTransaction:
		return t.Hash
	case *TONTransaction:
		return t.Hash
	default:
		return "unknown"
	}
}

// 2. GetTotalFees() float64
// GetTotalFees 计算所有交易的总手续费
func (tm *TransactionManager) GetTotalFees() float64 {
	totalFees := 0.0
	// 遍历所有交易，累加每笔交易的手续费
	for _, tx := range tm.transactions {
		totalFees += tx.GetFee()
	}
	return totalFees
}

// 3. GetTransactionCount() int
// GetTransactionCount 返回管理器中的交易总数
func (tm *TransactionManager) GetTransactionCount() int {
	return len(tm.transactions)
}

// PrintAllTransactions 打印所有交易的详细信息
func (tm *TransactionManager) PrintAllTransactions() {
	fmt.Println("=== 所有交易记录 ===")
	for i, tx := range tm.transactions {
		fmt.Printf("\n交易 #%d:\n%s\n", i+1, tx.GetInfo())
	}
}

func main() {
	// 测试交易系统
	fmt.Println("=== 交易记录系统测试 ===\n")

	// 创建交易管理器
	manager := NewTransactionManager()

	// 创建以太坊交易示例
	ethTx1 := &EthTransaction{
		BaseTransaction: BaseTransaction{
			Hash:      "0x1a2b3c4d5e6f...",
			From:      "0xABCD1234...",
			To:        "0xEFGH5678...",
			Amount:    1.5,
			Timestamp: time.Now(),
		},
		GasUsed:  21000,   // 标准转账消耗的 Gas
		GasPrice: 0.00002, // Gas 价格（Gwei）
	}

	ethTx2 := &EthTransaction{
		BaseTransaction: BaseTransaction{
			Hash:      "0x9f8e7d6c5b4a...",
			From:      "0xIJKL9012...",
			To:        "0xMNOP3456...",
			Amount:    0.8,
			Timestamp: time.Now().Add(-1 * time.Hour), // 1小时前的交易
		},
		GasUsed:  45000, // 智能合约交互消耗更多 Gas
		GasPrice: 0.000025,
	}

	// 创建 TON 交易示例
	tonTx1 := &TONTransaction{
		BaseTransaction: BaseTransaction{
			Hash:      "ton_abc123...",
			From:      "EQ...ABC",
			To:        "EQ...DEF",
			Amount:    10.0,
			Timestamp: time.Now().Add(-30 * time.Minute), // 30分钟前的交易
		},
		ForwardFee: 0.005, // TON 转发费用
	}

	tonTx2 := &TONTransaction{
		BaseTransaction: BaseTransaction{
			Hash:      "ton_xyz789...",
			From:      "EQ...GHI",
			To:        "EQ...JKL",
			Amount:    25.5,
			Timestamp: time.Now().Add(-2 * time.Hour), // 2小时前的交易
		},
		ForwardFee: 0.008,
	}

	// 添加交易到管理器
	fmt.Println("--- 添加交易 ---")
	manager.AddTransaction(ethTx1)
	manager.AddTransaction(ethTx2)
	manager.AddTransaction(tonTx1)
	manager.AddTransaction(tonTx2)

	// 打印所有交易信息
	manager.PrintAllTransactions()

	// 统计信息
	fmt.Println("\n=== 统计信息 ===")
	fmt.Printf("交易总数: %d\n", manager.GetTransactionCount())
	fmt.Printf("总手续费: %.6f\n", manager.GetTotalFees())

	// 单独查看某笔交易
	fmt.Println("\n=== 单笔交易详情 ===")
	fmt.Println(ethTx1.GetInfo())
}
