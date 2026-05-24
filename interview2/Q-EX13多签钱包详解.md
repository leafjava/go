# Q-EX13: 多签钱包（Gnosis Safe）详解

> **适用场景**：HashKey 全栈工程师面试 — Web3 安全进阶题  
> **难度**：中等偏难，理解链下签名 + 链上验证的协作模式是关键  
> **关联题目**：[Q9 杠杆交易订单系统设计](../week1/hashkey-fullstack-interview-questions.md#q9-如何设计一个安全的高频交易杠杆交易订单系统)、[Q10 智能合约前端交互安全](../week1/hashkey-fullstack-interview-questions.md#q10-智能合约前端交互中如何防范常见攻击)

---

## 一、什么是多签钱包？

### 一句话定义

> 多签钱包（Multi-Signature Wallet）是一份智能合约，要求 **m-of-n** 个预设的 owner 中，至少 **m** 个同意签名，才能执行一笔交易。

- **n**：所有者的总数（比如 5 个核心团队成员）
- **m**：最小签名阈值（比如至少 3 人同意）

### 为什么需要多签？

```
单签钱包的问题：
  私钥 A ──→ 转走全部资产
  ↑
  一个人知道私钥 = 单点故障
  泄露 / 离职 / 被胁迫 → 资金全丢

多签钱包的方案：
  私钥 A ─┐
  私钥 B ─┼──→ 至少 3/5 签名 ──→ 转走资产
  私钥 C ─┤
  私钥 D ─┤
  私钥 E ─┘
  ↑
  一人私钥泄露 ≠ 资金丢失
  需要攻破至少 3 个人
```

**HashKey 这样的持牌交易所，资产管理大概率走 3/5 或 5/8 多签方案。**

---

## 二、Gnosis Safe 核心机制

### 2.1 整体流程

```
┌─────────────────────────────────────────────────────────┐
│                     Gnosis Safe 流程                      │
│                                                          │
│  ① 任意 owner 提交交易到 Safe 合约                        │
│     ↓                                                    │
│  ② 合约返回 transactionHash（链上存证，但签名在链下完成）    │
│     ↓                                                    │
│  ③ 其他 owner 用 EIP-712 标准对 transactionHash 做链下签名 │
│     ↓                                                    │
│  ④ 收集到足够的签名（≥ 阈值 m）                            │
│     ↓                                                    │
│  ⑤ 任意人将所有签名打包，调用 execTransaction 上链执行      │
│     ↓                                                    │
│  ⑥ 合约验证：签名数量 ≥ 阈值？每个签名都是有效 owner？       │
│     ↓                                                    │
│  ⑦ 全部通过 → 执行交易（call/transfer）                    │
└─────────────────────────────────────────────────────────┘
```

### 2.2 核心设计思想：链下签名 + 链上验证

```
为什么签名放在链下？

❌ 如果每个 owner 都发链上交易来签名：
   5 个人就是 5 笔交易 = 5 次 Gas 费
   流程慢，成本高，体验差

✅ EIP-712 链下签名：
   5 个人在本地签名（免费）
   1 个人把 5 个签名打包上链执行（1 次 Gas 费）
   
   省 80% 的 Gas！
```

---

## 三、代码实现

### 3.1 核心合约（Solidity — 简化版 Gnosis Safe）

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title 简化版多签钱包
 * @notice 实现 m-of-n 多签逻辑，支持 EIP-712 链下签名
 */
contract MultiSigWallet {
    // ====== 存储 ======
    
    address[] public owners;               // 所有 owner 地址
    mapping(address => bool) public isOwner; // 快速判断是否为 owner
    uint256 public threshold;              // 最少需要的签名数（m）
    uint256 public nonce;                  // 防重放，每笔交易递增

    // ====== 事件 ======
    
    event Deposit(address indexed sender, uint256 amount);
    event Submit(uint256 indexed txId);
    event Confirm(address indexed owner, uint256 indexed txId);
    event Execute(uint256 indexed txId);
    event Revoke(address indexed owner, uint256 indexed txId);
    
    // ====== 构造函数 ======
    
    constructor(address[] memory _owners, uint256 _threshold) {
        require(_owners.length >= _threshold, "threshold > owners");
        require(_threshold > 0, "threshold = 0");
        
        for (uint256 i = 0; i < _owners.length; i++) {
            address owner = _owners[i];
            require(owner != address(0), "zero address");
            require(!isOwner[owner], "duplicate owner");
            
            isOwner[owner] = true;
            owners.push(owner);
        }
        
        threshold = _threshold;
    }
    
    // ====== 修饰符 ======
    
    modifier onlyOwner() {
        require(isOwner[msg.sender], "not owner");
        _;
    }
    
    modifier txExists(uint256 _txId) {
        require(_txId < transactions.length, "tx not exist");
        _;
    }
    
    modifier notConfirmed(uint256 _txId) {
        require(!confirmations[_txId][msg.sender], "already confirmed");
        _;
    }
    
    // ====== 接收 ETH ======
    
    receive() external payable {
        emit Deposit(msg.sender, msg.value);
    }

    // ====== 交易数据结构 ======
    
    struct Transaction {
        address to;       // 目标地址
        uint256 value;    // 发送金额（ETH 单位 wei）
        bytes data;       // calldata（如果只是转账则为空）
        bool executed;    // 是否已执行
        uint256 confirmCount; // 已确认数量
    }
    
    Transaction[] public transactions;
    
    // 确认记录：mapping(交易ID => mapping(owner => 是否确认))
    mapping(uint256 => mapping(address => bool)) public confirmations;

    // ====== 核心函数 ======
    
    /**
     * @notice 提交新交易
     * @param _to 目标地址
     * @param _value 金额
     * @param _data calldata（调用合约时使用）
     * @return txId 交易ID
     */
    function submitTransaction(
        address _to,
        uint256 _value,
        bytes memory _data
    ) public onlyOwner returns (uint256 txId) {
        txId = transactions.length;
        
        transactions.push(Transaction({
            to: _to,
            value: _value,
            data: _data,
            executed: false,
            confirmCount: 0
        }));
        
        emit Submit(txId);
    }
    
    /**
     * @notice Owner 确认交易（链上确认 — 高 Gas）
     * @param _txId 交易ID
     */
    function confirmTransaction(
        uint256 _txId
    ) public onlyOwner txExists(_txId) notConfirmed(_txId) {
        Transaction storage tx_ = transactions[_txId];
        
        confirmations[_txId][msg.sender] = true;
        tx_.confirmCount += 1;
        
        emit Confirm(msg.sender, _txId);
        
        // 如果此时达到阈值，直接执行
        if (tx_.confirmCount >= threshold) {
            _executeTransaction(_txId);
        }
    }
    
    /**
     * @notice 链下签名版执行 — 收集签名后一次性提交上链
     * @param _txId 交易ID（用 getTransactionHash 算出）
     * @param _signatures 所有 owner 的签名拼接（每个 65 字节）
     * 
     * 这是 Gnosis Safe 的核心省 Gas 机制：
     * N 个 owner 在链下签名（免费），由 1 个人把签名打包上链执行
     */
    function execTransaction(
        address _to,
        uint256 _value,
        bytes memory _data,
        bytes memory _signatures  // 多个签名拼接
    ) public returns (bool success) {
        // 1. 计算本笔交易的哈希
        bytes32 txHash = getTransactionHash(_to, _value, _data, nonce);
        
        // 2. 验证签名数量
        require(_signatures.length >= threshold * 65, "not enough signatures");
        
        // 3. 逐个验证签名
        address lastOwner = address(0);
        
        for (uint256 i = 0; i < threshold; i++) {
            // 每个签名 65 字节：r(32) + s(32) + v(1)
            bytes32 r;
            bytes32 s;
            uint8 v;
            
            assembly {
                let sigPos := add(_signatures, add(32, mul(i, 65)))
                r := mload(sigPos)
                s := mload(add(sigPos, 32))
                v := byte(0, mload(add(sigPos, 64)))
            }
            
            // EIP-712 消息前缀
            // \x19\x01 = EIP-712 前缀
            // domainSeparator = keccak256(abi.encode(合约名、版本、链ID、合约地址))
            bytes32 digest = keccak256(
                abi.encodePacked("\x19\x01", domainSeparator(), txHash)
            );
            
            address signer = ecrecover(digest, v, r, s);
            
            require(signer != address(0), "invalid signature");
            require(isOwner[signer], "signer not owner");
            require(signer > lastOwner, "signatures not sorted"); // 防重放
            lastOwner = signer;
        }
        
        // 4. 执行交易
        nonce++; // 防重放
        (success, ) = _to.call{value: _value}(_data);
        require(success, "tx failed");
        
        emit Execute(nonce - 1);
    }
    
    /**
     * @notice 计算交易哈希（EIP-712 格式）
     */
    function getTransactionHash(
        address _to,
        uint256 _value,
        bytes memory _data,
        uint256 _nonce
    ) public view returns (bytes32) {
        // keccak256 交易结构体
        return keccak256(
            abi.encode(
                keccak256(
                    abi.encodePacked(
                        "SafeTx(address to,uint256 value,bytes data,uint256 nonce)"
                    )
                ),
                _to,
                _value,
                keccak256(_data),
                _nonce
            )
        );
    }
    
    /**
     * @notice EIP-712 Domain Separator
     */
    function domainSeparator() public view returns (bytes32) {
        return keccak256(
            abi.encode(
                keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
                keccak256(bytes("MultiSigWallet")),
                keccak256(bytes("1")),
                block.chainid,
                address(this)
            )
        );
    }
    
    // ====== 内部函数 ======
    
    function _executeTransaction(uint256 _txId) internal {
        Transaction storage tx_ = transactions[_txId];
        require(!tx_.executed, "already executed");
        
        tx_.executed = true;
        
        (bool success, ) = tx_.to.call{value: tx_.value}(tx_.data);
        require(success, "execution failed");
        
        emit Execute(_txId);
    }
    
    // ====== 视图函数 ======
    
    function getOwners() public view returns (address[] memory) {
        return owners;
    }
    
    function getTransactionCount() public view returns (uint256) {
        return transactions.length;
    }
}
```

---

### 3.2 后端 — Go 调用 Gnosis Safe

```go
package main

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "math/big"
    
    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

// SafeTx EIP-712 结构体
type SafeTx struct {
    To             common.Address
    Value          *big.Int
    Data           []byte
    Operation      uint8   // 0 = Call, 1 = DelegateCall
    SafeTxGas      *big.Int
    BaseGas        *big.Int
    GasPrice       *big.Int
    GasToken       common.Address
    RefundReceiver common.Address
    Nonce          *big.Int
}

// 计算 EIP-712 SafeTx 哈希
func (s *SafeTx) GetTransactionHash(
    chainID *big.Int,
    safeAddress common.Address,
) common.Hash {
    // 1. EIP-712 Domain Separator
    domainSeparator := crypto.Keccak256Hash(
        []byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
        crypto.Keccak256([]byte("Gnosis Safe")),
        crypto.Keccak256([]byte("1")),
        common.LeftPadBytes(chainID.Bytes(), 32),
        common.LeftPadBytes(safeAddress.Bytes(), 32),
    )
    
    // 2. SafeTx 类型哈希
    safeTxTypeHash := crypto.Keccak256([]byte(
        "SafeTx(address to,uint256 value,bytes data,uint8 operation,uint256 safeTxGas,uint256 baseGas,uint256 gasPrice,address gasToken,address refundReceiver,uint256 nonce)",
    ))
    
    // 3. 编码 SafeTx 数据
    safeTxHash := crypto.Keccak256Hash(
        safeTxTypeHash.Bytes(),
        common.LeftPadBytes(s.To.Bytes(), 32),
        common.LeftPadBytes(s.Value.Bytes(), 32),
        crypto.Keccak256(s.Data),
        common.LeftPadBytes([]byte{s.Operation}, 32),
        common.LeftPadBytes(s.SafeTxGas.Bytes(), 32),
        common.LeftPadBytes(s.BaseGas.Bytes(), 32),
        common.LeftPadBytes(s.GasPrice.Bytes(), 32),
        common.LeftPadBytes(s.GasToken.Bytes(), 32),
        common.LeftPadBytes(s.RefundReceiver.Bytes(), 32),
        common.LeftPadBytes(s.Nonce.Bytes(), 32),
    )
    
    // 4. 最终 EIP-712 哈希
    // keccak256("\x19\x01" ‖ domainSeparator ‖ safeTxHash)
    return crypto.Keccak256Hash(
        []byte{0x19, 0x01},
        domainSeparator.Bytes(),
        safeTxHash.Bytes(),
    )
}

// SignSafeTx 使用 EIP-712 签名 Safe 交易
func SignSafeTx(
    safeTx *SafeTx,
    chainID *big.Int,
    safeAddress common.Address,
    privateKey *ecdsa.PrivateKey,
) ([]byte, error) {
    // 1. 计算交易哈希
    txHash := safeTx.GetTransactionHash(chainID, safeAddress)
    
    // 2. EIP-712 签名
    signature, err := crypto.Sign(txHash.Bytes(), privateKey)
    if err != nil {
        return nil, fmt.Errorf("签名失败: %w", err)
    }
    
    // 3. 调整 v 值（EIP-712: v ∈ {27, 28}）
    // crypto.Sign 返回的 v 是 0 或 1，EIP-712 需要 +27
    signature[64] += 27
    
    return signature, nil
}

// BuildExecTransactionData 构造 execTransaction 的 calldata
func BuildExecTransactionData(
    safeTx *SafeTx,
    signatures []byte,
) []byte {
    // Gnosis Safe 的 execTransaction 方法签名
    methodID := crypto.Keccak256([]byte(
        "execTransaction(address,uint256,bytes,uint8,uint256,uint256,uint256,address,address,bytes)",
    ))[:4]
    
    // 将参数 ABI 编码
    // 实际项目中应使用 go-ethereum 的 abi 包来编码
    // 这里展示核心逻辑
    
    return append(methodID, signatures...) // 简化示意
}

// 使用示例：构造并签名一笔多签提现
func ExampleMultiSigWithdraw() {
    client, _ := ethclient.Dial("https://eth.llamarpc.com")
    chainID, _ := client.ChainID(context.Background())
    
    safeAddress := common.HexToAddress("0xYourSafeAddress")
    toAddress := common.HexToAddress("0xRecipientAddress")
    
    // 1. 构造 SafeTx
    safeTx := &SafeTx{
        To:             toAddress,
        Value:          big.NewInt(1000000000000000000), // 1 ETH
        Data:           []byte{},                        // 纯转账，无 calldata
        Operation:      0,                               // Call
        SafeTxGas:      big.NewInt(0),
        BaseGas:        big.NewInt(0),
        GasPrice:       big.NewInt(0),
        GasToken:       common.Address{},
        RefundReceiver: common.Address{},
        Nonce:          big.NewInt(5),                   // 当前 nonce
    }
    
    // 2. 三个 owner 分别用自己私钥签名（链下，不花 Gas）
    privKey1, _ := crypto.HexToECDSA("owner1_private_key")
    privKey2, _ := crypto.HexToECDSA("owner2_private_key")
    privKey3, _ := crypto.HexToECDSA("owner3_private_key")
    
    sig1, _ := SignSafeTx(safeTx, chainID, safeAddress, privKey1)
    sig2, _ := SignSafeTx(safeTx, chainID, safeAddress, privKey2)
    sig3, _ := SignSafeTx(safeTx, chainID, safeAddress, privKey3)
    
    // 3. 拼接签名（按 owner 地址升序排列，防止重放）
    // 每个签名 65 字节：r(32) + s(32) + v(1)
    allSignatures := append(append(sig1, sig2...), sig3...)
    
    fmt.Printf("收集到 %d 个签名，共 %d 字节\n", 3, len(allSignatures))
    // 输出：收集到 3 个签名，共 195 字节
    
    // 4. 构造 execTransaction calldata
    calldata := BuildExecTransactionData(safeTx, allSignatures)
    
    // 5. 任意人（不需要是 owner）调用 Safe 合约执行
    tx := types.NewTransaction(
        0,              // nonce（EOA 的 nonce，不是 Safe 的）
        safeAddress,    // 发给 Safe 合约
        big.NewInt(0),  // 不需要发 ETH 给 Safe
        300000,         // gas limit
        nil,            // gas price（EIP-1559 用 dynamic fee 替代）
        calldata,
    )
    
    fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())
    // 前端或后端用任意一个热钱包签名并广播这笔交易
    _ = tx
}
```

---

### 3.3 前端 — TypeScript 多签交互

```typescript
// ====== 前端调用多签钱包的完整流程 ======

