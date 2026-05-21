package main

import (
	"errors"
	"fmt"
)

// TODO: 定义 NFT 结构体
type NFT struct {
	// TokenID, Name, Owner, Price, IsListed
	TokenID  int
	Name     string
	Owner    string
	Price    float64
	IsListed bool
}

// TODO: 实现 NFT 方法
// 1. List(price float64) - 上架
// 2. Unlist() - 下架
// 3. Transfer(newOwner string) - 转移
// 4. GetInfo() string - 获取信息

func (n *NFT) List(price float64) {
	n.Price = price
	n.IsListed = true
}

func (n *NFT) Unlist() {
	n.IsListed = false
}

func (n *NFT) Transfer(newOwner string) {
	n.Owner = newOwner
	n.IsListed = false
}

func (n *NFT) GetInfo() string {
	status := "未上架"
	if n.IsListed {
		status = "已上架"
	}
	return fmt.Sprintf("TokenID: %d, Name: %s, Owner: %s, Price: %.2f, Status: %s", n.TokenID, n.Name, n.Owner, n.Price, status)
}

// TODO: 定义 Marketplace 结构体
type Marketplace struct {
	// NFTs map[int]*NFT
	// TotalSales float64
	NFTs       map[int]*NFT
	TotalSales float64
}

// TODO: 实现 Marketplace 方法
// 1. AddNFT(nft *NFT)
// 2. BuyNFT(tokenID int, buyer string) error
// 3. GetListedNFTs() []*NFT
// 4. GetTotalSales() float64

func NewMarketplace() *Marketplace {
	return &Marketplace{
		NFTs: make(map[int]*NFT),
	}
}

func (m *Marketplace) AddNFT(nft *NFT) {
	m.NFTs[nft.TokenID] = nft
}

func (m *Marketplace) BuyNFT(tokenID int, buyer string) error {
	nft, ok := m.NFTs[tokenID]
	if !ok {
		return errors.New("NFT 未找到")
	}
	if !nft.IsListed {
		return errors.New("NFT 未上架")
	}
	m.TotalSales += nft.Price
	nft.Transfer(buyer)
	nft.Unlist()
	return nil
}

func (m *Marketplace) GetListedNFTs() []*NFT {
	var res []*NFT
	for _, nft := range m.NFTs {
		if nft.IsListed {
			res = append(res, nft)
		}
	}
	return res
}

func (m *Marketplace) GetTotalSales() float64 {
	return m.TotalSales
}

func main() {
	market := NewMarketplace()

	// 测试你的代码
	nft1 := &NFT{TokenID: 1, Name: "CryptoCat", Owner: "Alice"}
	nft2 := &NFT{TokenID: 2, Name: "PixelPunk", Owner: "Bob"}

	market.AddNFT(nft1)
	market.AddNFT(nft2)

	nft1.List(2.5)
	nft2.List(5.0)

	fmt.Println("当前上架的 NFT:")
	for _, n := range market.GetListedNFTs() {
		fmt.Println(n.GetInfo())
	}

	fmt.Println("\n买家 Carol 购买 TokenID 1...")
	if err := market.BuyNFT(1, "Carol"); err != nil {
		fmt.Println("购买失败:", err)
	} else {
		fmt.Println("购买成功")
	}

	fmt.Println("\n购买后 NFT 状态:")
	for _, n := range market.NFTs {
		fmt.Println(n.GetInfo())
	}

	fmt.Printf("\n市场总成交额: %.2f\n", market.GetTotalSales())
}
