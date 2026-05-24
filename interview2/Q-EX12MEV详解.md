# Q-EX12: MEV（矿工可提取价值）详解

> **适用场景**：HashKey 全栈工程师面试 — Web3 安全进阶题  
> **难度**：中等偏难，理解链上交易生命周期是关键  
> **关联题目**：[Q10 智能合约前端交互安全](../week1/hashkey-fullstack-interview-questions.md#q10-智能合约前端交互中如何防范常见攻击)、[Q9 杠杆交易订单系统设计](../week1/hashkey-fullstack-interview-questions.md#q9-如何设计一个安全的高频交易杠杆交易订单系统)

---

## 一、什么是 MEV？

MEV（Maximal Extractable Value，最大可提取价值），原名 Miner Extractable Value（矿工可提取价值），是以太坊转向 PoS 后改的名——因为现在不是矿工而是验证者（Validator）在操作。

**一句话定义**：

> MEV 是区块提议者（验证者）通过**重新排序、插入或审查**区块内的交易来获取的额外利润。

### 为什么会有 MEV？

核心原因是以太坊的交易排序机制：

```
用户提交交易 → 交易进入内存池（mempool，公开可见）
                  ↓
         验证者从 mempool 挑选交易
         决定谁先谁后、谁进谁不进
                  ↓
         打包成区块，上链
```

**任何人都能看到 mempool 里待处理的交易**，这就是 MEV 的根源。就像你在牌桌上把自己的牌亮给别人看，别人就能针对你的动作做出反应。

---

## 二、MEV 的三种主要攻击类型

### 2.1 三明治攻击（Sandwich Attack）— 最常见

**攻击过程**（以 DEX 大额买入为例）：

```
时间线 →
1. 你提交了一笔大额买入 ETH 的交易（mempool 公开可见）
2. 攻击者看到你的交易，抢在你前面买入 ETH（推高价格）
3. 你的交易成交（因为价格已被推高，你买贵了）
4. 攻击者立即卖出（赚取价差）

结果：你多付了钱，攻击者赚到了差价
```

**图示**：

```
价格 ↑
  │     攻击者买入    你的交易    攻击者卖出
  │      ┌─┐         ┌─┐        ┌─┐
  │      │ │   ┌───┐ │ │  ┌───┐ │ │
  │      │ │   │   │ │ │  │   │ │ │
  │  ────┘ └───┘   └─┘ └──┘   └─┘ └───
  │  正常价格 ↑      ↑ 你买贵了  ↑
  │              你的交易
  └──────────────────────────────────────→ 时间
```

**代码示例 — 攻击者合约（Solidity）**：

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// ⚠️ 仅用于教学，理解攻击原理
contract SandwichBot {
    address private owner;
    IUniswapV2Router private router;
    
    constructor(address _router) {
        owner = msg.sender;
        router = IUniswapV2Router(_router);
    }
    
    // 攻击者在 mempool 中检测到受害者的大额交易后，发起此函数
    function sandwichAttack(
        address victimToken,    // 受害者要买的代币
        address baseToken,      // WETH / USDC
        uint256 victimAmount,   // 受害者买入金额（从 mempool 解码得到）
        uint256 frontrunAmount  // 攻击者抢跑金额
    ) external {
        require(msg.sender == owner, "not owner");
        
        // 第一步：抢跑买入（Frontrun）
        // 抢在受害者交易之前买入，推高价格
        address[] memory path = new address[](2);
        path[0] = baseToken;
        path[1] = victimToken;
        
        router.swapExactETHForTokens{value: frontrunAmount}(
            0,              // 先不设最小输出（实际要设滑点保护自身的攻击交易）
            path,
            address(this),
            block.timestamp + 300
        );
        
        // 受害者交易在此处被打包（推高价格后）
        // ... 受害者以更高的价格买入 ...
        
        // 第二步：卖出获利（Backrun）
        // 受害者的交易推高价格后，攻击者卖出
        path[0] = victimToken;
        path[1] = baseToken;
        
        IERC20(victimToken).approve(address(router), type(uint256).max);
        router.swapExactTokensForETH(
            IERC20(victimToken).balanceOf(address(this)),
            0,              // 同理
            path,
            address(this),
            block.timestamp + 300
        );
        
        // 利润 = 卖出所得 - 买入花费 - Gas 费
    }
    
    receive() external payable {}
}
```

---

### 2.2 抢跑（Frontrunning）— 套利机会被抢先

**场景**：你发现了一个套利机会——Uniswap 上 ETH 比 Sushiswap 便宜 0.5%，你提交套利交易。攻击者的 bot 检测到你的交易，复制后给更高 Gas 费，抢先执行，你的交易因为价格变动而失败。

**代码示例 — 套利 bot 检测逻辑（TypeScript/Node.js）**：

```typescript
// ⚠️ 仅用于教学，理解抢跑原理
import { createPublicClient, http, parseAbiItem } from 'viem'
import { mainnet } from 'viem/chains'

const publicClient = createPublicClient({
  chain: mainnet,
  transport: http('https://eth.llamarpc.com'),
})

// 监听 mempool 中待处理的 Swap 交易
async function watchMempool() {
  // 注意：普通 RPC 无法直接访问 mempool
  // 实际中攻击者自建节点或使用 Flashbots MEV-Geth 等私有客户端
  publicClient.watchPendingTransactions({
    onTransactions: async (hashes) => {
      for (const hash of hashes) {
        const tx = await publicClient.getTransaction({ hash })
        
        // 检测是否为对 Uniswap Router 的大额 Swap
        if (isLargeSwapOnUniswap(tx)) {
          const profit = await calculateFrontrunProfit(tx)
          
          if (profit > gasCost) {
            // 构造抢跑交易，设更高 Gas 费抢先执行
            const frontrunTx = await buildFrontrunTransaction(tx)
            // 用比原交易高 10% 的 Gas 费发送
            await sendWithHigherGas(frontrunTx, tx.gasPrice! * 110n / 100n)
          }
        }
      }
    },
  })
}

function isLargeSwapOnUniswap(tx: any): boolean {
  // 检查 to 地址是否为 Uniswap Router
  // 检查 value 或 calldata 中的金额是否够大
  const UNISWAP_ROUTER = '0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D'
  return tx.to?.toLowerCase() === UNISWAP_ROUTER.toLowerCase()
    && (tx.value > parseEther('10') || decodeAmountFromCalldata(tx.input) > parseEther('10'))
}

// 计算抢跑利润
async function calculateFrontrunProfit(victimTx: any): Promise<bigint> {
  // 1. 模拟受害者交易，计算价格影响
  // 2. 估算抢跑买入 + 受害者买入后卖出的利润
  // 3. 减去 Gas 费
  // 返回预期利润
  // ...
  return 0n
}
```

---

### 2.3 清算抢跑（Liquidation Frontrunning）

**场景**：在借贷协议（如 Aave、Compound）中，当某个仓位的抵押率低于清算阈值时，任何人都可以调用清算函数获得清算奖金（通常 5%-15%）。攻击者通过监控链上价格预言机更新，在其他清算者之前抢先执行清算。

```solidity
// 攻击者合约 — 抢先调用 Aave 的清算函数
contract LiquidationBot {
    IAaveLendingPool private pool = IAaveLendingPool(0x7d2...);

    function liquidate(address collateral, address debt, address user, uint256 amount) external {
        // 1. 原子套利：在同一个交易里先 Flash Loan 借币
        // 2. 用借来的币触发清算，拿到折扣抵押品
        // 3. 卖出抵押品，归还 Flash Loan，剩下的全是利润
        
        // 用高 Gas 费确保自己的交易在别人前面被打包
        pool.liquidationCall(collateral, debt, user, amount, false);
    }
}
```
现实中的场景

你用 1 ETH（值 $2000）借了 1000 USDC
这 1000 USDC 你已经花掉了（买其他币 / 付钱了）
手里没现金了

ETH 突然跌到 $1200

你想自救 → 需要 1000 USDC 来还债 → 但你手里没有！
想卖 ETH 还债 → 可 ETH 押在协议里，必须先还钱才能取出来
                   ↑
              死循环：没钱还 → 取不出 ETH → 没法卖 ETH 换钱

这就是清算存在的意义：
  你被套住了 → 协议允许别人替你还 → 他们拿走打折的 ETH 作为报酬
清算后

清算人帮你还了 500 USDC
拿走价值 ~$550 的 ETH（奖励 ~10%）

你还剩：
  约 0.54 ETH 的抵押品（可以取回来了）
  还欠 500 USDC
  仓位安全了（抵押率恢复正常）

清算人赚了 $50
协议避免了坏账
你虽然亏了点抵押品，但仓位没全爆


---

## 三、防护手段（面试重点）

### 3.1 滑点保护（Slippage Protection）— 第一道防线

**原理**：设置交易可接受的最小输出量，如果实际成交低于这个值，交易自动回滚。三明治攻击会导致价格偏离，滑点保护能阻止交易在价格被操纵时成交。

**前端代码（TypeScript + Wagmi/Viem）**：

```typescript
// 计算最小输出量 = 预期输出 × (1 - 滑点百分比)
async function swapWithSlippage(
  tokenIn: string,
  tokenOut: string,
  amountIn: bigint,
  slippagePercent: number = 0.5  // 默认 0.5%
) {
  // 1. 获取链上报价
  const amounts = await publicClient.readContract({
    address: UNISWAP_ROUTER,
    abi: uniswapRouterAbi,
    functionName: 'getAmountsOut',
    args: [amountIn, [tokenIn, tokenOut]],
  })
  const expectedOut = (amounts as bigint[])[1]

  // 2. 计算最小接受量（扣除滑点）
  const minAmountOut = expectedOut * BigInt(Math.floor((1 - slippagePercent / 100) * 1000))
    / 1000n

  console.log(`预期输出: ${expectedOut}, 最小接受: ${minAmountOut}`)

  // 3. 发送交易，带 minAmountOut 保护
  const hash = await walletClient.sendTransaction({
    to: UNISWAP_ROUTER,
    data: encodeFunctionData({
      abi: uniswapRouterAbi,
      functionName: 'swapExactTokensForTokens',
      args: [amountIn, minAmountOut, [tokenIn, tokenOut], userAddress, DEADLINE],
    }),
  })

  return hash
}
```

**合约层对应代码（Solidity）**：

```solidity
function swap(uint256 amountIn, uint256 minAmountOut, address[] calldata path) external {
    uint256[] memory amounts = router.swapExactTokensForTokens(
        amountIn,
        minAmountOut,  // ← 关键：如果实际输出 < minAmountOut，整个交易 revert
        path,
        msg.sender,
        block.timestamp + 300  // 30 分钟过期
    );
    
    // Uniswap 内部会做这个检查：
    // require(amounts[1] >= minAmountOut, "INSUFFICIENT_OUTPUT_AMOUNT");
}
```

**前端根据用户等级动态调整滑点**：

```typescript
// 根据交易金额、代币流动性动态建议滑点
function suggestSlippage(
  amountInUSD: number,
  tokenLiquidity: 'high' | 'medium' | 'low'
): number {
  // 小额、高流动性 — 低滑点
  if (amountInUSD < 1000 && tokenLiquidity === 'high') return 0.1
  
  // 中额 — 中等滑点
  if (amountInUSD < 10000) return 0.5
  
  // 大额或低流动性 — 高滑点
  if (amountInUSD >= 10000 || tokenLiquidity === 'low') return 1.0
  
  return 0.5 // 默认
}

// 用户界面
function SwapForm() {
  const [slippage, setSlippage] = useState(0.5)
  
  return (
    <div>
      <label>滑点容忍度</label>
      <select value={slippage} onChange={e => setSlippage(+e.target.value)}>
        <option value={0.1}>0.1%（低）</option>
        <option value={0.5}>0.5%（推荐）</option>
        <option value={1.0}>1.0%（高）</option>
      </select>
      {slippage >= 1.0 && (
        <p style={{ color: 'orange' }}>⚠ 高滑点设置，容易遭受三明治攻击</p>
      )}
    </div>
  )
}
```

---

### 3.2 Flashbots — 绕过公开内存池

**原理**：普通交易提交到公开 mempool，所有人都能看到。Flashbots 提供私有交易中继（Relay），你的交易直接发给区块提议者，不在公开 mempool 中曝光，MEV 搜索者看不到你的交易。

**架构对比**：

```
普通交易路径（透明，可被监控）：
  用户 → 公开 mempool → 区块提议者 → 上链
           ↑
    攻击者能看到你的交易 ✗

Flashbots 交易路径（私有）：
  用户 → Flashbots Relay → 区块提议者 → 上链
         （加密通道，mempool 中不可见）
                              ↑
                    攻击者看不到 ✓
```

**前端代码（TypeScript + Flashbots SDK）**：

```typescript
// ❌ 普通发送 — 交易暴露在 mempool
const hash = await walletClient.sendTransaction({
  to: UNISWAP_ROUTER,
  data: swapCalldata,
  gas: 300000n,
})

// ✅ Flashbots 私有发送 — mempool 中不可见
import { FlashbotsBundleProvider } from '@flashbots/ethers-provider-bundle'

async function sendViaFlashbots(tx: TransactionRequest): Promise<string> {
  // 1. 初始化 Flashbots Provider
  const flashbotsProvider = await FlashbotsBundleProvider.create(
    provider,        // 普通 ethers provider
    Wallet.createRandom(),  // Flashbots 签名钱包（不影响交易本身）
    'https://relay.flashbots.net'  // Flashbots Relay 地址
  )

  // 2. 构造交易 Bundle（可以包含多笔交易打包执行）
  const signedTx = await wallet.signTransaction({
    ...tx,
    chainId: 1,
    // 关键：设 maxFeePerGas 和 maxPriorityFeePerGas
    // Flashbots 要求按 EIP-1559 格式
  })

  const bundle = [
    {
      signedTransaction: signedTx,
    },
  ]

  // 3. 模拟交易（确认不会失败）
  const targetBlock = await provider.getBlockNumber() + 5
  const simulation = await flashbotsProvider.simulate(bundle, targetBlock)
  
  if ('error' in simulation) {
    throw new Error(`Flashbots 模拟失败: ${simulation.error.message}`)
  }
  console.log(`模拟成功，预估 Gas: ${simulation.gasUsed}`)

  // 4. 发送私有交易 — 不在公开 mempool 中曝光
  const flashbotsTransactionResponse = await flashbotsProvider.sendBundle(
    bundle,
    targetBlock  // 目标区块号
  )

  // 5. 等待上链
  const resolution = await flashbotsTransactionResponse.wait()
  
  if (resolution === FlashbotsBundleResolution.BundleIncluded) {
    console.log('交易通过 Flashbots 成功上链')
    return signedTx.hash!
  } else {
    // 如果这个区块没被包含，可以重试下一个区块
    throw new Error(`Flashbots 未包含: ${resolution}`)
  }
}

// 使用示例：大额 Swap 走 Flashbots
function SwapButton({ amountIn }: { amountIn: bigint }) {
  const handleSwap = async () => {
    const calldata = buildSwapCalldata(amountIn)

    // 大额交易（> 10 ETH）走 Flashbots
    if (amountIn > parseEther('10')) {
      await sendViaFlashbots({
        to: UNISWAP_ROUTER,
        data: calldata,
        gasLimit: 300000n,
      })
    } else {
      // 小额交易正常发
      await walletClient.sendTransaction({
        to: UNISWAP_ROUTER,
        data: calldata,
      })
    }
  }

  return <button onClick={handleSwap}>兑换</button>
}
```

**Flashbots 的局限性**：
- 不是所有验证者都接入了 Flashbots Relay（目前约 90%+ 覆盖率）
- 交易不一定在下一个区块就被包含，可能需要等待几个区块
- 需要验证者支持，不能 100% 保证交易不被公开 mempool 看到

---

### 3.3 TWAP — 时间加权平均价格

**原理**：使用一段时间内的平均价格作为参考价，而不是当前瞬间价格。攻击者要在足够长的时间内维持价格偏离，成本极高。

```
即时价格（容易被操纵）：
  ┌──┐
  │  │ ← 攻击者在一个区块内瞬间拉高价格
──┘  └──────
  ↑ 你的交易读到这个假价格

TWAP（难以操纵）：
──────────────────────
  ↑ 30 分钟的平均值，拉高一个点影响很小
```

**合约代码（Solidity — Uniswap V2 TWAP 实现）**：

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// Uniswap V2 内置的 TWAP 机制
contract TWAPExample {
    IUniswapV2Pair public pair;
    
    // 记录价格累积值和时间戳
    uint256 public price0CumulativeLast;
    uint256 public price1CumulativeLast;
    uint32 public blockTimestampLast;
    
    // Uniswap V2 的 TWAP 核心：
    // 每个区块都累加 price * timeElapsed，相当于积分
    // priceAverage = (priceCumulative(now) - priceCumulative(then)) / (now - then)
    
    function update() external {
        // 每个区块调用一次，获取最新的价格累积值
        (uint price0Cumulative, uint price1Cumulative, uint32 blockTimestamp) =
            UniswapV2OracleLibrary.currentCumulativePrices(address(pair));
        
        // 两次调用的差值 / 时间差 = 时间加权平均价格
        uint32 timeElapsed = blockTimestamp - blockTimestampLast;
        
        if (timeElapsed > 0) {
            // 价格已按时间加权累加
            price0CumulativeLast = price0Cumulative;
            price1CumulativeLast = price1Cumulative;
            blockTimestampLast = blockTimestamp;
        }
    }
    
    // 查询 30 分钟 TWAP
    function consult(address tokenIn, uint amountIn) external view returns (uint amountOut) {
        // Uniswap 官方库提供的 TWAP 查询
        // 内部使用上面累积的 priceCumulative 计算 30 分钟均价
        return UniswapV2OracleLibrary.consult(address(pair), tokenIn, amountIn, 30 minutes);
    }
}
```

**前端使用 TWAP 获取报价（TypeScript）**：

```typescript
// 从 Uniswap V2 获取 TWAP 价格（用于价格展示、清算判断等）
async function getTWAPPrice(
  pairAddress: string,
  tokenIn: string,
  amountIn: bigint,
  periodSeconds: number = 1800  // 默认 30 分钟
): Promise<bigint> {
  // 调用 Uniswap V2 的 consult 方法
  // 该方法内部计算 (priceCumulativeLast - prevPriceCumulative) / timeElapsed
  const amountOut = await publicClient.readContract({
    address: UNISWAP_V2_ROUTER,
    abi: uniswapRouterAbi,
    functionName: 'getAmountsOut',
    args: [amountIn, [tokenIn, WETH_ADDRESS]],
  })

  // 对于需要 TWAP 的场景（如清算判断），使用专门的 Oracle 合约
  const twapPrice = await publicClient.readContract({
    address: TWAP_ORACLE,
    abi: oracleAbi,
    functionName: 'consult',
    args: [tokenIn, amountIn],
  })

  return twapPrice as bigint
}
```

**对比**：

| 价格类型 | 响应速度 | 防操纵 | 适用场景 |
|---------|---------|--------|---------|
| 即时价格（Spot） | 实时 | 弱 | 小额 Swap 展示 |
| TWAP（30min） | 30 分钟滞后 | 强 | 清算阈值判断、资产估值 |
| TWAP（1h） | 1 小时滞后 | 最强 | 借贷协议抵押率计算 |

---

### 3.4 链下限价单 — 交易上链前不可见

**原理**：用户将限价单签名后交给链下中继器（如 Gelato、Keep3r），中继器在条件满足时代为执行。因为交易条件（价格、数量）在匹配之前不在 mempool 中曝光，无法被抢跑。

```
传统限价单（透明）：
  用户 → 提交限价单到链上合约（mempool 公开）
         → 所有人都能看到你的挂单信息
         → 可能被抢跑 / 三明治

链下限价单（私有）：
  用户 → 签名链下消息 → 发送给 Gelato 中继器
         → mempool 中不可见
         → 中继器在价格触及时帮你执行
         → 交易上链时已经是成交，没有抢跑空间
```

**前端代码 — 通过 Gelato 提交链下限价单**：

```typescript
import { GelatoRelay } from '@gelatonetwork/relay-sdk'

const relay = new GelatoRelay()

async function submitLimitOrder(
  tokenIn: string,
  tokenOut: string,
  amountIn: bigint,
  targetPrice: number,  // 触发价格
  userAddress: string
) {
  // 1. 构造交易数据
  const swapData = encodeFunctionData({
    abi: uniswapRouterAbi,
    functionName: 'swapExactTokensForTokens',
    args: [
      amountIn,
      0, // minAmountOut — Gelato 执行时会实时计算
      [tokenIn, tokenOut],
      userAddress,
      Math.floor(Date.now() / 1000) + 3600, // 1 小时过期
    ],
  })

  // 2. 构造 Gelato 任务
  const request = {
    chainId: 1,
    target: UNISWAP_ROUTER,
    data: swapData,
    user: userAddress,
    // 关键：只在条件满足时执行
    // Gelato 的 resolver 合约实时检查链上条件
  }

  // 3. 签名（用户不需要发送交易到 mempool！）
  // 链下 EIP-712 签名，不消耗 Gas
  const signature = await walletClient.signTypedData({
    // Gelato 的 EIP-712 结构
    domain: { name: 'Gelato', version: '1', chainId: 1, verifyingContract: GELATO_CONTRACT },
    types: { /* Gelato 任务签名类型 */ },
    primaryType: 'Task',
    message: request,
  })

  // 4. 提交给 Gelato（链下 HTTP 请求）
  const response = await fetch('https://api.gelato.digital/tasks', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ request, signature }),
  })

  console.log('限价单已提交给 Gelato，mempool 中不可见')
  return response.json()
}
```

---

## 四、完整防护体系（总结）

```
防护层次（由浅到深）：

第一层  滑点保护           → 防价格偏离过大，最常见、最基本
第二层  交易模拟           → 发送前本地模拟，提前发现失败
第三层  Gas Limit 上限     → 防止 Gas 消耗超预期
第四层  Flashbots 私有池   → 绕过公开 mempool，交易不可见
第五层  TWAP 价格源        → 用时间加权均价，操纵成本高
第六层  链下限价单         → 交易条件不上链，极致隐私
```

**面试口述话术（30 秒版）**：

> "MEV 的根源是 mempool 交易公开可见，导致三明治攻击和抢跑。防护分六层——前端设滑点保护（最小输出量 ≥ 预期的 99.5%），发送前做交易模拟验证，大额交易走 Flashbots 私有 Relay 绕过公开 mempool，价格反馈用 TWAP 时间加权均价而非即时价，敏感操作走链下限价单中继器比如 Gelato，再加一层 Gas Limit 上限防意外消耗。六道防线层层递进，从防价格偏离到完全隐藏交易意图。"

---

## 五、面试常见追问

### Q: 滑点设多少合适？

> "看代币流动性和交易金额。主流币（ETH/USDC）小额设 0.1%，中额设 0.5%；土狗或低流动性代币设 1%-3%。原则是在交易失败风险和价格损失之间取平衡——滑点太低交易容易被抢跑交易挤掉而失败，太高则容易被三明治攻击。"

### Q: Flashbots 的缺点是什么？

> "一是延迟——不保证下一个区块就包含你的交易，可能需要等 1-3 个区块；二是覆盖率——大约 90%+ 的验证者接入，不覆盖全部；三是中心化争议——Flashbots Relay 本身是中心化基础设施，虽然开源透明但仍是单点。"

### Q: MEV 完全是坏事吗？

> "不完全是。MEV 也有正向的一面——比如清算 bot 保证了借贷协议坏账能被及时清算，维持了协议的偿付能力；套利 bot 让不同 DEX 之间的价格趋于一致。问题不在于 MEV 本身，而在于提取 MEV 的手段是否损害普通用户利益。三明治攻击损害用户，但清算 bot 反而保护了协议。"

---

## 六、参考资源

- [Flashbots 官方文档](https://docs.flashbots.net)
- [Uniswap V2 TWAP Oracle 实现](https://docs.uniswap.org/contracts/v2/guides/smart-contract-integration/building-an-oracle)
- [Ethereum.org — MEV 解释](https://ethereum.org/en/developers/docs/mev/)
- [Gelato Network — 链下自动化](https://docs.gelato.network)