import { encodeFunctionData, encodePacked, keccak256, toBytes } from 'viem'
import { useAccount, useWalletClient, usePublicClient } from 'wagmi'

// ---------- 常量 ----------
const SAFE_ABI = [
  {
    name: 'execTransaction',
    type: 'function',
    inputs: [
      { name: 'to', type: 'address' },
      { name: 'value', type: 'uint256' },
      { name: 'data', type: 'bytes' },
      { name: 'signatures', type: 'bytes' },
    ],
    outputs: [{ name: 'success', type: 'bool' }],
  },
  {
    name: 'getTransactionHash',
    type: 'function',
    inputs: [
      { name: 'to', type: 'address' },
      { name: 'value', type: 'uint256' },
      { name: 'data', type: 'bytes' },
      { name: 'nonce', type: 'uint256' },
    ],
    outputs: [{ name: '', type: 'bytes32' }],
  },
] as const

// ---------- 工具函数 ----------

/**
 * 计算 EIP-712 的 domain separator
 */
function getDomainSeparator(chainId: number, safeAddress: string): `0x${string}` {
  return keccak256(
    encodePacked(
      ['bytes', 'bytes32', 'bytes32', 'uint256', 'address'],
      [
        toBytes('EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)'),
        keccak256(toBytes('Gnosis Safe')),
        keccak256(toBytes('1')),
        BigInt(chainId),
        safeAddress as `0x${string}`,
      ],
    ),
  )
}

