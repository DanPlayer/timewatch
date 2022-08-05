package tests

import (
	"encoding/json"
	"fmt"
	"github.com/DanPlayer/timewatch"
	"github.com/DanPlayer/timewatch/cache"
	"testing"
	"time"
)

func main(m *testing.M) {
	fmt.Println("here is main test")
}

func TestSimpleExample(t *testing.T) {
	var watch = timewatch.Service(timewatch.Options{
		Key: "MsgWatch",
		Cache: cache.NewRedis(cache.RedisOptions{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		}),
		OutTimeAct: true,
	})

	// start watch service
	err := watch.Start()
	if err != nil {
		fmt.Println(err)
	}

	// watch plan add
	_, err = watch.AfterFunc(5*time.Second, timewatch.Watch{
		Field:                "TestField",
		CustomizedAttributes: nil, // could use some self make that u want set attributes in watch.StartWithCheckRestart
	}, func() {
		fmt.Println("plan to func")
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// watch reset
	watch.Reset("TestField", 10 * time.Second)

	time.Sleep(15 * time.Second)

	// watch stop
	watch.Stop("TestField")
}

func TestCustomizedAttributesExample(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	var watch = timewatch.Service(timewatch.Options{
		Key: "MsgWatch",
		Cache: cache.NewRedis(cache.RedisOptions{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		}),
		OutTimeAct: true,
	})

	// watch plan add
	timer, err := watch.AfterFunc(5*time.Second, timewatch.Watch{
		Field: "TestField",
		CustomizedAttributes: User{
			Name: "Dan",
			Age:  20,
		}, // could use some self make that u want set attributes in watch.StartWithCheckRestart
	}, func() {
		fmt.Println("plan to func")
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// check for exception shutdown and restart watch task
	err = watch.StartWithCheckRestart(func(c timewatch.Watch) {
		fmt.Println(c)
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

	// watch reset
	timer.Reset(10 * time.Second)

	time.Sleep(11 * time.Second)

	// watch stop
	timer.Stop()
}
