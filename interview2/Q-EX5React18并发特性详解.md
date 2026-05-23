# React 18 并发特性（Concurrent Features）详解

---

## 前置知识：React 18 的核心突破

React 18 之前，渲染是**同步且不可中断**的——一次 `setState` 触发重渲染，React 必须一口气跑完整个组件树，中间不能停下来处理别的事。

这就导致一个问题：如果渲染很慢（比如一个复杂的列表/图表），用户的**下一次点击或输入就会被卡住**，页面看起来像"冻住了"。

React 18 的答案：**Concurrent Mode（并发模式）**——渲染可以被中断、暂停、恢复，React 会在合适的时机"插队"处理更高优先级的更新。

```
React 17（同步渲染）：
  用户输入 → 排队等 → 等前面渲染完 → 才能响应
  ─────────────────────────────────────→
  [      慢渲染      ][   响应用户输入   ]

React 18（并发渲染）：
  用户输入 → 立即打断慢渲染 → 先响应用户 → 再继续慢渲染
  ─────────────────────────────────────→
  [ 慢渲染 ][ 用户输入 ][   继续慢渲染  ]
```

---

## 一、useTransition —— 标记低优先级更新

### 是什么？

`useTransition` 让你把某些状态更新标记为**低优先级**。当高优先级更新（如用户输入）进来时，低优先级渲染会被打断，优先响应用户。

### 基本用法

```jsx
import { useState, useTransition } from 'react';

function SearchPage() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [isPending, startTransition] = useTransition();

  function handleChange(e) {
    const value = e.target.value;

    // ✅ 高优先级：立即更新输入框的文字
    setQuery(value);

    // ✅ 低优先级：搜索结果可以慢一点
    startTransition(() => {
      // 假设这是一个开销很大的过滤操作
      const filtered = hugeList.filter(item =>
        item.title.includes(value)
      );
      setResults(filtered);
    });
  }

  return (
    <div>
      <input value={query} onChange={handleChange} />
      {/* isPending 为 true 时显示加载态 */}
      {isPending && <span>搜索中...</span>}
      <ResultList results={results} />
    </div>
  );
}
```

### 效果对比

```
❌ 不用 useTransition：
   用户快速输入 "hello"
   → h 触发全量过滤（卡 200ms）
   → he 触发全量过滤（卡 200ms）
   → hel 触发全量过滤（卡 200ms）
   → ...
   用户感觉：打字的时候页面卡顿，输入框跟不上手指

✅ 用 useTransition：
   用户快速输入 "hello"
   → h → he → hel → hell → hello（每次输入框都立即响应）
   → React 发现搜索还在进行中，丢弃中间的过滤
   → 只执行最后一次（hello）的过滤
   用户感觉：打字流畅，结果稍后出现
```

### 适合的场景

- 搜索框输入 + 结果列表过滤
- Tab 切换 + 不同 Tab 的内容渲染
- 筛选/排序操作，UI 状态立即响应但数据计算可以慢一点

---

## 二、useDeferredValue —— 延迟使用旧值

### 是什么？

`useDeferredValue` 告诉 React："这个值如果不着急，可以先用旧版本，等空闲了再更新"。

### 和 useTransition 的区别

| | useTransition | useDeferredValue |
|---|---|---|
| 谁控制优先级 | **你**在 setState 时标记 | React 自动决定 |
| 使用方式 | 包裹 `startTransition(() => setXxx())` | 包裹值 `const d = useDeferredValue(value)` |
| 何时用 | 你能控制 setState 的地方 | 值来自外部（props / 第三方库），你无法控制 setState |

### 基本用法

```jsx
import { useState, useDeferredValue, useMemo } from 'react';

function ProductList({ searchQuery }) {
  // searchQuery 从 props 传进来，你没法"标记它为低优先级"
  // 用 useDeferredValue 让 React 延迟更新
  const deferredQuery = useDeferredValue(searchQuery);

  // 用 deferredQuery 做重渲染，用 searchQuery 显示加载态
  const isStale = searchQuery !== deferredQuery;

  const filteredProducts = useMemo(() => {
    return hugeProductList.filter(p =>
      p.name.includes(deferredQuery)
    );
  }, [deferredQuery]);

  return (
    <div style={{ opacity: isStale ? 0.5 : 1 }}>
      {filteredProducts.map(p => (
        <ProductCard key={p.id} product={p} />
      ))}
    </div>
  );
}
```

### 效果

```
searchQuery 变化:
  "iP" → "iPh" → "iPho" → "iPhon" → "iPhone"

deferredQuery 可能只更新:
  "iP"  →  ..........  →  "iPhone"
  （中间值被跳过，只保留最后一次）

同时组件:
  - opacity 变半透明提示"数据不是最新的"
  - 用旧列表先顶着，不让页面白屏
```

---

## 三、Suspense —— 数据加载时的"悬念"

### React 为什么叫它"悬念"？

"Suspense" 取的是 "suspend"（挂起/暂停）的含义：

- 组件在数据没到齐时，**先"悬停"，不崩溃**
- 就像电影制造悬念——主角推开门，画面黑了，观众"悬着心"等揭晓
- React 也是：组件"悬着等你"（等待数据），等数据到了继续渲染

**一句话**：Suspense = "数据没到？先挂起，别崩溃，到了再继续"。

### 基本用法：数据加载

