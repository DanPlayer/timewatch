# Timewatch
To [Chinese Doc](https://github.com/DanPlayer/timewatch/blob/main/README_zh.md)

Timewatch is a distributed monitoring task scheduling tool that can mark and recover lost tasks after service restart. It is written in Go and requires Redis as a cache. It provides some simple APIs, such as AfterFunc, Reset and Stop, to add, modify and stop monitoring tasks. It also supports custom attributes that can be used in check restart. The purpose of this project is to improve the reliability and flexibility of monitoring tasks.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [Create a monitoring service](#create-a-monitoring-service)
  - [Start the monitoring service](#start-the-monitoring-service)
  - [Add a monitoring task](#add-a-monitoring-task)
  - [Modify or stop a monitoring task](#modify-or-stop-a-monitoring-task)
- [References](#references)
- [Contact Information](#contact-information)

## Installation

Use the following command to install Timewatch:

```bash
$ go get -u github.com/DanPlayer/timewatch
```

Before installing, please make sure your system meets the following requirements:

- Go version >= 1.16
- Redis version >= 6.0

## Usage

### Create a monitoring service

First, you need to create a monitoring service, specify a unique Key, a cache instance (currently only supports Redis), and whether to enable timeout behavior (if enabled, the task will be executed once when it times out):

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

### Start the monitoring service

Then, you need to start the monitoring service, which will create a background coroutine to check the tasks in the cache:

```go
// start watch service
err := watch.Start()
if err != nil {
    fmt.Println(err)
}
```

You can also use the StartWithCheckRestart method to start the monitoring service, and check if there are any lost tasks caused by abnormal shutdown at startup. If so, it will call the function you provide to recover the tasks:

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

### Add a monitoring task

Next, you can use the AfterFunc method to add a monitoring task, specify a delay time, a Watch structure (containing a field name and some custom attributes), and a function to execute:

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

You can also use custom attributes to store some extra information that can be used in check restart:

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

### Modify or stop a monitoring task

Finally, you can use the Reset method to modify the delay time of a monitoring task:

```go
// watch reset
timer.Reset(10 * time.Second) // change the delay time to 10 seconds
```

Or use the Stop method to stop a monitoring task:

```go
// watch stop
timer.Stop() // stop the task and remove it from the cache
```

## Improvements

To make this document clearer and more complete, I suggest you can make the following improvements:

- Add a table of contents at the beginning of the document for easy navigation to interested sections.
- Add some system requirements and dependency instructions in the installation section, such as Go version and Redis version.
- Add some code comments and output results in the usage section to make it easier for readers to understand each API's functionality and effect.
- Add some contact information or feedback channels at the end of the document so that readers can ask you questions or suggestions.

## References

- [GitHub - DanPlayer/timewatch](https://github.com/DanPlayer/timewatch)

## Contact Information

If you have any questions or suggestions, please contact me through the following ways:

- Email: 344080699@qq.com
- GitHub: https://github.com/DanPlayer
