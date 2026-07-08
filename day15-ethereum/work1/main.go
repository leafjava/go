package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

// BalanceResponse 余额查询响应
type BalanceResponse struct {
	Address string  `json:"address"`
	Wei     string  `json:"wei"`     // wei 原始值
	Ether   float64 `json:"ether"`   // ETH 单位
	USD     float64 `json:"usd"`     // 美元价值
}

var (
	ethClient *ethclient.Client
)

func main() {
	// 1. 连接以太坊节点
	client, err := ethclient.Dial("https://ethereum.publicnode.com")
	if err != nil {
		log.Fatalf("连接以太坊节点失败: %v", err)
	}
	ethClient = client
	defer client.Close()

	// 2. 设置 Gin 路由
	r := gin.Default()

	api := r.Group("/api/v1/ethereum")
	{
		api.GET("/balance/:address", handleGetBalance)
	}

	// 3. 启动服务
	fmt.Println("🚀 服务启动: http://localhost:9090")
	fmt.Println("📡 API: GET /api/v1/ethereum/balance/:address")
	if err := r.Run(":9090"); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}

// handleGetBalance 处理余额查询请求
func handleGetBalance(c *gin.Context) {
	address := c.Param("address")

	// 1. 验证地址格式
	if !common.IsHexAddress(address) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("无效的以太坊地址: %s", address),
		})
		return
	}

	// 2. 查询 ETH 余额
	addr := common.HexToAddress(address)
	balanceWei, err := ethClient.BalanceAt(c.Request.Context(), addr, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("查询余额失败: %v", err),
		})
		return
	}

	// 3. 获取 ETH/USD 价格（价格获取失败不影响余额返回）
	ethPrice, priceErr := getETHPrice()
	if priceErr != nil {
		fmt.Printf("⚠️  获取 ETH 价格失败: %v\n", priceErr)
	}

	// 4. 计算 ETH 和 USD 值
	wei := new(big.Float).SetInt(balanceWei)
	ether := new(big.Float).Quo(wei, big.NewFloat(1e18)) // wei / 10^18
	etherFloat, _ := ether.Float64()
	usdValue := etherFloat * ethPrice

	// 5. 返回结果
	c.JSON(http.StatusOK, BalanceResponse{
		Address: address,
		Wei:     balanceWei.String(),
		Ether:   etherFloat,
		USD:     usdValue,
	})
}

// getETHPrice 获取 ETH/USD 价格，依次尝试多个数据源
func getETHPrice() (float64, error) {
	// 数据源列表：Binance → Coinbase → CoinGecko
	sources := []struct {
		name string
		url  string
		parse func(body []byte) (float64, error)
	}{
		{
			name: "Binance",
			url:  "https://api.binance.com/api/v3/ticker/price?symbol=ETHUSDT",
			parse: func(body []byte) (float64, error) {
				var resp struct {
					Price string `json:"price"`
				}
				if err := json.Unmarshal(body, &resp); err != nil {
					return 0, err
				}
				return parseFloat(resp.Price)
			},
		},
		{
			name: "Coinbase",
			url:  "https://api.coinbase.com/v2/prices/ETH-USD/spot",
			parse: func(body []byte) (float64, error) {
				var resp struct {
					Data struct {
						Amount string `json:"amount"`
					} `json:"data"`
				}
				if err := json.Unmarshal(body, &resp); err != nil {
					return 0, err
				}
				return parseFloat(resp.Data.Amount)
			},
		},
	}

	for _, src := range sources {
		price, err := fetchPrice(src.url, src.parse)
		if err == nil {
			fmt.Printf("✅ ETH 价格 (来自 %s): $%.2f\n", src.name, price)
			return price, nil
		}
		fmt.Printf("⚠️  %s 价格获取失败: %v\n", src.name, err)
	}

	return 0, fmt.Errorf("所有价格数据源均不可用")
}

func fetchPrice(url string, parse func([]byte) (float64, error)) (float64, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}

	return parse(body)
}

func parseFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("解析价格 '%s' 失败: %w", s, err)
	}
	if f <= 0 {
		return 0, fmt.Errorf("价格无效: %f", f)
	}
	return f, nil
}
