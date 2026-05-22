package main

import (
	"fmt"
)

type Transaction struct {
	Hash     string  // 交易哈希（唯一标识）
	From     string  // 发送方地址
	To       string  // 接收方地址
	Amount   float64 // 转账金额
	GasPrice float64 // Gas 价格（越高越优先打包）
	Status   string  // 状态：pending / confirmed / failed
}

type TxPool struct {
	Pending   []*Transaction // 待处理交易队列
	Confirmed []*Transaction // 已确认交易
	Failed    []*Transaction // 失败交易
}

// AddPending 添加一笔待处理交易到交易池
func (tp *TxPool) AddPending(tx *Transaction) {
	tx.Status = "pending"                         // 统一设置状态为 pending
	tp.Pending = append(tp.Pending, tx)           // 追加到 Pending 切片（指针，共享同一块内存）
}

// ConfirmTransaction 根据哈希确认一笔交易：从 Pending 移到 Confirmed
func (tp *TxPool) ConfirmTransaction(hash string) bool {
	for i, tx := range tp.Pending {               // 遍历 Pending 队列，i=索引，tx=元素
		if tx.Hash == hash {                       // 匹配哈希
			tx.Status = "confirmed"                // 更改状态（指针→原始数据同步修改）
			tp.Confirmed = append(tp.Confirmed, tx) // 加入已确认列表
			tp.Pending = append(tp.Pending[:i], tp.Pending[i+1:]...) // 从 Pending 中删除（切片拼接技巧）
			return true                            // 找到并处理，返回 true
		}
	}
	return false // 没找到这笔交易
}

// FailTransaction 根据哈希标记一笔交易失败：从 Pending 移到 Failed
func (tp *TxPool) FailTransaction(hash string) bool {
	for i, tx := range tp.Pending {
		if tx.Hash == hash {
			tx.Status = "failed"
			tp.Failed = append(tp.Failed, tx)
			tp.Pending = append(tp.Pending[:i], tp.Pending[i+1:]...) // 切片删除：跳过第 i 个元素
			return true
		}
	}
	return false
}

// GetHighestGasPriceTx 返回 Pending 中 Gas 出价最高的交易（矿工优先打包）
func (tp *TxPool) GetHighestGasPriceTx() *Transaction {
	if len(tp.Pending) == 0 {                     // 防御：Pending 为空
		return nil                                 // 返回 nil 指针
	}
	highest := tp.Pending[0]                      // 先假设第一个最高
	for _, tx := range tp.Pending {               // 遍历所有 Pending 交易
		if tx.GasPrice > highest.GasPrice {        // 发现更高的 Gas
			highest = tx                           // 更新最高记录
		}
	}
	return highest                                // 返回指针（不拷贝，调用方可直接修改）
}

// RemoveConfirmedTransactions 清空已确认的交易列表（打包后清理）
func (tp *TxPool) RemoveConfirmedTransactions() {
	tp.Confirmed = nil                            // 赋值 nil 让 GC 回收，等价于清空切片
}

