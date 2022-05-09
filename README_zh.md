# Timewatch 监控计划
监控任务计划，一般监控计划会因为服务重启而失去监控能力

Timewatch 可以标记已经失去的监控计划并重启它

### 安装
```
$ go get -u github.com/DanPlayer/timewatch
```

### 使用案例
```
func TestSimpleExample(t *testing.T) {
    // 初始化监控计划
	var watch = timewatch.Service(timewatch.Options{
		Key:        "MsgWatch", // 监控的key
		Cache:      cache.NewRedis(cache.RedisOptions{
			Addr:     "127.0.0.1",
			Password: "",
			DB:       0,
		}), // 缓存
		OutTimeAct: true, // 重启异常失败的监控时是否执行已经失效的计划
	})

	// 检查异常关闭的监控计划而且重启它们（不需要重启异常的监控计划可以不使用）
	err := watch.CheckRestart(func(c timewatch.Watch) {
		fmt.Println(c)
		fmt.Println("do that u want")
	})
	if err != nil {
		fmt.Println(err)
	}

	// 监控计划增加
	timer, err := watch.AfterFunc(5*time.Second, timewatch.Watch{
		Field:                "TestField",
		CustomizedAttributes: nil, // 自定义的属性参数在 watch.CheckRestart 中使用
	}, func() {
		fmt.Println("plan to func")
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// 重新设置监控计划
	timer.Reset(10*time.Second)

	time.Sleep(11*time.Second)

	// 停止监控计划
	timer.Stop()
}
```