/**
 * 计算 SafeTx 的 EIP-712 哈希
 */
function getSafeTxHash(
  to: string,
  value: bigint,
  data: string,
  nonce: bigint,
  chainId: number,
  safeAddress: string,
): `0x${string}` {
  const domainSeparator = getDomainSeparator(chainId, safeAddress)

  const safeTxTypeHash = keccak256(
    toBytes('SafeTx(address to,uint256 value,bytes data,uint256 nonce)'),
  )

  const encodedSafeTx = encodePacked(
    ['bytes32', 'address', 'uint256', 'bytes32', 'uint256'],
    [
      safeTxTypeHash,
      to as `0x${string}`,
      value,
      keccak256(data as `0x${string}`),
      nonce,
    ],
  )

  const safeTxHash = keccak256(encodedSafeTx)

  // 最终 EIP-712 消息哈希
  return keccak256(
    encodePacked(
      ['string', 'bytes32', 'bytes32'],
      ['\x19\x01', domainSeparator, safeTxHash],
    ),
  )
}

// ---------- React Hook ----------

/**
 * 多签交易流程 Hook
 */
function useMultiSig(safeAddress: string, safeNonce: bigint) {
  const { address } = useAccount()
  const { data: walletClient } = useWalletClient()
  const publicClient = usePublicClient()

  /**
   * 步骤 1：构造交易
   */
  const buildTransaction = (to: string, value: bigint, data: string = '0x') => {
    return { to, value, data, nonce: safeNonce }
  }

  /**
   * 步骤 2：当前用户签名（链下，不花 Gas）
   */
  const signTransaction = async (tx: { to: string; value: bigint; data: string; nonce: bigint }) => {
    if (!walletClient || !address) throw new Error('请先连接钱包')

    const chainId = await walletClient.getChainId()

    // 计算 EIP-712 消息哈希
    const messageHash = getSafeTxHash(
      tx.to,
      tx.value,
      tx.data,
      tx.nonce,
      chainId,
      safeAddress,
    )

    // EIP-712 签名
    const signature = await walletClient.signTypedData({
      domain: {
        name: 'Gnosis Safe',
        version: '1',
        chainId,
        verifyingContract: safeAddress as `0x${string}`,
      },
      types: {
        SafeTx: [
          { name: 'to', type: 'address' },
          { name: 'value', type: 'uint256' },
          { name: 'data', type: 'bytes' },
          { name: 'nonce', type: 'uint256' },
        ],
      },
      primaryType: 'SafeTx',
      message: {
        to: tx.to as `0x${string}`,
        value: tx.value,
        data: tx.data as `0x${string}`,
        nonce: tx.nonce,
      },
    })

    return {
      signer: address,
      signature,
      txHash: messageHash,
    }
  }

  /**
   * 步骤 3：收集到足够签名后，打包上链执行
   */
  const executeTransaction = async (
    tx: { to: string; value: bigint; data: string },
    signatures: string[],  // 每个签名 65 字节 hex
  ) => {
    if (!walletClient) throw new Error('请先连接钱包')

    // 拼接所有签名
    const packedSignatures = signatures
      .map(s => s.startsWith('0x') ? s.slice(2) : s)
      .join('')

    // 调用 Safe 合约的 execTransaction
    const calldata = encodeFunctionData({
      abi: SAFE_ABI,
      functionName: 'execTransaction',
      args: [
        tx.to as `0x${string}`,
        tx.value,
        tx.data as `0x${string}`,
        `0x${packedSignatures}` as `0x${string}`,
      ],
    })

    const hash = await walletClient.sendTransaction({
      to: safeAddress as `0x${string}`,
      data: calldata,
    })

    // 等待上链
    const receipt = await publicClient.waitForTransactionReceipt({ hash })
    return receipt
  }

  return { buildTransaction, signTransaction, executeTransaction }
}

