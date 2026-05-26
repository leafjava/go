package main

// 后台事件处理器 — 永远不退出
func eventLoopLeak(eventCh chan Event) {
	go func() {
		for {
			select {
			case event := <-eventCh:
				processEvent(event)
				// ⚠️ 没有 ctx.Done() 退出分支
				// 服务关闭时这个 goroutine 永远卡在这里
			}
		}
	}()
}