```jsx
import { Suspense } from 'react';

// 一个有 Suspense 能力的组件
function UserProfile({ userId }) {
  // 这个 hook 内部会"抛出一个 Promise"，React 捕获后"挂起"渲染
  const user = useUser(userId); // 伪代码，实际用 React Query / SWR / Relay 等

  return (
    <div>
      <h2>{user.name}</h2>
      <p>{user.bio}</p>
    </div>
  );
}

function App() {
  return (
    <div>
      <h1>首页</h1>

      {/* Suspense 边界：UserProfile 没加载完时，显示 fallback */}
      <Suspense fallback={<div>用户信息加载中...</div>}>
        <UserProfile userId={1} />
      </Suspense>
    </div>
  );
}
```

### 核心优势：不同区块独立加载

```
❌ 传统做法（没有 Suspense）：
   if (loading1 || loading2 || loading3) return <Spinner />
   → 三个组件有一个没加载完，整个页面都显示 Spinner

✅ Suspense 做法：
   <Suspense fallback={<Skeleton1 />}>
     <Widget1 />
   </Suspense>
   <Suspense fallback={<Skeleton2 />}>
     <Widget2 />
   </Suspense>
   <Suspense fallback={<Skeleton3 />}>
     <Widget3 />
   </Suspense>
   → 每个 Widget 各自加载，谁先好谁先显示
   → Widget2 慢了？不影响 Widget1 和 Widget3
```

### 真实例子：配合 React Query / TanStack Query

```jsx
import { Suspense } from 'react';
import { useSuspenseQuery } from '@tanstack/react-query';

function WalletBalance() {
  const { data } = useSuspenseQuery({
    queryKey: ['wallet-balance'],
    queryFn: () => fetch('/api/wallet').then(r => r.json()),
  });

  return <div>余额: {data.balance} ETH</div>;
}

function TransactionHistory() {
  const { data } = useSuspenseQuery({
    queryKey: ['transactions'],
    queryFn: () => fetch('/api/transactions').then(r => r.json()),
  });

  return (
    <ul>
      {data.map(tx => (
        <li key={tx.hash}>{tx.amount} ETH → {tx.to}</li>
      ))}
    </ul>
  );
}

function Dashboard() {
  return (
    <div>
      {/* 余额先加载好就先显示，不被交易列表拖累 */}
      <Suspense fallback={<div className="skeleton">余额加载中...</div>}>
        <WalletBalance />
      </Suspense>

      {/* 交易列表慢也不影响余额显示 */}
      <Suspense fallback={<div className="skeleton">交易记录加载中...</div>}>
        <TransactionHistory />
      </Suspense>
    </div>
  );
}
```

---

## 四、三者对比总结

| 特性 | 解决什么问题 | 使用场景 | 本质 |
|------|------------|---------|------|
| `useTransition` | 低优先级更新卡住用户输入 | 搜索框、Tab 切换 | 你主动标记"这个 setState 不重要" |
| `useDeferredValue` | 值变化频繁导致的重渲染 | 第三方库传入的频繁变化的值 | React 自动延迟"不太重要的值" |
| `Suspense` | 数据加载时整页白屏 | 任何需要异步数据的组件 | 组件"挂起"等待数据，不崩溃 |

---

## 五、Web3 实战场景

```jsx
import { useState, useTransition, Suspense } from 'react';

function DApp() {
  // 钱包连接状态 → 高优先级（立即响应）
  const [account, setAccount] = useState(null);

  // 交易历史 → 低优先级（可以慢一点）
  const [txs, setTxs] = useState([]);
  const [isPending, startTransition] = useTransition();

  async function connectWallet() {
    // ✅ 高优先级：钱包连接状态立即更新
    const accounts = await window.ethereum.request({
      method: 'eth_requestAccounts'
    });
    setAccount(accounts[0]);

    // ✅ 低优先级：获取历史交易慢慢来
    startTransition(async () => {
      const history = await fetchTxHistory(accounts[0]);
      setTxs(history);
    });
  }

  return (
    <div>
      {/* 钱包状态：Suspense 保证不阻塞页面 */}
      <Suspense fallback={<button disabled>连接中...</button>}>
        <WalletCard account={account} onConnect={connectWallet} />
      </Suspense>

      {/* 交易列表：低优先级 + 独立 Suspense */}
      <Suspense fallback={<div>交易记录加载中...</div>}>
        <TransactionList
          txs={txs}
          loading={isPending}
        />
      </Suspense>
    </div>
  );
}
```

**优先级分配原则**：

| 高优先级（立即响应） | 低优先级（可以等） |
|-------------------|------------------|
| 钱包连接/断开 | 历史交易列表 |
| 余额变化 | NFT 图片加载 |
| 输入框文字 | 搜索/过滤结果 |
| 按钮点击反馈 | 数据统计图表 |

---

## 六、面试口述话术（30 秒版）

> "React 18 的核心突破是——渲染可以被中断了。以前一次 setState 触发重渲染就必须跑完，现在可以用 useTransition 标记低优先级更新，比如搜索框输入——文字本身立即显示（高优先级），搜索结果列表延迟渲染（低优先级），用户打字时不会被搜索结果卡住。useDeferredValue 也是类似思路，允许组件先用旧值渲染，等空闲了再更新。Suspense 搭配数据获取组件可以做到——不同区块独立加载，一个慢了不影响另一个。对 Web3 来说特别有用，比如钱包状态用高优先级保持响应，历史交易列表用低优先级慢慢加载。"

---

## 七、要点速记

- 渲染可中断——`useTransition` 标记低优先级更新，高优先级任务（输入文字）不会被低优先级（搜索结果）卡住
- `useDeferredValue`——组件先用旧值渲染，空闲后再更新，适合列表/图表等重渲染场景
- Suspense——组件数据没到先"挂起"，不崩溃不倒逼整页白屏，不同区块各自"悬"，一个慢了不影响另一个
- Web3 实践——钱包状态走高优先级保持响应，历史交易列表走低优先级慢慢加载