// ---------- UI 组件 ----------

/**
 * 多签提现 UI 组件
 */
function MultiSigWithdraw({
  safeAddress,
  owners,
  threshold,
  currentNonce,
}: {
  safeAddress: string
  owners: string[]
  threshold: number
  currentNonce: bigint
}) {
  const [to, setTo] = useState('')
  const [amount, setAmount] = useState('')
  const [signatures, setSignatures] = useState<string[]>([])
  const [status, setStatus] = useState<'idle' | 'pending' | 'signing' | 'executing' | 'done'>('idle')

  const { buildTransaction, signTransaction, executeTransaction } = useMultiSig(
    safeAddress,
    currentNonce,
  )

  // 当前用户签名
  const handleSign = async () => {
    setStatus('signing')
    const tx = buildTransaction(to, parseEther(amount))
    const result = await signTransaction(tx)

    setSignatures(prev => {
      // 防止重复签名
      if (prev.find(s => s === result.signature)) return prev
      return [...prev, result.signature]
    })
    setStatus('idle')
  }

  // 达到阈值后执行
  const handleExecute = async () => {
    setStatus('executing')
    const tx = buildTransaction(to, parseEther(amount))
    await executeTransaction(tx, signatures)
    setStatus('done')
  }

  const confirmedCount = signatures.length
  const canExecute = confirmedCount >= threshold

  return (
    <div style={{ maxWidth: 500, margin: '0 auto' }}>
      <h3>多签提现</h3>

      {/* Safe 信息 */}
      <div style={{ background: '#f0f0f0', padding: 12, borderRadius: 8, marginBottom: 16 }}>
        <p>多签地址: {safeAddress}</p>
        <p>阈值: {confirmedCount} / {threshold} (已确认 / 最少需要)</p>
        <p>Owner 数量: {owners.length}</p>
      </div>

      {/* 提现表单 */}
      <input
        placeholder="接收地址"
        value={to}
        onChange={e => setTo(e.target.value)}
        style={{ width: '100%', padding: 8, marginBottom: 8 }}
      />
      <input
        placeholder="金额 (ETH)"
        value={amount}
        onChange={e => setAmount(e.target.value)}
        style={{ width: '100%', padding: 8, marginBottom: 8 }}
      />

      {/* 进度条 */}
      <div style={{ background: '#e0e0e0', height: 8, borderRadius: 4, marginBottom: 16 }}>
        <div
          style={{
            width: `${(confirmedCount / threshold) * 100}%`,
            height: '100%',
            background: canExecute ? '#4caf50' : '#2196f3',
            borderRadius: 4,
            transition: 'width 0.3s',
          }}
        />
      </div>

      {/* 操作按钮 */}
      <div style={{ display: 'flex', gap: 8 }}>
        <button onClick={handleSign} disabled={status !== 'idle'} style={{ flex: 1 }}>
          {status === 'signing' ? '签名中...' : '我签名（免费）'}
        </button>

        <button
          onClick={handleExecute}
          disabled={!canExecute || status === 'executing'}
          style={{ flex: 1, background: canExecute ? '#4caf50' : '#ccc' }}
        >
          {status === 'executing'
            ? '执行中...'
            : confirmedCount < threshold
              ? `还需 ${threshold - confirmedCount} 人签名`
              : '上链执行'}
        </button>
      </div>

      {/* 已签名的 owner */}
      {signatures.length > 0 && (
        <div style={{ marginTop: 16 }}>
          <p>已签名：{confirmedCount} / {threshold}</p>
          <ul>
            {signatures.map((sig, i) => (
              <li key={i}>签名 {i + 1}: {sig.slice(0, 20)}...</li>
            ))}
          </ul>
        </div>
      )}

      {status === 'done' && (
        <div style={{ marginTop: 16, color: '#4caf50' }}>交易已成功执行！</div>
      )}
    </div>
  )
}
```

---

## 四、安全要点

### 4.1 签名排序防重放

```solidity
// ⚠️ 签名必须按 owner 地址升序排列
// 目的：防止同一个签名被重复使用（重放攻击）

