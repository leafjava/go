# useEffect 依赖数组原理 + Web3 自定义 Hook 实战

> 涵盖 useEffect 工作机制、依赖数组陷阱、以及 Web3 项目中三个核心自定义 Hook 的完整实现。

---

## 目录

- [useEffect 依赖数组原理](#useeffect-依赖数组原理)
  - [三种形态](#三种形态)
  - [浅比较机制](#浅比较机制)
  - [常见陷阱](#常见陷阱)
  - [Vue 对比](#vue-对比)
- [自定义 Hook 设计思想](#自定义-hook-设计思想)
- [实战一：useWallet](#实战一usewallet)
- [实战二：useContractRead](#实战二usecontractread)
- [实战三：useTransaction](#实战三usetransaction)
- [Hook 组合使用示例](#hook-组合使用示例)
- [面试口述话术](#面试口述话术)

---

## useEffect 依赖数组原理

### 三种形态

```typescript
import { useEffect, useState } from 'react'

function UserProfile({ userId }: { userId: number }) {

  // ===== 形态一：无第二个参数 =====
  // 每次渲染后都执行（极少使用，容易造成死循环）
  useEffect(() => {
    console.log('每次渲染都执行，慎用！')
  })

  // ===== 形态二：空数组 [] =====
  // 只在组件挂载时执行一次，相当于 Vue 的 mounted
  // cleanup 函数在组件卸载时执行，相当于 Vue 的 unmounted
  useEffect(() => {
    console.log('组件挂载，只执行一次')

    // 订阅全局事件
    window.addEventListener('resize', handleResize)

    return () => {
      // cleanup：组件卸载时清理
      console.log('组件卸载，清理副作用')
      window.removeEventListener('resize', handleResize)
    }
  }, [])

  // ===== 形态三：有依赖 [userId] =====
  // 首次渲染执行 + 每次 userId 变化时重新执行
  useEffect(() => {
    console.log('userId 变了，重新获取用户数据:', userId)
    fetchUserData(userId)
  }, [userId])

  return <div>...</div>
}
```

### 三种形态速查

| 依赖数组 | 执行时机 | Vue 类比 | 使用场景 |
|---------|---------|---------|---------|
| 不传 | 每次渲染后 | `updated`（无条件） | 极少使用 |
| `[]` | 挂载一次 | `mounted` | 初始化监听、订阅 |
| `[a, b]` | 挂载 + a/b 变化时 | `watch([a, b])` | 响应数据变化 |
| `return () => {}` | 卸载 / 重新执行前 | `unmounted` / `watch` 的 flush | 清理定时器、取消订阅 |

### 浅比较机制

useEffect 判断"依赖是否变了"用的是 **`Object.is` 浅比较**，不是深比较。

```typescript
// Object.is 的行为
Object.is(1, 1)           // true  → 不重新执行
Object.is('hello', 'hello') // true  → 不重新执行
Object.is(NaN, NaN)       // true  → 不重新执行（与 === 不同！）
Object.is(+0, -0)         // false → 会重新执行（与 === 不同！）

// ⚠️ 引用类型的陷阱
Object.is({ name: 'leaf' }, { name: 'leaf' })  // false → 会重新执行！
Object.is([1, 2, 3], [1, 2, 3])               // false → 会重新执行！
```

**这意味着**：如果你把对象、数组、函数直接写在依赖数组里，**每次渲染都会触发 effect**。因为每次渲染时创建的都是新引用。

```typescript
function BadComponent() {
  const [count, setCount] = useState(0)

  // ❌ 错误：config 每次渲染都是新对象 → effect 每次渲染都执行
  const config = { timeout: 5000, retries: 3 }

  useEffect(() => {
    fetchData(config)
  }, [config]) // 永远不"相等"，每次都跑

  // ❌ 错误：fetchOptions 每次渲染都是新函数 → 同理
  const fetchOptions = () => ({ headers: { Authorization: 'Bearer xxx' } })

  useEffect(() => {
    fetchData(fetchOptions())
  }, [fetchOptions]) // 永远不"相等"

  return <div>{count}</div>
}
```

**修复方案**：

```typescript
function GoodComponent() {
  // ✅ 方案一：把原始值拆开写入依赖
  const timeout = 5000
  const retries = 3

  useEffect(() => {
    fetchData({ timeout, retries })
  }, [timeout, retries]) // 原始值，值相同就不会重新执行

  // ✅ 方案二：用 useMemo 固定引用（当对象确实需要传引用时）
  const config = useMemo(() => ({ timeout: 5000, retries: 3 }), [])

  useEffect(() => {
    fetchData(config)
  }, [config]) // config 引用不变，不会重复执行

  // ✅ 方案三：用 useCallback 固定函数引用
  const fetchOptions = useCallback(() => {
    return { headers: { Authorization: 'Bearer xxx' } }
  }, [])

  useEffect(() => {
    fetchData(fetchOptions())
  }, [fetchOptions]) // fetchOptions 引用不变
}
```

### 常见陷阱

#### 陷阱一：不完全的依赖数组（React 官方推荐用 lint 规则 `react-hooks/exhaustive-deps`）

```typescript
// ❌ 危险：count 在 effect 里用到了，但没写进依赖
function Counter() {
  const [count, setCount] = useState(0)

  useEffect(() => {
    const timer = setInterval(() => {
      console.log(count) // 永远打印 0！因为闭包捕获了初始值
    }, 1000)
    return () => clearInterval(timer)
  }, []) // 缺少 count

  return <button onClick={() => setCount(c => c + 1)}>+1</button>
}

// ✅ 修复：用函数式更新避免依赖外部变量
function Counter() {
  const [count, setCount] = useState(0)

  useEffect(() => {
    const timer = setInterval(() => {
      setCount(c => {
        console.log(c) // 每次拿到的都是最新值
        return c + 1
      })
    }, 1000)
    return () => clearInterval(timer)
  }, []) // 不需要 count 作为依赖
}
```

#### 陷阱二：在 effect 里做不必要的 setState，导致死循环

```typescript
// ❌ 死循环：fetch → setState → 重新渲染 → fetch → ...
function BadFetch() {
  const [data, setData] = useState(null)

  useEffect(() => {
    fetch('/api/data')
      .then(r => r.json())
      .then(setData)  // setData 触发重渲染 → effect 再次执行
  }) // 没有依赖数组 → 每次渲染执行 → 死循环！

  return <div>{data}</div>
}

// ✅ 空数组，只执行一次
function GoodFetch() {
  const [data, setData] = useState(null)

  useEffect(() => {
    fetch('/api/data')
      .then(r => r.json())
      .then(setData)
  }, []) // 只挂载时执行一次
}
```

#### 陷阱三：忘记 cleanup，导致内存泄漏

```typescript
// ❌ 组件卸载后仍然尝试 setState（内存泄漏 + 警告）
function BadSubscriber() {
  const [data, setData] = useState(null)

  useEffect(() => {
    const ws = new WebSocket('wss://...')
    ws.onmessage = (e) => setData(JSON.parse(e.data))
    // 没有 cleanup → 组件卸载后 ws 仍然连接，setData 报 warning
  }, [])

  return <div>{data}</div>
}

// ✅ cleanup 中关闭 WebSocket 并标记已卸载
function GoodSubscriber() {
  const [data, setData] = useState(null)

  useEffect(() => {
    let cancelled = false           // 卸载标记
    const ws = new WebSocket('wss://...')

    ws.onmessage = (e) => {
      if (!cancelled) {            // 已卸载就不再 setState
        setData(JSON.parse(e.data))
      }
    }

    return () => {
      cancelled = true             // 标记已卸载
      ws.close()                   // 关闭连接
    }
  }, [])

  return <div>{data}</div>
}
```

### Vue 对比

如果你有 Vue 背景，这样类比 React useEffect：

| React | Vue (Composition API) |
|-------|----------------------|
| `useEffect(fn, [])` | `onMounted(fn)` |
| `useEffect(() => { return cleanup }, [])` | `onMounted(fn); onUnmounted(cleanup)` |
| `useEffect(fn, [a, b])` | `watch([a, b], fn, { immediate: true })` |
| `useEffect(() => { return cleanup }, [a])` | `watch(a, (nv, ov) => { cleanup(); fn() })` |
| `useMemo` | `computed` |
| `useCallback` | 不需要（Vue 的响应式系统自动处理函数引用） |

```typescript
// ===== React =====
useEffect(() => {
  fetchData(userId)
}, [userId])

// ===== Vue 3 =====
watch(() => props.userId, (newUserId) => {
  fetchData(newUserId)
}, { immediate: true }) // immediate 让它在首次也执行
```

---

## 自定义 Hook 设计思想

### 为什么要抽自定义 Hook？

React 官方定义：**自定义 Hook 是复用状态逻辑的机制，不是复用状态本身**。

```
问题场景：
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  SwapPage     │    │  PoolPage     │    │  AssetsPage   │
│              │    │              │    │              │
│  const [tx,  │    │  const [tx,  │    │  const [tx,  │
│   setTx] =   │    │   setTx] =   │    │   setTx] =   │
│   useState() │    │   useState() │    │   useState() │
│              │    │              │    │              │
│  一堆重复的   │    │  一堆重复的   │    │  一堆重复的   │
│  交易处理逻辑  │    │  交易处理逻辑  │    │  交易处理逻辑  │
│  (50 行)     │    │  (50 行)     │    │  (50 行)     │
└──────────────┘    └──────────────┘    └──────────────┘

                    ▼ 抽出

┌─────────────────────────────────────────────────────────┐
│  useTransaction()                                       │
│  ├─ status: idle → pending → confirming → confirmed     │
│  ├─ hash, error, send, reset                           │
│  └─ 可复用：Swap 页、Pool 页、跨链桥、任意发交易的组件     │
└─────────────────────────────────────────────────────────┘
```

### 设计原则

1. **一个 Hook 只做一件事**：`useWallet` 只管连接/断开，`useTransaction` 只管交易状态
2. **返回值语义化**：返回 `{ data, loading, error, refetch }` 这种约定俗成的结构
3. **可组合**：`useContractRead` 内部调 `useEffect`，`useTransaction` 内部调 `useContractRead`
4. **稳定的函数引用**：所有返回的函数用 `useCallback` 包裹，避免子组件无意义重渲染

---

## 实战一：useWallet

### 功能说明

封装 Wagmi 的钱包连接、断开、切换链逻辑，对外暴露简单 API。底层用 Wagmi 的 `useAccount` / `useConnect` 等 Hook，上层组件不需要关心这些细节。

### 完整代码

```typescript
// hooks/useWallet.ts
import { useCallback, useMemo } from 'react'
import {
  useAccount,
  useConnect,
  useDisconnect,
  useChainId,
  useSwitchChain,
  useBalance,
} from 'wagmi'

/** 支持的连接器类型 */
type ConnectorType = 'metaMask' | 'walletConnect' | 'coinbaseWallet'

interface UseWalletReturn {
  // ===== 状态 =====
  /** 当前钱包地址，未连接时为 undefined */
  address: `0x${string}` | undefined
  /** 地址的短格式显示（0x1234...abcd） */
  shortAddress: string | undefined
  /** 是否已连接 */
  isConnected: boolean
  /** 当前链 ID */
  chainId: number | undefined
  /** 原生币余额 */
  balance: string | undefined
  /** 连接中 */
  isConnecting: boolean

  // ===== 操作 =====
  /** 连接指定类型钱包 */
  connect: (connectorType: ConnectorType) => Promise<void>
  /** 断开连接 */
  disconnect: () => void
  /** 切换到目标链 */
  switchChain: (chainId: number) => void

  // ===== 衍生 =====
  /** 是否在主网 */
  isMainnet: boolean
  /** 当前链的浏览器地址 */
  explorerUrl: string | undefined
}

/**
 * 钱包连接 Hook
 *
 * 封装 Wagmi 的底层 Hook，提供统一的钱包操作 API。
 * 所有组件通过这个 Hook 操作钱包，而不是直接依赖 Wagmi。
 *
 * 好处：
 * 1. 换钱包库时（如从 Wagmi 换成 RainbowKit），只改这一个文件
 * 2. 统一错误处理和日志
 * 3. 添加业务逻辑（如连接成功后自动切换链、埋点上报）
 */
export function useWallet(): UseWalletReturn {
  const { address, isConnected } = useAccount()
  const { connectors, connectAsync } = useConnect()
  const { disconnect } = useDisconnect()
  const chainId = useChainId()
  const { switchChain: wagmiSwitchChain } = useSwitchChain()

  // 查询原生币余额
  const { data: balance } = useBalance({ address })

  // ✅ 用 useCallback 稳定引用
  const connect = useCallback(
    async (connectorType: ConnectorType) => {
      const connector = connectors.find(c => c.id === connectorType)
      if (!connector) {
        console.error(`连接器 ${connectorType} 不可用`)
        return
      }
      try {
        await connectAsync({ connector })
        console.log('钱包连接成功:', connectorType)
      } catch (error) {
        console.error('钱包连接失败:', error)
        throw error
      }
    },
    [connectors, connectAsync]
  )

  const switchChain = useCallback(
    (targetChainId: number) => {
      try {
        wagmiSwitchChain({ chainId: targetChainId })
      } catch (error) {
        console.error('切换链失败:', error)
      }
    },
    [wagmiSwitchChain]
  )

  // ✅ 用 useMemo 缓存派生值
  const shortAddress = useMemo(
    () => (address ? `${address.slice(0, 6)}...${address.slice(-4)}` : undefined),
    [address]
  )

  // 主网链 ID 列表（可根据项目扩展）
  const MAINNET_IDS = [1, 56, 137, 42161]

  const isMainnet = useMemo(
    () => (chainId ? MAINNET_IDS.includes(chainId) : false),
    [chainId]
  )

  // 根据 chainId 生成浏览器地址
  const explorers: Record<number, string> = {
    1: 'https://etherscan.io',
    56: 'https://bscscan.com',
    137: 'https://polygonscan.com',
    42161: 'https://arbiscan.io',
  }

  const explorerUrl = useMemo(
    () => (chainId ? `${explorers[chainId]}/address/${address}` : undefined),
    [chainId, address]
  )

  return {
    address,
    shortAddress,
    isConnected,
    chainId,
    balance: balance?.formatted,
    isConnecting: false, // Wagmi 内部处理，这里简化
    connect,
    disconnect,
    switchChain,
    isMainnet,
    explorerUrl,
  }
}
```

### 使用示例

```typescript
// components/WalletPanel.tsx
import { useWallet } from '../hooks/useWallet'
import { useAppStore } from '../stores/useAppStore'

function WalletPanel() {
  const {
    address,
    shortAddress,
    isConnected,
    chainId,
    balance,
    connect,
    disconnect,
    explorerUrl,
  } = useWallet()

  // 从 Zustand 拿当前链的详细信息
  const chainInfo = useAppStore(s =>
    s.supportedChains.find(c => c.id === chainId)
  )

  if (!isConnected) {
    return (
      <div className="wallet-panel">
        <button onClick={() => connect('metaMask')}>
          <img src="/metamask.svg" alt="MetaMask" />
          MetaMask
        </button>
        <button onClick={() => connect('walletConnect')}>
          <img src="/walletconnect.svg" alt="WalletConnect" />
          WalletConnect
        </button>
      </div>
    )
  }

  return (
    <div className="wallet-panel">
      <span>{shortAddress}</span>
      <span>
        {balance} {chainInfo?.nativeCurrency.symbol}
      </span>
      <a href={explorerUrl} target="_blank" rel="noopener noreferrer">
        浏览器查看 ↗
      </a>
      <button onClick={disconnect}>断开</button>
    </div>
  )
}
```

---

## 实战二：useContractRead

### 功能说明

封装合约读取的加载态、错误态、数据、重试逻辑。组件只需要关心"我要读哪个合约的哪个方法"，不需要手动管理 loading 和 try-catch。

### 设计思路

```
┌─────────────────────────────────────────────┐
│              useContractRead                 │
│                                             │
│  输入:                                       │
│    address   ← 合约地址                       │
│    abi       ← 合约 ABI                      │
│    fnName    ← 调用的函数名                    │
│    args      ← 函数参数（可选）                 │
│    enabled   ← 是否启用（可选，默认 true）       │
│    watch     ← 是否监听变化自动重取（默认 false） │
│                                             │
│  输出:                                       │
│    data      ← 合约返回的数据                  │
│    loading   ← 是否加载中                      │
│    error     ← 错误信息（含用户友好的提示）       │
│    refetch   ← 手动重新读取                     │
│                                             │
│  内置处理:                                    │
│    ✅ 自动 loading / error 状态管理            │
│    ✅ 合约调用失败 → 解析 revert reason         │
│    ✅ 组件卸载 → 取消未完成请求                  │
│    ✅ 参数变化 → 自动重新读取                    │
└─────────────────────────────────────────────┘
```

### 完整代码

```typescript
// hooks/useContractRead.ts
import { useState, useEffect, useCallback, useRef } from 'react'
import { usePublicClient } from 'wagmi'
import type { Abi } from 'viem'

interface UseContractReadOptions {
  /** 是否启用自动读取，默认 true */
  enabled?: boolean
  /** 是否监听参数变化自动重取，默认 true */
  watch?: boolean
  /** 错误重试次数，默认 0（不重试） */
  retryCount?: number
  /** 重试间隔（ms），默认 1000 */
  retryDelay?: number
}

interface UseContractReadReturn<T> {
  /** 合约返回的数据 */
  data: T | null
  /** 是否加载中 */
  loading: boolean
  /** 错误信息 */
  error: string | null
  /** 手动重新读取 */
  refetch: () => Promise<void>
}

/**
 * 合约读取 Hook
 *
 * 封装 Viem 的 publicClient.readContract，自动处理加载态、错误态、
 * 参数变化重取、组件卸载取消等通用逻辑。
 *
 * 示例：
 * const { data: balance, loading, error } = useContractRead(
 *   tokenAddress,
 *   erc20Abi,
 *   'balanceOf',
 *   [userAddress],
 * )
 */
export function useContractRead<T = unknown>(
  contractAddress: `0x${string}` | undefined,
  abi: Abi,
  functionName: string,
  args: unknown[] = [],
  options: UseContractReadOptions = {}
): UseContractReadReturn<T> {
  const {
    enabled = true,
    watch = true,
    retryCount = 0,
    retryDelay = 1000,
  } = options

  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 用于 cleanup：组件卸载后不 setState
  const mountedRef = useRef(true)

  // Wagmi 的 publicClient
  const publicClient = usePublicClient()

  const read = useCallback(async () => {
    if (!contractAddress || !publicClient) return

    setLoading(true)
    setError(null)

    let lastError: Error | null = null

    // 重试逻辑
    for (let attempt = 0; attempt <= retryCount; attempt++) {
      // 不是第一次，等待重试间隔
      if (attempt > 0) {
        await new Promise(r => setTimeout(r, retryDelay))
      }

      // 检查组件是否还在挂载中
      if (!mountedRef.current) return

      try {
        const result = await publicClient.readContract({
          address: contractAddress,
          abi,
          functionName,
          args,
        })

        // 成功，setState 前再次检查挂载状态
        if (mountedRef.current) {
          setData(result as T)
          setLoading(false)
          setError(null)
        }
        return
      } catch (e: any) {
        lastError = e
        console.warn(
          `合约读取失败 (${attempt + 1}/${retryCount + 1}):`,
          functionName,
          e.shortMessage || e.message
        )
      }
    }

    // 所有重试都失败了
    if (mountedRef.current) {
      // 尝试解析 revert reason
      const friendlyError = parseContractError(lastError, functionName)
      setError(friendlyError)
      setLoading(false)
    }
  }, [contractAddress, abi, functionName, JSON.stringify(args), publicClient, retryCount, retryDelay])

  // 首次执行 + watch 为 true 时参数变化自动重取
  useEffect(() => {
    if (!enabled) return

    mountedRef.current = true
    read()

    return () => {
      mountedRef.current = false
    }
  }, [enabled, watch ? read : undefined])

  // 如果 watch 为 false，只在 enabled 变化和手动 refetch 时执行
  // 上面的 effect 已覆盖 watch=true 的情况
  useEffect(() => {
    if (!enabled || watch) return

    mountedRef.current = true
    read()

    return () => {
      mountedRef.current = false
    }
  }, [enabled])

  return { data, loading, error, refetch: read }
}

/**
 * 把合约调用错误解析成用户友好的提示
 */
function parseContractError(error: Error | null, functionName: string): string {
  if (!error) return '未知错误'

  const msg = 'shortMessage' in error
    ? (error as any).shortMessage
    : error.message

  // 根据错误信息返回中文提示
  if (msg?.includes('insufficient')) return '余额不足'
  if (msg?.includes('revert')) return `合约调用失败: ${functionName}`
  if (msg?.includes('timeout')) return '网络超时，请重试'
  if (msg?.includes('user rejected')) return '用户取消了操作'

  return msg || '合约读取失败'
}
```

### 使用示例

```typescript
// pages/TokenInfo.tsx — 查询代币信息
import { useContractRead } from '../hooks/useContractRead'
import { erc20Abi } from '../abi/erc20'
import { useWallet } from '../hooks/useWallet'

function TokenInfo({ tokenAddress }: { tokenAddress: `0x${string}` }) {
  const { address } = useWallet()

  // 查代币名称
  const { data: name, loading: nameLoading } = useContractRead<string>(
    tokenAddress,
    erc20Abi,
    'name',
  )

  // 查代币符号
  const { data: symbol } = useContractRead<string>(
    tokenAddress,
    erc20Abi,
    'symbol',
  )

  // 查当前用户的代币余额 — 依赖 userAddress，地址变自动重取
  const { data: balance, loading: balanceLoading, error: balanceError } = useContractRead<bigint>(
    tokenAddress,
    erc20Abi,
    'balanceOf',
    [address!],
    { enabled: !!address, retryCount: 2 },
  )

  // 查精度 (decimals 不会变，不 watch)
  const { data: decimals } = useContractRead<number>(
    tokenAddress,
    erc20Abi,
    'decimals',
    [],
    { watch: false },
  )

  if (nameLoading) return <div>加载代币信息...</div>

  return (
    <div>
      <h3>{name} ({symbol})</h3>
      {balanceLoading ? (
        <span>加载余额...</span>
      ) : balanceError ? (
        <span style={{ color: 'red' }}>⚠ {balanceError}</span>
      ) : (
        <span>
          余额: {formatUnits(balance ?? 0n, decimals ?? 18)} {symbol}
        </span>
      )}
    </div>
  )
}
```

### useContractRead vs 直接用 RTK Query

| 场景 | 用什么 | 原因 |
|------|--------|------|
| 前端直接调合约（读） | `useContractRead` | 调用 Viem 的 publicClient，不经过后端 |
| 通过后端 API 读链上数据 | RTK Query | 后端聚合/缓存，需要轮询 + 失效策略 |
| 两者混合 | 内部 API 用 RTK Query，前端直调用 `useContractRead` | 简单查询直调合约更快 |

---

## 实战三：useTransaction

### 功能说明

管理一笔链上交易从发送到确认的完整生命周期。这是 Web3 项目中最常用的 Hook。

### 状态转移图

```
                    ┌──────────────────────┐
                    │       idle           │ ← reset()
                    │  (初始/已完成/已重置)   │
                    └──────┬───────────────┘
                           │ send(tx)
                           ▼
                    ┌──────────────────────┐
                    │      pending         │
                    │  (已提交到 mempool)    │
                    └──────┬───────────────┘
                           │ 拿到 txHash
                           ▼
                    ┌──────────────────────┐
                    │     confirming       │──→ 等待 confirmations 达标
                    │  (等待区块确认)        │
                    └──────┬───────────────┘
                           │
                  ┌────────┴────────┐
                  ▼                 ▼
           ┌───────────┐    ┌───────────┐
           │ confirmed │    │  failed   │
           │  ✅ 成功   │    │  ❌ 失败   │
           └───────────┘    └───────────┘
```

### 完整代码

```typescript
// hooks/useTransaction.ts
import { useState, useCallback, useRef } from 'react'
import { usePublicClient, useWalletClient } from 'wagmi'
import type { TransactionReceipt } from 'viem'

// ===== 类型定义 =====

/** 交易状态枚举 */
export type TxStatus =
  | 'idle'        // 未开始
  | 'pending'     // 已提交到 mempool，等待被打包
  | 'confirming'  // 已上链，等待足够确认数
  | 'confirmed'   // 达成确认数，成功
  | 'failed'      // 失败（revert / 用户取消 / 超时）

interface TransactionState {
  status: TxStatus
  /** 交易 hash，pending 之后才有 */
  hash: `0x${string}` | null
  /** 错误信息，failed 时才有 */
  error: string | null
  /** 当前确认块数 */
  confirmations: number
  /** 交易回执，confirmed 时才有 */
  receipt: TransactionReceipt | null
}

interface SendTransactionOptions {
  /** 期望确认数，默认 1 */
  confirmations?: number
  /** 超时时间（ms），默认 60000（1 分钟） */
  timeout?: number
  /** 交易成功回调 */
  onSuccess?: (receipt: TransactionReceipt) => void
  /** 交易失败回调 */
  onError?: (error: Error) => void
  /** Gas 限制（可选，默认自动估算） */
  gas?: bigint
}

interface UseTransactionReturn extends TransactionState {
  /** 发送交易 */
  send: (tx: any, options?: SendTransactionOptions) => Promise<void>
  /** 重置为 idle 状态 */
  reset: () => void
  /** 是否在交易中（pending 或 confirming） */
  isProcessing: boolean
  /** 用户友好的状态中文描述 */
  statusText: string
}

// ===== 初始状态 =====

const initialState: TransactionState = {
  status: 'idle',
  hash: null,
  error: null,
  confirmations: 0,
  receipt: null,
}

// ===== 状态中文映射 =====

const statusTextMap: Record<TxStatus, string> = {
  idle: '待发送',
  pending: '交易待确认',
  confirming: '区块确认中',
  confirmed: '交易成功',
  failed: '交易失败',
}

/**
 * 交易管理 Hook
 *
 * 管理一笔链上交易的完整生命周期：
 * 发送 → 等待上链 → 等待确认 → 成功/失败
 *
 * 示例：
 * const { send, status, hash, error, isProcessing } = useTransaction()
 *
 * await send(
 *   { to: '0x...', value: parseEther('0.1') },
 *   { confirmations: 2, onSuccess: (r) => console.log('done!', r) },
 * )
 */
export function useTransaction(): UseTransactionReturn {
  const [state, setState] = useState<TransactionState>(initialState)

  const { data: walletClient } = useWalletClient()
  const publicClient = usePublicClient()

  // 用 ref 实现可取消：用户点了 cancel / 组件卸载，不再更新状态
  const cancelledRef = useRef(false)

  const reset = useCallback(() => {
    cancelledRef.current = false
    setState(initialState)
  }, [])

  const send = useCallback(
    async (tx: any, options: SendTransactionOptions = {}) => {
      const {
        confirmations = 1,
        timeout = 60_000,
        onSuccess,
        onError,
        gas,
      } = options

      if (!walletClient) {
        setState(s => ({ ...s, status: 'failed', error: '钱包未连接' }))
        return
      }

      cancelledRef.current = false

      // 阶段一：发送交易
      setState(s => ({ ...s, status: 'pending', error: null }))

      let hash: `0x${string}`
      try {
        hash = await walletClient.sendTransaction({
          ...tx,
          gas,
          account: walletClient.account,
        })
      } catch (e: any) {
        if (!cancelledRef.current) {
          const message = e.shortMessage || e.message || '交易发送失败'
          setState(s => ({ ...s, status: 'failed', error: message }))
          onError?.(e)
        }
        return
      }

      if (cancelledRef.current) return

      setState(s => ({ ...s, status: 'confirming', hash }))

      // 阶段二：等待交易回执（上链）
      let receipt: TransactionReceipt
      try {
        receipt = await withTimeout(
          publicClient!.waitForTransactionReceipt({
            hash,
            confirmations,
            timeout,
          }),
          timeout,
          `交易确认超时 (${timeout / 1000}s)`
        )
      } catch (e: any) {
        if (!cancelledRef.current) {
          const message = e.message || '交易失败'
          setState(s => ({ ...s, status: 'failed', error: message, hash }))
          onError?.(e)
        }
        return
      }

      if (cancelledRef.current) return

      // 阶段三：确认成功
      setState({
        status: 'confirmed',
        hash,
        error: null,
        confirmations,
        receipt,
      })

      onSuccess?.(receipt)
    },
    [walletClient, publicClient]
  )

  // 衍生状态
  const isProcessing = state.status === 'pending' || state.status === 'confirming'
  const statusText = statusTextMap[state.status]

  return {
    ...state,
    send,
    reset,
    isProcessing,
    statusText,
  }
}

/**
 * 带超时的 Promise
 */
async function withTimeout<T>(
  promise: Promise<T>,
  ms: number,
  message: string
): Promise<T> {
  const timeout = new Promise<never>((_, reject) =>
    setTimeout(() => reject(new Error(message)), ms)
  )
  return Promise.race([promise, timeout])
}
```

### 使用示例

```typescript
// pages/SwapPage.tsx — 完整的 Swap 流程
import { useState } from 'react'
import { parseEther } from 'viem'
import { useTransaction } from '../hooks/useTransaction'
import { useContractRead } from '../hooks/useContractRead'
import { useWallet } from '../hooks/useWallet'

function SwapPage() {
  const [amount, setAmount] = useState('')
  const { address, isConnected } = useWallet()

  // 交易状态由 useTransaction 管理
  const { send, status, hash, error, isProcessing, statusText, reset } = useTransaction()

  // 合约数据由 useContractRead 管理
  const { data: allowance } = useContractRead<bigint>(
    '0xTOKEN_ADDRESS',
    erc20Abi,
    'allowance',
    [address!, '0xROUTER_ADDRESS'],
    { enabled: !!address },
  )

  const handleSwap = async () => {
    await send(
      // 交易参数
      {
        to: '0xROUTER_ADDRESS',
        data: encodeSwapData(amount),
        value: parseEther(amount),
      },
      // 配置项
      {
        confirmations: 2,                    // 两次确认才认为成功
        timeout: 90_000,                     // 90 秒超时
        onSuccess: (receipt) => {
          console.log('Swap 成功! block:', receipt.blockNumber)
          // RTK Query 的 invalidatesTags 会自动刷新余额
        },
        onError: (err) => {
          console.error('Swap 失败:', err)
        },
      },
    )
  }

  if (!isConnected) return <div>请先连接钱包</div>

  return (
    <div>
      <input value={amount} onChange={e => setAmount(e.target.value)} placeholder="ETH 数量" />

      {/* 交易按钮：根据状态切换 UI */}
      <button onClick={handleSwap} disabled={isProcessing}>
        {isProcessing ? (
          <>
            <Spinner />
            {statusText}
            {status === 'confirming' && hash && (
              <a href={`https://etherscan.io/tx/${hash}`} target="_blank">
                查看 ↗
              </a>
            )}
          </>
        ) : (
          '确认 Swap'
        )}
      </button>

      {/* 错误提示 */}
      {status === 'failed' && (
        <div style={{ color: 'red' }}>
          <p>{error}</p>
          <button onClick={reset}>重试</button>
        </div>
      )}

      {/* 成功提示 */}
      {status === 'confirmed' && (
        <div style={{ color: 'green' }}>
          <p>Swap 成功!</p>
          <a href={`https://etherscan.io/tx/${hash}`} target="_blank">
            交易详情: {hash?.slice(0, 10)}... ↗
          </a>
        </div>
      )}
    </div>
  )
}
```

---

## Hook 组合使用示例

三个 Hook 在同一个页面中协同工作：

```typescript
// pages/TokenSwapPage.tsx — 完整示例
function TokenSwapPage() {
  // ===== useWallet — 钱包连接 =====
  const { address, isConnected, chainId, connect, switchChain } = useWallet()

  // ===== useContractRead — 查余额和授权额度 =====
  const { data: balance, loading: balanceLoading } = useContractRead<bigint>(
    TOKEN_ADDRESS,
    erc20Abi,
    'balanceOf',
    [address!],
    { enabled: !!address, retryCount: 2 },
  )

  const { data: allowance, loading: allowanceLoading } = useContractRead<bigint>(
    TOKEN_ADDRESS,
    erc20Abi,
    'allowance',
    [address!, ROUTER_ADDRESS],
    { enabled: !!address },
  )

  // ===== useTransaction — 管理交易发送 =====
  const approveTx = useTransaction()
  const swapTx = useTransaction()

  // ===== 业务逻辑 =====
  const needsApproval = allowance !== null && allowance === 0n

  const handleApprove = async () => {
    await approveTx.send(
      {
        to: TOKEN_ADDRESS,
        data: encodeApproveData(ROUTER_ADDRESS, MAX_UINT256),
      },
      {
        confirmations: 1,
        onSuccess: () => console.log('授权成功，可以 Swap 了'),
      },
    )
  }

  const handleSwap = async () => {
    await swapTx.send(
      {
        to: ROUTER_ADDRESS,
        data: encodeSwapData(amount, TOKEN_ADDRESS, WETH_ADDRESS),
      },
      {
        confirmations: 2,
        timeout: 120_000,
        onSuccess: (receipt) => console.log('Swap 成功!', receipt.transactionHash),
      },
    )
  }

  return (
    <div>
      {/* 钱包状态 */}
      {!isConnected ? (
        <button onClick={() => connect('metaMask')}>连接钱包</button>
      ) : (
        <p>钱包: {address?.slice(0, 6)}...{address?.slice(-4)}</p>
      )}

      {/* 余额 */}
      {balanceLoading ? (
        <p>加载余额...</p>
      ) : (
        <p>余额: {formatEther(balance ?? 0n)} TOKEN</p>
      )}

      {/* 授权或 Swap */}
      {needsApproval ? (
        <button onClick={handleApprove} disabled={approveTx.isProcessing}>
          {approveTx.isProcessing ? approveTx.statusText : '授权代币'}
        </button>
      ) : (
        <button onClick={handleSwap} disabled={swapTx.isProcessing}>
          {swapTx.isProcessing ? swapTx.statusText : 'Swap'}
        </button>
      )}

      {/* 交易状态展示 */}
      {(approveTx.isProcessing || swapTx.isProcessing) && (
        <TxProgress
          statusText={approveTx.isProcessing ? approveTx.statusText : swapTx.statusText}
          hash={approveTx.hash || swapTx.hash}
        />
      )}

      {/* 错误展示 */}
      {approveTx.status === 'failed' && <p style={{ color: 'red' }}>授权失败: {approveTx.error}</p>}
      {swapTx.status === 'failed' && <p style={{ color: 'red' }}>Swap 失败: {swapTx.error}</p>}
    </div>
  )
}
```

---

## 面试口述话术

> "useEffect 的依赖数组本质是浅比较——React 在每次渲染后，用 `Object.is` 对比依赖数组里每个值是否变了，变了就重新执行 effect。三种形态：不传参每次渲染都执行、空数组只执行一次相当于 Vue 的 mounted、有依赖的话依赖变化就重新执行。需要注意对象和数组引用的问题——每次渲染都创建新引用，直接放依赖会导致无意义的重执行，需要用 useMemo/useCallback 固定引用。
>
> 在 Web3 项目里我封装了三个核心自定义 Hook：
>
> **useWallet**——封装 Wagmi 底层的 useAccount、useConnect、useDisconnect，对外暴露简单的 connect/disconnect/switchChain API，同时做了统一错误处理和派生状态（shortAddress、isMainnet、explorerUrl）。好处是如果将来换钱包库，只改这一个文件就行。
>
> **useContractRead**——封装 Viem publicClient.readContract，自动管理 loading、data、error 三种状态，支持重试、参数变化自动重取、组件卸载取消请求。把合约调用失败的错误信息解析成用户友好的中文提示。
>
> **useTransaction**——管理一笔交易的完整生命周期：idle → pending（提交到 mempool）→ confirming（等待区块确认）→ confirmed/failed。支持自定义确认数、超时时间、成功/失败回调，还有 isProcessing 和 statusText 方便 UI 展示。
>
> 三个 Hook 组合使用——useWallet 提供钱包、useContractRead 读合约数据、useTransaction 发送交易，每个 Hook 只做一件事，底层复杂逻辑全部封装好，上层组件只关心业务。CoderWhy 课程专门有一章讲自定义 Hook 的设计思想——把可复用状态逻辑从组件里抽出来。"
