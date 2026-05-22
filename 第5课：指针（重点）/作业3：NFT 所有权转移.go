package main

import (
	"errors"
	"fmt"
)

type NFT struct {
	TokenID int    // NFT 唯一编号
	Name    string // NFT 名称
	Owner   *User  // 指针指向所有者（指针→改一处到处生效）
}

type User struct {
	Address string  // 用户钱包地址
	NFTs    []*NFT  // 用户拥有的 NFT 列表（切片里存的是指针）
}

// TransferNFT 转移单个 NFT：从 from 转给 to
func TransferNFT(nft *NFT, from, to *User) error {
	// 1. 验证 from 是当前所有者
	if nft == nil {                                 // 防御：NFT 不能为空指针
		return errors.New("NFT 不存在")
	}
	if from == nil || to == nil {                   // 防御：用户不能为空
		return errors.New("用户不存在")
	}
	if nft.Owner != from {                          // 比较指针地址，判断 from 是不是真正的所有者
		return errors.New("转移失败：from 不是该 NFT 的所有者")
	}

	// 2. 从 from.NFTs 中移除
	found := false                                  // 标记是否在 from 的 NFT 列表中找到
	for i, n := range from.NFTs {                   // 遍历 from 持有的 NFT
		if n == nft {                               // 指针比较，找到同一个 NFT
			from.NFTs = append(from.NFTs[:i], from.NFTs[i+1:]...) // 切片删除：跳过第 i 个
			found = true
			break                                   // 找到就停止
		}
	}
	if !found {                                     // 防御：NFT 不在 from 的列表中
		return errors.New("转移失败：NFT 不在 from 的持有列表中")
	}

	// 3. 添加到 to.NFTs
	to.NFTs = append(to.NFTs, nft)                  // 追加指针到 to 的 NFT 列表

	// 4. 更新 nft.Owner
	nft.Owner = to                                  // 修改 Owner 指向新用户（指针赋值）

	return nil
}

// BatchTransferNFTs 批量转移 NFT：一次转多个，失败则回滚
func BatchTransferNFTs(nfts []*NFT, from, to *User) error {
	// 转移多个 NFT

	if from == nil || to == nil {                   // 防御：用户不能为空
		return errors.New("用户不存在")
	}

	var transferred []*NFT                          // 记录已成功转移的 NFT，用于回滚
	for i, nft := range nfts {                      // 遍历所有要转移的 NFT
		if err := TransferNFT(nft, from, to); err != nil { // 调用单次转移
			// 回滚：把已转移的 NFT 原路退回
			for _, t := range transferred {         // 遍历已转移的
				TransferNFT(t, to, from)            // 逆向转移：从 to 转回 from
			}
			return fmt.Errorf("批量转移第 %d 个 NFT 失败: %w", i+1, err)
		}
		transferred = append(transferred, nft)      // 记录成功转移，加入回滚列表
	}
	return nil
}

func main() {
	// —————— 1. 创建用户 ——————
	alice := &User{Address: "0xAlice001"}           // &User{} 创建结构体并取指针
	bob := &User{Address: "0xBob002"}
	charlie := &User{Address: "0xCharlie003"}

	// —————— 2. 铸造 NFT 给 Alice ——————
	nft1 := &NFT{TokenID: 1, Name: "无聊猿 #1", Owner: alice}
	nft2 := &NFT{TokenID: 2, Name: "加密朋克 #2", Owner: alice}
	nft3 := &NFT{TokenID: 3, Name: "阿兹特克 #3", Owner: alice}

	alice.NFTs = []*NFT{nft1, nft2, nft3}           // Alice 初始持有 3 个 NFT

	// —————— 3. 打印初始状态 ——————
	fmt.Println("===== 初始持有 =====\n")
	printUser(alice)                                 // 打印用户持有的 NFT
	printUser(bob)

	// —————— 4. 单次转移 ——————
	fmt.Println("===== 单次转移：无聊猿 #1 Alice → Bob =====\n")
	err := TransferNFT(nft1, alice, bob)
	if err != nil {
		fmt.Println("转移失败:", err)
	}
	fmt.Printf("nft1.Owner 指向: %s\n", nft1.Owner.Address) // 验证 Owner 已更新
	printUser(alice)
	printUser(bob)

	// —————— 5. 测试非所有者转移（应失败）——————
	fmt.Println("===== 非所有者转移（Charlie 试图转 Bob 的阿兹特克）=====\n")
	err = TransferNFT(nft3, charlie, bob)            // charlie 不是 nft3 的 Owner
	if err != nil {
		fmt.Println("预期错误:", err)                 // 应报错
	}

	// —————— 6. 批量转移 ——————
	fmt.Println("\n===== 批量转移：剩余 2 个 Alice → Charlie =====\n")
	err = BatchTransferNFTs([]*NFT{nft2, nft3}, alice, charlie)
	if err != nil {
		fmt.Println("批量转移失败:", err)
	}
	printUser(alice)
	printUser(bob)
	printUser(charlie)

	// —————— 7. 测试批量回滚 ——————
	fmt.Println("===== 批量回滚测试 =====\n")
	// 先把无聊猿还给 Alice
	TransferNFT(nft1, bob, alice)
	fmt.Println("转移前:")
	printUser(alice)
	printUser(bob)

	// 制造失败场景：nft2 所有者已是 charlie，所以从 alice 转移 nft2 会失败
	// 但 nft1 在第一轮成功转移，触发回滚后 nft1 应退回
	nftFake := &NFT{TokenID: 999, Name: "假 NFT", Owner: charlie} // 不属于 alice
	err = BatchTransferNFTs([]*NFT{nft1, nftFake}, alice, bob)
	if err != nil {
		fmt.Println("批量转移失败（预期）:", err)
	}
	fmt.Println("\n转移后（nft1 应回滚回 Alice）:")
	printUser(alice)
	printUser(bob)

	// —————— 8. nil 防御测试 ——————
	fmt.Println("===== nil 防御 =====\n")
	err = TransferNFT(nil, alice, bob)
	fmt.Println("nil NFT:", err)
	err = TransferNFT(nft1, nil, bob)
	fmt.Println("nil from:", err)
}

// printUser 打印用户持有的 NFT 列表
func printUser(u *User) {
	if u == nil || len(u.NFTs) == 0 {               // 防御：用户为空或无 NFT
		fmt.Printf("%s 持有: (无)\n", u.Address)
		return
	}
	fmt.Printf("%s 持有: ", u.Address)
	for i, nft := range u.NFTs {                    // 遍历 NFT 切片
		if i > 0 {
			fmt.Print(", ")                          // 多个 NFT 之间加逗号
		}
		fmt.Printf("[%s]", nft.Name)                 // 打印 NFT 名称
	}
	fmt.Println()
}