// ❌ 不排序的验证方式有漏洞：
// signatures = [sig_A, sig_B, sig_C]
// 攻击者可提取 sig_A + sig_C 去签另一笔交易（如果有另一组 2/3 阈值的话）

// ✅ 要求签名地址升序：
// 1. 前端/后端在拼接签名时，按 owner 地址从小到大排序
// 2. 合约中检查 signer > lastOwner
require(signer > lastOwner, "signatures not sorted");
```

### 4.2 Nonce 防跨交易重放

```solidity
// 每笔交易有唯一的 nonce
// nonce 执行后递增，同一 nonce 的交易无法再次执行

bytes32 txHash = getTransactionHash(_to, _value, _data, nonce);
// ... 验证签名 ...
nonce++; // ← 执行后递增，链上记录，同一 nonce 再也验证不过
```

### 4.3 EIP-712 防钓鱼

```
为什么必须是 EIP-712？

❌ 普通 eth_sign：
   MetaMask 显示一串 hex 哈希
   用户根本不知道在签什么
   容易被钓鱼——以为是登录，其实在签名转账

✅ EIP-712 signTypedData：
   MetaMask 显示结构化数据：
   ┌──────────────────────────────┐
   │ Gnosis Safe                  │
   │ to: 0xRecipient...           │
   │ value: 1 ETH                 │
   │ nonce: 5                     │
   └──────────────────────────────┘
   用户能看到"我在签什么" ← 防钓鱼
