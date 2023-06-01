# Timewatch

Timewatch是一个分布式监控任务计划的工具，它可以在服务重启后标记和恢复丢失的任务。它使用Go语言编写，需要安装Redis作为缓存。它提供了一些简单的API，如AfterFunc，Reset和Stop，来添加，修改和停止监控任务。它还支持自定义属性，可以在检查重启时使用。这个项目的目的是提高监控任务的可靠性和灵活性。

## 目录

- [安装](#安装)
- [使用](#使用)
  - [创建一个监控服务](#创建一个监控服务)
  - [启动监控服务](#启动监控服务)
  - [添加一个监控任务](#添加一个监控任务)
  - [修改或停止一个监控任务](#修改或停止一个监控任务)
- [参考](#参考)
- [联系方式](#联系方式)

## 安装

使用以下命令安装Timewatch：

```bash
$ go get -u github.com/DanPlayer/timewatch
```

安装前，请确保你的系统满足以下要求：

- Go版本 >= 1.16
- Redis版本 >= 6.0

## 使用

### 创建一个监控服务

首先，你需要创建一个监控服务，指定一个唯一的Key，一个缓存实例（目前只支持Redis），以及是否开启超时行为（如果开启，当任务超时时会执行一次）：

```go
var watch = timewatch.Service(timewatch.Options{
    Key:       "MsgWatch",
    Cache:     cache.NewRedis(cache.RedisOptions{
        Addr:     "127.0.0.1:6379",
        Password: "",
        DB:       0,
    }),
    OutTimeAct: true,
})
```

### 启动监控服务

然后，你需要启动监控服务，这会创建一个后台协程来检查缓存中的任务：

```go
// start watch service
err := watch.Start()
if err != nil {
    fmt.Println(err)
}
```

你也可以使用StartWithCheckRestart方法来启动监控服务，并且在启动时检查是否有异常关闭导致的丢失任务，如果有，就会调用你提供的函数来恢复任务：

```go
// check for exception shutdown and restart watch task
err = watch.StartWithCheckRestart(func(c timewatch.Watch) {
    fmt.Println(c) // print the watch struct
    infoMap := c.CustomizedAttributes.(map[string]interface{})
    marshal, _ := json.Marshal(infoMap)
    var info User
    _ = json.Unmarshal(marshal, &info)

    fmt.Println(fmt.Sprintf("User struct name: %s", info.Name))
    fmt.Println(fmt.Sprintf("User struct age: %d", info.Age))
    fmt.Println("do that u want")
})
if err != nil {
    fmt.Println(err)
}
```

### 添加一个监控任务

接下来，你可以使用AfterFunc方法来添加一个监控任务，指定一个延迟时间，一个Watch结构体（包含一个字段名和一些自定义属性），以及一个要执行的函数：

```go
// watch plan add
timer, err := watch.AfterFunc(5*time.Second, timewatch.Watch{
    Field:                "TestField",
    CustomizedAttributes: nil, // could use some self make that u want set attributes in watch.CheckRestart
}, func() {
    fmt.Println("plan to func") // print the message when the task is executed
})
if err != nil {
    fmt.Println(err)
    return
}
```

你也可以使用自定义属性来存储一些额外的信息，在检查重启时可以使用：

```go
type User struct {
    Name string
    Age  int
}

// watch plan add
timer, err := watch.AfterFunc(5*time.Second, timewatch.Watch{
    Field: "TestField",
    CustomizedAttributes: User{
        Name: "Dan",
        Age:  20,
    }, // could use some self make that u want set attributes in watch.CheckRestart
}, func() {
    fmt.Println("plan to func") // print the message when the task is executed
})
if err != nil {
    fmt.Println(err)
    return
}
```

### 修改或停止一个监控任务

最后，你可以使用Reset方法来修改一个监控任务的延迟时间：

```go
// watch reset
timer.Reset(10 * time.Second) // change the delay time to 10 seconds
```

或者使用Stop方法来停止一个监控任务：

```go
// watch stop
timer.Stop() // stop the task and remove it from the cache
```

## 参考

- [GitHub - DanPlayer/timewatch](https://github.com/DanPlayer/timewatch)

## 联系方式

如果你有任何问题或建议，请通过以下方式联系我：

- Email: 344080699@qq.com
- GitHub: https://github.com/DanPlayer
