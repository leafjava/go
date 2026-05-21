package main

import "fmt"

// TODO: 使用 iota 定义区块链网络枚举
const (
	// Ethereum = ?
	Ethereum = iota + 1
	// BSC = ?
	BSC
	// Polygon = ?
	Polygon
	// Arbitrum = ?
	Arbitrum
	// Optimism = ?
	Optimism
	// TON = ?
	TON
)

// TODO: 使用 iota 定义交易状态枚举
const (
	// Pending = ?
	Pending = iota
	// Confirmed = ?
	Confirmed
	// Failed = ?
	Failed
)

func getNetworkName(network int) string {
	// TODO: 根据网络ID返回网络名称
	switch network {
	case Ethereum:
		return "Ethereum"
	case BSC:
		return "BSC"
	case Polygon:
		return "Polygon"
	case Arbitrum:
		return "Arbitrum"
	case Optimism:
		return "Optimism"
	case TON:
		return "TON"
	default:
		return "Unknown"
	}
}

func getStatusName(status int) string {
	// TODO: 根据状态ID返回状态名称
	switch status {
	case Pending:
		return "Pending"
	case Confirmed:
		return "Confirmed"
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

func main() {
	// 测试你的枚举
	fmt.Println("网络:", getNetworkName(TON))
	fmt.Println("状态:", getStatusName(Confirmed))
}