```

---

## 五、话术总结

### 30 秒口述版

> "多签就是 m-of-n 模式，n 个 owner 中至少 m 个同意才能执行交易。以 Gnosis Safe 为例，核心设计是链下签名+链上验证——各 owner 用 EIP-712 对交易哈希做链下签名，免费且 MetaMask 能显示结构化数据防钓鱼；收集到 m 个签名后，任意人把签名拼接上链调用 execTransaction，合约验证签名数量和有效性后执行。整个过程只花一次 Gas。HashKey 这类持牌交易所的资产管理大概率走 3/5 多签，私钥分散保管，单人不等于控制资产。"

### 面试追问应对

**Q: 如果 owner 丢失私钥怎么办？**

> "分两种情况。如果剩余 owner 数量仍 ≥ 阈值，可以提交一笔交易把丢失的 owner 替换成新地址。如果剩余 owner 数量不足阈值，那就死锁了——这就是为什么多签方案里阈值不要设太高（比如 5/5），一般用 3/5 或 5/8 留缓冲。也有项目引入恢复机制——一个冷存储的超级管理员可以在超长时间锁后介入。"

**Q: 和 MPC（多方计算）钱包比有什么优劣？**

> "多签是智能合约层的方案，链上透明可审计，但每笔交易仍上链有 Gas 成本。MPC 是把私钥分片，各持一份碎片，签名时通过多方计算协议拼出完整签名，结果就是一笔普通的单签交易，链上无法区分，隐私更好也省 Gas。但 MPC 的算法复杂，碎片轮换和恢复机制比多签门槛高。很多交易所实际是 MPC + 多签混用。"
