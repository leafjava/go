package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

// TONService 封装 TON 区块链的核心操作，包括查询余额、转账、获取链信息等
type TONService struct {
	api    *ton.APIClient            // TON API 客户端，用于与链上交互
	client *liteclient.ConnectionPool // Lite 客户端连接池，管理底层网络连接
}

// NewTONService 使用 Lite 客户端配置 URL 创建 TON 服务实例
// 通过连接 TON 网络的 Lite 服务器来建立通信通道
func NewTONService(configURL string) (*TONService, error) {
	// 步骤1: 创建 Lite 客户端连接池，用于管理与 TON 网络的多个连接
	client := liteclient.NewConnectionPool()

	// 步骤2: 从配置 URL 加载节点信息并建立连接
	// configURL 指向一个包含 TON 网络节点公钥和地址的 JSON 配置文件
	err := client.AddConnectionsFromConfigUrl(context.Background(), configURL)
	if err != nil {
		return nil, fmt.Errorf("连接 TON 失败: %w", err)
	}

	// 步骤3: 基于已建立的连接池创建 TON API 客户端，并封装为服务返回
	return &TONService{
		api:    ton.NewAPIClient(client),
		client: client,
	}, nil
}

// GetBalance 查询指定地址的 TON 余额
// 返回以 TON 为单位（非 nanoTON）的余额浮点数
func (s *TONService) GetBalance(addrStr string) (float64, error) {
	// 步骤1: 将字符串地址解析为 TON 地址结构体
	addr, err := address.ParseAddr(addrStr)
	if err != nil {
		return 0, fmt.Errorf("地址格式错误: %w", err)
	}

	// 步骤2: 获取主链最新区块信息（GetAccount 需要指定查询的区块）
	block, err := s.api.GetMasterchainInfo(context.Background())
	if err != nil {
		return 0, fmt.Errorf("获取主链信息失败: %w", err)
	}

	// 步骤3: 从链上获取该地址在指定区块的完整账户信息（包括余额、状态、代码等）
	account, err := s.api.GetAccount(context.Background(), block, addr)
	if err != nil {
		return 0, fmt.Errorf("查询账户失败: %w", err)
	}

	// 步骤4: 将 nanoTON（10^-9 TON）转换为 TON 单位返回
	// 链上余额以 nanoTON 存储，除以 1e9 得到人类可读的 TON 数量
	return float64(account.State.Balance.Nano().Uint64()) / 1e9, nil
}

// GetAccountStatus 获取账户状态
// 返回状态字符串，如 "active"（正常）、"uninitialized"（未激活）、"frozen"（冻结）
func (s *TONService) GetAccountStatus(addrStr string) (string, error) {
	// 步骤1: 将字符串地址解析为 TON 地址结构体
	addr, err := address.ParseAddr(addrStr)
	if err != nil {
		return "", fmt.Errorf("地址格式错误: %w", err)
	}

	// 步骤2: 获取主链最新区块（GetAccount 需要指定查询的区块）
	block, err := s.api.GetMasterchainInfo(context.Background())
	if err != nil {
		return "", fmt.Errorf("获取主链信息失败: %w", err)
	}

	// 步骤3: 从链上获取账户信息
	account, err := s.api.GetAccount(context.Background(), block, addr)
	if err != nil {
		return "", fmt.Errorf("查询账户失败: %w", err)
	}

	// 步骤4: 返回账户的状态字段（AccountStatus 是 string 的类型别名，需显式转换）
	return string(account.State.Status), nil
}

// SendTON 发送 TON 转账
// mnemonic: 发送方钱包的助记词（24个单词，空格分隔）
// toAddrStr: 接收方地址
// amountTON: 转账金额（TON 单位，如 "1.5"）
// comment: 转账附言，会写入交易消息体中
func (s *TONService) SendTON(mnemonic string, toAddrStr string, amountTON string, comment string) (string, error) {
	// 步骤1: 将助记词按空格拆分为单词数组，用于恢复钱包
	words := strings.Split(mnemonic, " ")

	// 步骤2: 通过助记词恢复钱包私钥，使用 V4R2 钱包合约版本
	// V4R2 是目前 TON 最常用的钱包合约版本，支持插件和订阅功能
	w, err := wallet.FromSeed(s.api, words, wallet.V4R2)
	if err != nil {
		return "", fmt.Errorf("恢复钱包失败: %w", err)
	}

	// 步骤3: 解析接收方地址字符串为 TON 地址结构体
	toAddr, err := address.ParseAddr(toAddrStr)
	if err != nil {
		return "", fmt.Errorf("接收地址格式错误: %w", err)
	}

	// 步骤4: 将字符串金额（如 "1.5"）转换为 TON 链上的 Coins 类型（nanoTON 精度）
	amount, err := tlb.FromTON(amountTON)
	if err != nil {
		return "", fmt.Errorf("金额格式错误: %w", err)
	}

	// 步骤5: 使用钱包的 BuildTransfer 方法构建转账消息
	// 内部自动处理附言编码、消息模式（PayGasSeparately + IgnoreErrors = 3）等
	msg, err := w.BuildTransfer(toAddr, amount, false, comment)
	if err != nil {
		return "", fmt.Errorf("构建转账消息失败: %w", err)
	}

	// 步骤6: 发送交易并等待区块确认（最后一个参数 true 表示等待确认）
	err = w.Send(context.Background(), msg, true)
	if err != nil {
		return "", fmt.Errorf("转账失败: %w", err)
	}

	// 步骤7: 返回可读的转账成功消息
	return fmt.Sprintf("成功发送 %s TON 到 %s", amountTON, toAddrStr), nil
}

// GetWalletAddress 从助记词恢复并获取钱包地址
// 用于用户输入助记词后显示对应的钱包地址
func (s *TONService) GetWalletAddress(mnemonic string) (string, error) {
	// 步骤1: 将助记词拆分为单词数组
	words := strings.Split(mnemonic, " ")

	// 步骤2: 从助记词恢复钱包（V4R2 合约版本）
	w, err := wallet.FromSeed(s.api, words, wallet.V4R2)
	if err != nil {
		return "", fmt.Errorf("恢复钱包失败: %w", err)
	}

	// 步骤3: 获取钱包地址的字符串表示（raw 格式）
	return w.Address().String(), nil
}

// GetMasterchainInfo 获取 TON 主链（Masterchain）最新区块信息
// 返回最新区块的 ID 和扩展信息，可用于同步状态判断
func (s *TONService) GetMasterchainInfo() (*ton.BlockIDExt, error) {
	// 步骤1: 调用 API 获取主链信息
	// 注意：GetMasterchainInfo 直接返回 *BlockIDExt，不是包装结构体，无需访问 .Last
	info, err := s.api.GetMasterchainInfo(context.Background())
	if err != nil {
		return nil, fmt.Errorf("获取主链信息失败: %w", err)
	}

	// 步骤2: 直接返回区块信息（包含区块高度、哈希等）
	return info, nil
}

// Close 关闭 TON 服务，释放网络连接资源
func (s *TONService) Close() {
	// 仅在 Lite 客户端存在时才执行关闭（HTTP 模式不需要关闭连接池）
	if s.client != nil {
		s.client.Stop() // 停止连接池中所有连接
	}
}
