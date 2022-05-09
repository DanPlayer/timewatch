# Timewatch
To [Chinese Doc](https://github.com/DanPlayer/timewatch/blob/main/README_zh.md)

watch task plan

monitoring task plan, general monitoring plan will lose monitoring ability due to service restart

Timewatch can tag lost task and restart it

### Installation
```
$ go get -u github.com/DanPlayer/timewatch
```

### Example
```
func TestSimpleExample(t *testing.T) {
	var watch = timewatch.Service(timewatch.Options{
		Key:        "MsgWatch",
		Cache:      cache.NewRedis(cache.RedisOptions{
			Addr:     "127.0.0.1",
			Password: "",
			DB:       0,
		}),
		OutTimeAct: true,
	})

	// check for exception shutdown and restart watch task
	err := watch.CheckRestart(func(c timewatch.Watch) {
		fmt.Println(c)
		fmt.Println("do that u want")
	})
	if err != nil {
		fmt.Println(err)
	}

	// watch plan add
	timer, err := watch.AfterFunc(5*time.Second, timewatch.Watch{
		Field:                "TestField",
		CustomizedAttributes: nil, // could use some self make that u want set attributes in watch.CheckRestart
	}, func() {
		fmt.Println("plan to func")
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// watch reset
	timer.Reset(10*time.Second)

	time.Sleep(11*time.Second)

	// watch stop
	timer.Stop()
}
```