func main() {
	pool := &TxPool{} // &TxPool{} 创建交易池并取指针，后续方法都需要指针接收者

	// —————— 1. 添加交易到 Pending ——————
	fmt.Println("===== 添加交易 =====")
	tx1 := &Transaction{Hash: "0xaaa", From: "Alice", To: "Bob", Amount: 10, GasPrice: 20}
	tx2 := &Transaction{Hash: "0xbbb", From: "Bob", To: "Charlie", Amount: 5, GasPrice: 50}
	tx3 := &Transaction{Hash: "0xccc", From: "Charlie", To: "Alice", Amount: 2, GasPrice: 30}
	tx4 := &Transaction{Hash: "0xddd", From: "Alice", To: "Bob", Amount: 8, GasPrice: 15}

	pool.AddPending(tx1) // 把 tx1 指针加入交易池，共享同一块内存
	pool.AddPending(tx2)
	pool.AddPending(tx3)
	pool.AddPending(tx4)

	fmt.Printf("Pending 数量: %d\n", len(pool.Pending))    // len() 获取切片长度
	for _, tx := range pool.Pending {                      // range 遍历切片
		fmt.Printf("  %s | Gas: %.0f | %s → %s | %.2f ETH\n",
			tx.Hash, tx.GasPrice, tx.From, tx.To, tx.Amount)
	}

	// —————— 2. 查找 Gas 最高交易 ——————
	fmt.Println("\n===== 最高 Gas 交易 =====")
	highest := pool.GetHighestGasPriceTx()                 // 返回 *Transaction 指针
	if highest != nil {
		fmt.Printf("最高 Gas: %s (GasPrice: %.0f)\n", highest.Hash, highest.GasPrice)
	}

	// —————— 3. 确认一笔交易 ——————
	fmt.Println("\n===== 确认交易 0xbbb =====")
	ok := pool.ConfirmTransaction("0xbbb")                 // 确认 bbb，从 Pending→Confirmed
	fmt.Printf("确认结果: %v\n", ok)                        // %v 打印任意类型的值
	fmt.Printf("Pending: %d | Confirmed: %d | Failed: %d\n",
		len(pool.Pending), len(pool.Confirmed), len(pool.Failed))

	// 验证 tx2 的状态同步更新了（因为是同一指针）
	fmt.Printf("tx2 状态: %s\n", tx2.Status)               // 应为 "confirmed"

	// —————— 4. 标记失败交易 ——————
	fmt.Println("\n===== 标记失败 0xddd =====")
	ok = pool.FailTransaction("0xddd")                     // 标记 ddd 失败
	fmt.Printf("标记结果: %v | tx4 状态: %s\n", ok, tx4.Status)
	fmt.Printf("Pending: %d | Confirmed: %d | Failed: %d\n",
		len(pool.Pending), len(pool.Confirmed), len(pool.Failed))

	// —————— 5. 确认不存在的交易 ——————
	fmt.Println("\n===== 确认不存在的交易 =====")
	ok = pool.ConfirmTransaction("0xnotexist")             // 预期返回 false
	fmt.Printf("确认结果: %v（未找到）\n", ok)

	// —————— 6. 确认后 Gas 最高者变化 ——————
	fmt.Println("\n===== 确认后的最高 Gas =====")
	highest = pool.GetHighestGasPriceTx()                  // bbb 已被移走，最高变为 ccc
	fmt.Printf("最高 Gas: %s (GasPrice: %.0f)\n", highest.Hash, highest.GasPrice)

	// —————— 7. 清空已确认交易 ——————
	fmt.Println("\n===== 清空已确认交易 =====")
	pool.RemoveConfirmedTransactions()
	fmt.Printf("Confirmed 数量: %d（已清空）\n", len(pool.Confirmed))

	// —————— 8. Pending 为空时查找 ——————
	fmt.Println("\n===== 空 Pending 查找最高 Gas =====")
	emptyPool := &TxPool{}                                 // 新建空池
	result := emptyPool.GetHighestGasPriceTx()
	if result == nil {                                     // nil 判断，防止空指针 panic
		fmt.Println("Pending 为空，返回 nil")
	}

	// —————— 9. 最终状态汇总 ——————
	fmt.Println("\n===== 交易池最终状态 =====")
	fmt.Printf("Pending:   %d 笔\n", len(pool.Pending))
	for _, tx := range pool.Pending {
		fmt.Printf("  %s (%s)\n", tx.Hash, tx.Status)
	}
	fmt.Printf("Confirmed: %d 笔\n", len(pool.Confirmed))
	fmt.Printf("Failed:    %d 笔\n", len(pool.Failed))
	for _, tx := range pool.Failed {
		fmt.Printf("  %s (%s)\n", tx.Hash, tx.Status)
	}
}
