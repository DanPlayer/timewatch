package timewatch

import (
	"encoding/json"
	"errors"
	"github.com/rfyiamcool/go-timewheel"
	"net"
	"strconv"
	"sync"
	"time"
)

type TimeWatch struct {
	machineID  uint16               // machine mark
	key        string               // marked key
	watch      Watch                // watched attributes
	cache      Cache                // Redis and MemoryCache and more...
	Timer      sync.Map             // watch map key is Watch.Field
	outTimeAct bool                 // out time to action
	wheel      *timewheel.TimeWheel // time wheel timer
}

type Options struct {
	Key        string        // marked key
	Cache      Cache         // Redis and MemoryCache and more...
	OutTimeAct bool          // out time to action
	Tick       time.Duration // time wheel scale, default 1 * time.Second
	BucketsNum int           // Time Roulette, default 360
}

type Watch struct {
	Field                string      `json:"field"`                 // watched name
	TouchOffUnix         int64       `json:"touch_off_unix"`        // watch func trigger time, zero will auto set up
	CustomizedAttributes interface{} `json:"customized_attributes"` // customized struct or other
}

func Service(options Options) *TimeWatch {
	ip, err := lower16BitPrivateIP()
	if err != nil {
		return nil
	}
	return &TimeWatch{
		machineID:  ip,
		key:        options.Key,
		cache:      options.Cache,
		outTimeAct: options.OutTimeAct,
		Timer:      sync.Map{},
		wheel:      newWheel(options.Tick, options.BucketsNum),
	}
}

func newWheel(tick time.Duration, buckets int) *timewheel.TimeWheel {
	if tick == 0 {
		tick = 1 * time.Second
	}
	if buckets == 0 {
		buckets = 360
	}
	tw, err := timewheel.NewTimeWheel(tick, buckets, timewheel.TickSafeMode())
	if err != nil {
		panic(err)
	}
	return tw
}

func (w *TimeWatch) Start() error {
	locked, err := w.lock()
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("locked by cache")
	}
	defer w.unlock()

	// delete stop by abnormal shutdown
	all, err := w.cache.HGetAll(w.getCacheKey())
	if err != nil {
		return err
	}

	for k, s := range all {
		var info Watch
		_ = json.Unmarshal([]byte(s), &info)

		_ = w.cache.HDel(w.getCacheKey(), k)
	}

	return nil
}

// StartWithCheckRestart
// check and restart task that stop by abnormal shutdown
// use it in program service start
// func(c Watch) c is watched attributes
func (w *TimeWatch) StartWithCheckRestart(fc func(c Watch)) error {
	locked, err := w.lock()
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("locked by cache")
	}
	defer w.unlock()

	all, err := w.cache.HGetAll(w.getCacheKey())
	if err != nil {
		return err
	}

	for k, s := range all {
		var info Watch
		_ = json.Unmarshal([]byte(s), &info)

		_ = w.cache.HDel(w.getCacheKey(), k)

		left := time.Duration(time.Now().Unix()-info.TouchOffUnix) * time.Second
		if left > 0 {
			w.wheel.AfterFunc(left, func() {
				fc(info)
			})
		} else {
			if w.outTimeAct {
				fc(info)
			}
		}
	}
	return nil
}

func (w *TimeWatch) AfterFunc(t time.Duration, c Watch, f func()) (r *timewheel.Timer, err error) {
	if c.Field == "" {
		return nil, errors.New("field is empty")
	}
	if c.TouchOffUnix == 0 {
		c.TouchOffUnix = time.Now().Unix() + int64(t.Seconds())
	}
	bytes, _ := json.Marshal(c)
	err = w.cache.HSet(w.getCacheKey(), c.Field, string(bytes))
	if err != nil {
		return
	}
	timer := w.wheel.AfterFunc(t, func() {
		_ = w.cache.HDel(w.getCacheKey(), c.Field)
		f()
	})
	w.Timer.LoadOrStore(c.Field, timer)
	return timer, nil
}

func (w *TimeWatch) Stop(field string) {
	var timer *timewheel.Timer
	load, ok := w.Timer.Load(field)
	if ok {
		timer = load.(*timewheel.Timer)
		_ = w.cache.HDel(w.getCacheKey(), field)
		timer.Stop()
	}
}

func (w *TimeWatch) Reset(field string, d time.Duration) {
	var timer *timewheel.Timer
	load, ok := w.Timer.Load(field)
	if ok {
		timer = load.(*timewheel.Timer)
		get, err := w.cache.HGet(w.getCacheKey(), field)
		if err != nil {
			return
		}
		var c Watch
		err = json.Unmarshal([]byte(get), &c)
		if err != nil {
			return
		}
		c.TouchOffUnix = time.Now().Unix() + int64(d.Seconds())
		bytes, _ := json.Marshal(c)
		err = w.cache.HSet(w.getCacheKey(), c.Field, string(bytes))
		if err != nil {
			return
		}

		timer.Reset(d)
	}
}

func (w *TimeWatch) getCacheKey() string {
	return strconv.Itoa(int(w.machineID)) + ":" + w.key
}

const LockKey = "CheckLock"

func (w *TimeWatch) lock() (bool, error) {
	if w.getCacheKey() == "" {
		return false, errors.New("miss lock key")
	}
	return w.cache.SetNX(w.lockKey(), "LOCKED", 60*time.Second)
}

func (w *TimeWatch) unlock() {
	_ = w.cache.Del(w.lockKey())
}

func (w *TimeWatch) lockKey() string {
	return w.getCacheKey() + ":" + LockKey
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}
