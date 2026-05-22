# 大型 Web3 项目状态管理：四层分层策略

> 核心原则：**服务端状态与客户端状态分离管理**，不同生命周期的数据用不同的工具。

---

## 目录

- [为什么需要分层？](#为什么需要分层)
- [第一层：钱包连接状态 → Context](#第一层钱包连接状态--context)
- [第二层：多链配置 → Zustand](#第二层多链配置--zustand)
- [第三层：合约读写缓存 → RTK Query](#第三层合约读写缓存--rtk-query)
- [第四层：交易临时状态 → useState / useReducer](#第四层交易临时状态--usestate--usereducer)
- [完整架构图](#完整架构图)
- [各层对比速查表](#各层对比速查表)

---

## 为什么需要分层？

在 Web3 项目中，状态有四种完全不同的生命周期：

| 状态类型 | 生命周期 | 举例 |
|---------|---------|------|
| 钱包连接 | 应用级、跨所有页面 | 当前地址、是否已连接、链 ID |
| 全局配置 | 应用级、用户切换后变化 | 支持的链列表、Gas 价格、语言偏好 |
| 链上数据 | 服务端持有、需缓存和刷新 | 代币余额、交易历史、合约查询结果 |
| UI 临时态 | 组件级、关闭即销毁 | 交易弹窗开关、输入框内容、loading 态 |

**如果全部扔进一个 Store**，会导致：
- 服务端数据（余额）和客户端数据（弹窗开关）混在一起，难以追踪数据来源
- 缓存失效策略不清晰——什么时候该重新请求链上数据？
- 组件卸载后残留状态（弹窗关了但数据还在 Store 里）
- 不必要的重渲染——一个输入框变化触发整个 Store 订阅者更新

---

## 第一层：钱包连接状态 → Context

### 为什么用 Context 而不是 Redux/Zustand？

- **数据量小**：通常只有 `address`、`chainId`、`isConnected` 几个字段
- **极少变化**：用户不会每秒切换钱包，可能整个会话只变一两次
- **跨组件共享**：导航栏、Swap 页、资产页都需要知道当前钱包
- **不需要中间件/DevTools**：钱包状态没有复杂的 action 流需要调试
- **Context 值变化只重渲染消费者**：配合 `useContext` 精准订阅

### 为什么不把所有东西都放 Context？

Context 的值一旦变化，**所有消费者都重渲染**。如果把 Gas 价格（每 10 秒变一次）也放 Context，导航栏也会跟着每秒渲染，性能直接爆炸。

### 代码示例

```typescript
// contexts/WalletContext.tsx
import { createContext, useContext, useCallback } from 'react'
import { useAccount, useConnect, useDisconnect, useChainId, useSwitchChain } from 'wagmi'

interface WalletContextType {
  /** 当前钱包地址，未连接时为 undefined */
  address: string | undefined
  /** 当前链 ID（1=以太坊主网, 56=BSC） */
  chainId: number | undefined
  /** 是否已连接钱包 */
  isConnected: boolean
  /** 连接钱包（自动弹出 MetaMask） */
  connect: () => Promise<void>
  /** 断开连接 */
  disconnect: () => void
  /** 切换链 */
  switchChain: (chainId: number) => void
}

const WalletContext = createContext<WalletContextType>({} as WalletContextType)

export function WalletProvider({ children }: { children: React.ReactNode }) {
  const { address, isConnected } = useAccount()
  const { connectAsync, connectors } = useConnect()
  const { disconnect } = useDisconnect()
  const chainId = useChainId()
  const { switchChain } = useSwitchChain()

  const connect = useCallback(async () => {
    const metamask = connectors.find(c => c.id === 'metaMask')
    if (metamask) {
      await connectAsync({ connector: metamask })
    }
  }, [connectors, connectAsync])

  const value: WalletContextType = {
    address,
    chainId,
    isConnected,
    connect,
    disconnect,
    switchChain,
  }

  return (
    <WalletContext.Provider value={value}>
      {children}
    </WalletContext.Provider>
  )
}

/** 快捷 Hook — 任何组件想拿钱包信息只需调用它 */
export function useWallet() {
  return useContext(WalletContext)
}
```

### 使用示例

```typescript
// components/WalletButton.tsx
function WalletButton() {
  const { address, isConnected, connect, disconnect } = useWallet()

  if (!isConnected) {
    return <button onClick={connect}>连接钱包</button>
  }

  return (
    <div>
      <span>{address?.slice(0, 6)}...{address?.slice(-4)}</span>
      <button onClick={disconnect}>断开</button>
    </div>
  )
}
```

---

## 第二层：多链配置 → Zustand

### 为什么用 Zustand 而不是 Redux Toolkit？

- **Redux Toolkit 模板代码多**：需要 `createSlice` + `configureStore` + `Provider` + `useSelector`
- **Zustand 三步搞定**：`create` → `useAppStore` → 直接用
- **天然支持 selector 防重渲染**：`useAppStore(s => s.gasPrice)` 只在 `gasPrice` 变化时重渲染
- **不需要 Provider 包裹**：Zustand 的 store 是模块级单例，直接在组件里 import 就行
- **Redux DevTools 支持**：Zustand 也支持，一行 `devtools()` 包裹即可

### 什么数据适合放 Zustand？

- **支持的链列表**（Ethereum、BSC、Polygon...）— 全局配置，基本不变
- **当前选中的链** — 用户切换链后全局生效
- **Gas 价格缓存** — 后端返回的实时 Gas，多个组件都要用
- **用户偏好** — 语言、货币单位（USD/CNY）、滑点默认值

### 代码示例

```typescript
// stores/useAppStore.ts
import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

/** 单条链的配置信息 */
interface ChainConfig {
  id: number
  name: string
  shortName: string
  rpcUrl: string
  explorerUrl: string
  nativeCurrency: {
    symbol: string
    decimals: number
  }
  /** 该链的常用代币合约地址 */
  tokens: Record<string, string>
}

interface AppState {
  // ===== 链配置 =====
  supportedChains: ChainConfig[]
  currentChainId: number
  setCurrentChain: (chainId: number) => void

  // ===== Gas 价格 =====
  /** chainId → { slow, normal, fast } 单位 Gwei */
  gasPrice: Record<number, { slow: string; normal: string; fast: string }>
  setGasPrice: (chainId: number, prices: { slow: string; normal: string; fast: string }) => void

  // ===== 用户偏好 =====
  slippage: number
  setSlippage: (slippage: number) => void

  preferredCurrency: 'USD' | 'CNY'
  setPreferredCurrency: (currency: 'USD' | 'CNY') => void
}

export const useAppStore = create<AppState>()(
  devtools(
    (set) => ({
      // 默认值
      supportedChains: [
        {
          id: 1,
          name: 'Ethereum',
          shortName: 'ETH',
          rpcUrl: 'https://eth.llamarpc.com',
          explorerUrl: 'https://etherscan.io',
          nativeCurrency: { symbol: 'ETH', decimals: 18 },
          tokens: {
            USDT: '0xdAC17F958D2ee523a2206206994597C13D831ec7',
            USDC: '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48',
          },
        },
        {
          id: 56,
          name: 'BNB Smart Chain',
          shortName: 'BSC',
          rpcUrl: 'https://bsc-dataseed.binance.org',
          explorerUrl: 'https://bscscan.com',
          nativeCurrency: { symbol: 'BNB', decimals: 18 },
          tokens: {
            USDT: '0x55d398326f99059fF775485246999027B3197955',
            USDC: '0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d',
          },
        },
        {
          id: 137,
          name: 'Polygon',
          shortName: 'MATIC',
          rpcUrl: 'https://polygon-rpc.com',
          explorerUrl: 'https://polygonscan.com',
          nativeCurrency: { symbol: 'MATIC', decimals: 18 },
          tokens: {
            USDT: '0xc2132D05D31c914a87C6611C10748AEb04B58e8F',
            USDC: '0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359',
          },
        },
      ],
      currentChainId: 1,
      setCurrentChain: (chainId) => set({ currentChainId: chainId }),

      // Gas 价格
      gasPrice: {},
      setGasPrice: (chainId, prices) =>
        set((state) => ({
          gasPrice: { ...state.gasPrice, [chainId]: prices },
        })),

      // 用户偏好
      slippage: 0.5, // 默认 0.5%
      setSlippage: (slippage) => set({ slippage }),

      preferredCurrency: 'USD',
      setPreferredCurrency: (currency) => set({ preferredCurrency: currency }),
    }),
    { name: 'app-store' }
  )
)
```

### 使用示例 — selector 精准订阅，避免无意义重渲染

```typescript
// components/GasIndicator.tsx
function GasIndicator() {
  // ✅ 只订阅 gasPrice 字段，currentChainId 变了也不会重渲染
  const gasPrice = useAppStore(s => s.gasPrice)
  const currentChainId = useAppStore(s => s.currentChainId)
  const chainGas = gasPrice[currentChainId]

  if (!chainGas) return <span>加载 Gas...</span>

  return (
    <div>
      <span>慢 🐢 {chainGas.slow} Gwei</span>
      <span>正常 ⚡ {chainGas.normal} Gwei</span>
      <span>快 🚀 {chainGas.fast} Gwei</span>
    </div>
  )
}

// components/SlippageSetting.tsx
function SlippageSetting() {
  // ✅ 只订阅 slippage，Gas 价格变化不会导致这里重渲染
  const slippage = useAppStore(s => s.slippage)
  const setSlippage = useAppStore(s => s.setSlippage)

  return (
    <div>
      <label>滑点容忍度: {slippage}%</label>
      <input
        type="range"
        min="0.1"
        max="5"
        step="0.1"
        value={slippage}
        onChange={e => setSlippage(Number(e.target.value))}
      />
    </div>
  )
}
```

### Zustand vs Redux Toolkit 实际对比

```typescript
// ========== Zustand 版本 ==========
const useStore = create<State>((set) => ({
  count: 0,
  increment: () => set(s => ({ count: s.count + 1 })),
}))

function Counter() {
  const count = useStore(s => s.count)       // 读取
  const increment = useStore(s => s.increment) // action
  return <button onClick={increment}>{count}</button>
}

// ========== Redux Toolkit 版本（需要更多文件）==========
// store.ts
const store = configureStore({ reducer: { counter: counterSlice.reducer } })

// counterSlice.ts
const counterSlice = createSlice({
  name: 'counter',
  initialState: { count: 0 },
  reducers: { increment: (s) => { s.count += 1 } },
})

// App.tsx — 还要包 Provider
<Provider store={store}><App /></Provider>

// Counter.tsx
function Counter() {
  const count = useSelector((s: RootState) => s.counter.count)
  const dispatch = useDispatch()
  return <button onClick={() => dispatch(counterSlice.actions.increment())}>{count}</button>
}
```

**结论**：对于全局配置类状态（不是服务端数据），Zustand 在代码量和心智负担上都优于 Redux Toolkit。

---

## 第三层：合约读写缓存 → RTK Query

### 为什么链上数据要单独管理？

链上数据有几个特点：
1. **数据源头在服务端**（RPC 节点/后端 API），不在客户端
2. **需要定时刷新**（余额 30 秒、Gas 10 秒、价格 15 秒）
3. **写操作后要自动刷新相关读缓存**（Swap 完自动刷新余额）
4. **需要处理加载态、错误态、缓存失效**

这些是 **服务端状态** 的典型特征，而 Zustand/Redux 本质上是 **客户端状态** 管理工具。RTK Query（底层是 TanStack Query）专门解决服务端状态的缓存、轮询、失效问题。

### 为什么不用 Zustand 管理链上数据？

```typescript
// ❌ 用 Zustand 手动管理链上数据的痛点
const useBalanceStore = create((set) => ({
  balance: null,
  loading: false,
  error: null,
  fetchBalance: async (address) => {
    set({ loading: true, error: null })
    try {
      const res = await fetch(`/api/balance/${address}`)
      set({ balance: await res.json(), loading: false })
    } catch (e) {
      set({ error: e, loading: false })
    }
  },
}))

// 问题：
// 1. 每次都要手写 loading/error 状态
// 2. 没有缓存 —— 两个组件请求同一个地址的余额，会发两次请求
// 3. 没有自动轮询 —— 需要手写 setInterval
// 4. 没有失效机制 —— Swap 后怎么通知所有余额查询"你过期了"？
// 5. 没有请求去重 —— 同时 3 个组件 mount，发 3 次相同请求
```

### RTK Query 代码示例

```typescript
// services/contractApi.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export const contractApi = createApi({
  reducerPath: 'contractApi',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),

  // tagTypes 用于声明缓存标签，写操作后按标签失效缓存
  tagTypes: ['Balance', 'GasPrice', 'Transaction', 'Allowance', 'PoolInfo'],

  endpoints: (builder) => ({

    // ========== 合约读（Query）==========

    // 1. 原生币余额 — 30 秒轮询
    getNativeBalance: builder.query<string, { address: string; chainId: number }>({
      query: ({ address, chainId }) => `/balance/${chainId}/native/${address}`,
      providesTags: (result, error, { address }) => [{ type: 'Balance', id: `${address}-native` }],
      // 链上数据不会秒级变化，30 秒足够
      pollingInterval: 30_000,
    }),

    // 2. 代币余额 — 30 秒轮询
    getTokenBalance: builder.query<string, { address: string; token: string; chainId: number }>({
      query: ({ address, token, chainId }) => `/balance/${chainId}/${token}/${address}`,
      providesTags: (result, error, { address, token }) => [{ type: 'Balance', id: `${address}-${token}` }],
      pollingInterval: 30_000,
    }),

    // 3. Gas 价格 — 10 秒轮询（价格波动快）
    getGasPrice: builder.query<{ slow: string; normal: string; fast: string }, number>({
      query: (chainId) => `/gas/${chainId}`,
      providesTags: (result, error, chainId) => [{ type: 'GasPrice', id: chainId }],
      pollingInterval: 10_000,
    }),

    // 4. 代币 USD 价格 — 15 秒轮询
    getTokenPrice: builder.query<number, { token: string; vsCurrency: string }>({
      query: ({ token, vsCurrency }) => `/price/${token}/${vsCurrency}`,
      pollingInterval: 15_000,
    }),

    // 5. 交易历史（分页）
    getTransactionHistory: builder.query<Transaction[], { address: string; chainId: number; page: number }>({
      query: ({ address, chainId, page }) => `/txs/${chainId}/${address}?page=${page}&limit=20`,
      providesTags: (result) =>
        result
          ? [...result.map(tx => ({ type: 'Transaction' as const, id: tx.hash })), { type: 'Transaction', id: 'LIST' }]
          : [{ type: 'Transaction', id: 'LIST' }],
      // 交易历史不需要轮询，用户手动刷新或用 invalidatesTags 在发送交易后刷新
    }),

    // 6. 代币授权额度（查 Approval）
    getAllowance: builder.query<string, { owner: string; spender: string; token: string; chainId: number }>({
      query: ({ owner, spender, token, chainId }) =>
        `/allowance/${chainId}/${token}/${owner}/${spender}`,
      providesTags: (result, error, { owner, token }) => [{ type: 'Allowance', id: `${owner}-${token}` }],
      // 授权额度平时不变，仅在用户 Approve 后需要刷新
    }),

    // 7. 流动池信息（LP 总锁仓量、APR）
    getPoolInfo: builder.query<PoolInfo, { poolAddress: string; chainId: number }>({
      query: ({ poolAddress, chainId }) => `/pool/${chainId}/${poolAddress}`,
      providesTags: (result, error, { poolAddress }) => [{ type: 'PoolInfo', id: poolAddress }],
      pollingInterval: 60_000, // 1 分钟刷新一次即可
    }),

    // 8. 多链资产汇总
    getMultiChainAssets: builder.query<AssetSummary, string>({
      query: (address) => `/assets/${address}`,
      providesTags: (result, error, address) => [{ type: 'Balance', id: `${address}-all` }],
      pollingInterval: 30_000,
    }),

    // ========== 合约写（Mutation）==========

    // 9. 发送交易
    sendTransaction: builder.mutation<
      { hash: string },
      { to: string; data: string; value: string; chainId: number }
    >({
      query: (tx) => ({ url: `/tx/${tx.chainId}/send`, method: 'POST', body: tx }),
      // 发送后失效交易历史，UI 自动重新拉取最新列表
      invalidatesTags: [{ type: 'Transaction', id: 'LIST' }],
    }),

    // 10. 代币授权
    approveToken: builder.mutation<
      { hash: string },
      { token: string; spender: string; amount: string; chainId: number }
    >({
      query: (body) => ({ url: '/approve', method: 'POST', body }),
      // 授权成功后，立刻刷新这条 allowance 的缓存
      invalidatesTags: (result, error, { token, spender }) => [
        { type: 'Allowance', id: `${spender}-${token}` },
      ],
    }),

    // 11. Swap 兑换
    swapTokens: builder.mutation<{ hash: string; amountOut: string }, SwapParams>({
      query: (body) => ({ url: '/swap', method: 'POST', body }),
      // Swap 完成后，余额变了、交易历史变了，全部自动刷新
      invalidatesTags: (result, error, { fromAddress }) => [
        { type: 'Balance', id: `${fromAddress}-native` },
        { type: 'Balance', id: `${fromAddress}-all` },
        { type: 'Transaction', id: 'LIST' },
      ],
    }),
  }),
})

// 导出自动生成的 Hooks
export const {
  useGetNativeBalanceQuery,
  useGetTokenBalanceQuery,
  useGetGasPriceQuery,
  useGetTokenPriceQuery,
  useGetTransactionHistoryQuery,
  useGetAllowanceQuery,
  useGetPoolInfoQuery,
  useGetMultiChainAssetsQuery,
  useSendTransactionMutation,
  useApproveTokenMutation,
  useSwapTokensMutation,
} = contractApi
```

### Store 配置 —— 把 RTK Query 的 reducer 注册到 Redux Store

```typescript
// store.ts
import { configureStore } from '@reduxjs/toolkit'
import { contractApi } from './services/contractApi'

export const store = configureStore({
  reducer: {
    // RTK Query 的 reducer 必须注册到这里
    [contractApi.reducerPath]: contractApi.reducer,
    // 如果你还有其他 slice，也在这里加
    // counter: counterSlice.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(contractApi.middleware),
})

// Provider 在 App 入口包裹
// <Provider store={store}><App /></Provider>
```

### 使用示例 —— 组件中消费 RTK Query 数据

```typescript
// pages/Dashboard.tsx — 仪表盘页面
function Dashboard() {
  const { address, isConnected, chainId } = useWallet() // Context
  const currentChain = useAppStore(s => {
    return s.supportedChains.find(c => c.id === chainId)
  }) // Zustand

  // RTK Query — 自动加载、缓存、轮询
  const {
    data: nativeBalance,
    isLoading: balanceLoading,
    error: balanceError,
  } = useGetNativeBalanceQuery(
    { address: address!, chainId: chainId! },
    { skip: !isConnected } // 没连钱包时跳过请求
  )

  const { data: gasPrice } = useGetGasPriceQuery(chainId!, { skip: !chainId })

  if (!isConnected) return <div>请先连接钱包</div>
  if (balanceLoading) return <div>加载余额...</div>
  if (balanceError) return <div>加载失败，请重试</div>

  return (
    <div>
      <h2>{currentChain?.name} 资产</h2>
      <p>余额: {nativeBalance} {currentChain?.nativeCurrency.symbol}</p>
      <p>Gas: {gasPrice?.normal} Gwei</p>
    </div>
  )
}

// pages/SwapPage.tsx — Swap 页面
function SwapPage() {
  const { address, isConnected, chainId } = useWallet()
  const slippage = useAppStore(s => s.slippage) // 从 Zustand 取用户滑点偏好

  const { data: tokenBalance } = useGetTokenBalanceQuery(
    { address: address!, token: '0x...', chainId: chainId! },
    { skip: !isConnected }
  )

  const [swapTokens, { isLoading: isSwapping }] = useSwapTokensMutation()

  const handleSwap = async () => {
    const result = await swapTokens({
      fromAddress: address!,
      tokenIn: '0x...',
      tokenOut: '0x...',
      amountIn: '1.0',
      slippage,
      chainId: chainId!,
    }).unwrap()

    console.log('Swap 成功, txHash:', result.hash)
    // 余额会自动刷新！因为 swapTokens 的 invalidatesTags 失效了 Balance 缓存
  }

  return (
    <div>
      <p>余额: {tokenBalance}</p>
      <p>滑点: {slippage}%</p>
      <button onClick={handleSwap} disabled={isSwapping}>
        {isSwapping ? '交易中...' : 'Swap'}
      </button>
    </div>
  )
}
```

### RTK Query 的核心机制图解

```
┌─────────────────────────────────────────────────────────┐
│                    RTK Query 工作流程                      │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  组件 mount（如 Dashboard）                               │
│       │                                                 │
│       ▼                                                 │
│  useGetNativeBalanceQuery({ address, chainId })         │
│       │                                                 │
│       ▼                                                 │
│  ┌─────────────┐    未命中     ┌─────────────┐          │
│  │  检查缓存    │ ──────────→  │  发起请求    │          │
│  └─────────────┘              └──────┬──────┘          │
│       │ 命中                         │                  │
│       ▼                              ▼                  │
│  直接返回缓存数据              更新 Store + 通知组件      │
│                                     │                  │
│                                     ▼                  │
│                            ┌──────────────┐            │
│                            │ 定时轮询      │            │
│                            │ (30 秒后自动   │            │
│                            │  重新请求)    │            │
│                            └──────────────┘            │
│                                                         │
│  Mutation 触发（如 Swap）                                 │
│       │                                                 │
│       ▼                                                 │
│  swapTokens() → invalidatesTags: ['Balance']            │
│       │                                                 │
│       ▼                                                 │
│  所有 providesTags 为 Balance 的 Query 自动重新请求       │
│       │                                                 │
│       ▼                                                 │
│  组件拿到最新余额（无需手动 refetch）                       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 第四层：交易临时状态 → useState / useReducer

### 为什么不用全局 Store？

- **生命周期短**：弹窗关了状态就没用了，留在全局 Store 是垃圾
- **组件隔离**：A 组件的交易弹窗状态不应该影响 B 组件
- **简单直接**：一个 `useState(false)` 就能解决的，不值得写 reducer
- **避免污染**：全局 Store 里的 `isModalOpen` 多了以后，根本分不清是哪个弹窗

### 什么时候用 useState，什么时候用 useReducer？

```typescript
// ✅ useState — 简单状态（1-3 个独立状态）
function TransactionModal() {
  const [isOpen, setIsOpen] = useState(false)
  const [amount, setAmount] = useState('')
  const [slippage, setSlippage] = useState(0.5)

  // 各自独立 set，逻辑简单
}

// ✅ useReducer — 多状态联动（4 个以上，或状态之间有依赖关系）
type TxState = 'idle' | 'pending' | 'confirming' | 'confirmed' | 'failed'

interface TxStateObj {
  status: TxState
  hash: string | null
  error: string | null
  confirmations: number
}

type TxAction =
  | { type: 'SEND' }
  | { type: 'PENDING'; hash: string }
  | { type: 'CONFIRMING'; confirmations: number }
  | { type: 'CONFIRMED' }
  | { type: 'FAILED'; error: string }
  | { type: 'RESET' }

const initialState: TxStateObj = {
  status: 'idle',
  hash: null,
  error: null,
  confirmations: 0,
}

function txReducer(state: TxStateObj, action: TxAction): TxStateObj {
  switch (action.type) {
    case 'SEND':
      return { ...state, status: 'pending', error: null }
    case 'PENDING':
      return { ...state, status: 'pending', hash: action.hash }
    case 'CONFIRMING':
      return { ...state, status: 'confirming', confirmations: action.confirmations }
    case 'CONFIRMED':
      return { ...state, status: 'confirmed' }
    case 'FAILED':
      return { ...state, status: 'failed', error: action.error }
    case 'RESET':
      return initialState
    default:
      return state
  }
}

function SwapButton() {
  const [state, dispatch] = useReducer(txReducer, initialState)

  const handleSwap = async () => {
    dispatch({ type: 'SEND' })
    try {
      const hash = await walletClient.sendTransaction(tx)
      dispatch({ type: 'PENDING', hash })
      // 等待确认...
      const receipt = await publicClient.waitForTransactionReceipt({ hash })
      dispatch({ type: 'CONFIRMED' })
    } catch (e: any) {
      dispatch({ type: 'FAILED', error: e.message })
    }
  }

  return (
    <div>
      <button onClick={handleSwap} disabled={state.status === 'pending'}>
        {state.status === 'pending' ? '交易确认中...' : 'Swap'}
      </button>
      {state.status === 'failed' && <p style={{ color: 'red' }}>{state.error}</p>}
      {state.status === 'confirmed' && <p style={{ color: 'green' }}>交易成功! {state.hash}</p>}
    </div>
  )
}
```

### 交易状态流转图

```
idle ──→ pending ──→ confirming ──→ confirmed
  │                    │
  └────────────────────┴──→ failed
                              │
                              ▼
                           reset ──→ idle
```

---

## 完整架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        App 根组件                            │
│                                                             │
│  <Provider store={store}>        ← Redux Store（给 RTK Query）│
│    <WalletProvider>              ← Context（钱包连接）        │
│      <Router>                                              │
│        <Routes>                                            │
│          <Route path="/swap"    element={<SwapPage />} />  │
│          <Route path="/pool"    element={<PoolPage />} />  │
│          <Route path="/assets"  element={<AssetsPage />} />│
│        </Routes>                                           │
│      </Router>                                             │
│    </WalletProvider>                                       │
│  </Provider>                                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      数据流示意                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  钱包连接（Context）                                         │
│  ├─ address, chainId, isConnected                          │
│  ├─ 提供：useWallet() Hook                                 │
│  └─ 消费者：导航栏、Swap 页、资产页、所有需要钱包信息的组件     │
│                                                             │
│  多链配置（Zustand）                                         │
│  ├─ supportedChains, gasPrice, slippage, currency          │
│  ├─ 提供：useAppStore(selector) Hook                       │
│  └─ 消费者：Gas 指示器、链切换器、滑点设置                    │
│                                                             │
│  合约读写（RTK Query）                                       │
│  ├─ balance, gasPrice, price, txs, allowance, poolInfo     │
│  ├─ 提供：useGetXxxQuery() / useXxxMutation() Hooks        │
│  ├─ 特性：30s 轮询、缓存去重、写后自动失效                    │
│  └─ 消费者：Dashboard、Swap 页、交易历史、资产页              │
│                                                             │
│  交易临时态（useState / useReducer）                         │
│  ├─ isModalOpen, inputValue, txStatus, txHash, txError     │
│  ├─ 提供：组件内部的 state + dispatch                       │
│  └─ 消费者：仅限于当前组件（弹窗、输入框、交易进度条）          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 各层对比速查表

| 维度 | Context | Zustand | RTK Query | useState/useReducer |
|------|---------|---------|-----------|-------------------|
| **管理对象** | 钱包连接 | 全局配置 | 链上数据 | UI 临时状态 |
| **数据归属** | 客户端状态 | 客户端状态 | **服务端状态** | 客户端状态 |
| **生命周期** | 应用级 | 应用级 | 由缓存策略决定 | 组件级 |
| **变化频率** | 极低（会话级） | 低（手动切换） | 中高（定时轮询） | 高（用户交互） |
| **重渲染范围** | 所有消费者 | selector 精准 | 组件级 | 仅当前组件 |
| **缓存机制** | 无 | 内存持久 | **自动缓存 + 失效** | 无 |
| **轮询支持** | ❌ | ❌ | ✅ pollingInterval | ❌ |
| **写后刷新** | ❌ | ❌ | ✅ invalidatesTags | ❌ |
| **请求去重** | ❌ | ❌ | ✅ 内置 | ❌ |
| **是否需要 Provider** | ✅ | ❌ | ✅（Redux Provider） | ❌ |
| **DevTools 支持** | ❌ | ✅ | ✅ | ❌ |
| **典型数据** | address, chainId | 链列表, 滑点 | 余额, Gas, 价格 | 弹窗开关, 输入值 |

---

## 面试口述话术

> "我在大型 Web3 项目中使用四层分层策略管理状态，核心原则是**服务端状态和客户端状态分离管理**。
>
> **第一层**，钱包连接状态用 React Context。因为数据量小——就 address、chainId、isConnected 几个字段，而且极少变化，Context 轻量且不需要额外依赖。
>
> **第二层**，多链配置和用户偏好用 Zustand。比 Redux Toolkit 模板代码少很多，天然支持 selector 防重渲染——Gas 价格变化不会导致滑点设置组件也跟着渲染。
>
> **第三层**，合约读写的链上数据用 RTK Query。这是关键——余额、Gas、代币价格这些数据源头在服务端，需要缓存、轮询、写后自动刷新，这些正是 RTK Query 擅长的。比如用户做完一笔 Swap，RTK Query 的 invalidatesTags 会自动失效余额和交易历史缓存，UI 无需手动 refetch 就能拿到最新数据。
>
> **第四层**，交易弹窗、输入框这类组件内部临时状态直接用 useState 或 useReducer，以 txStatus 为例——从 idle 到 pending 到 confirmed 的状态流转用 useReducer 更清晰，组件卸载状态就销毁，不会污染全局 Store。
>
> 这套分层模式的来源是 CoderWhy 的 React 电商课程——他把商品列表、购物车、用户信息分层管理，我把这个思想迁移到了 Web3 场景。"
