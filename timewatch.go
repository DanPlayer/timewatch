package timewatch

import (
	"encoding/json"
	"errors"
	"github.com/rfyiamcool/go-timewheel"
	"time"
)

type TimeWatch struct {
	key        string                      // marked key
	watch      Watch                       // watched attributes
	cache      Cache                       // Redis and MemoryCache and more...
	Timer      map[string]*timewheel.Timer // watch map key is Watch.Field
	outTimeAct bool                        // out time to action
	wheel      *timewheel.TimeWheel        // time wheel timer
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
	return &TimeWatch{
		key:        options.Key,
		cache:      options.Cache,
		outTimeAct: options.OutTimeAct,
		Timer:      map[string]*timewheel.Timer{},
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
	if locked {
		return errors.New("locked by cache")
	}
	defer w.unlock()

	// delete stop by abnormal shutdown
	all, err := w.cache.HGetAll(w.key)
	if err != nil {
		return err
	}

	for k, s := range all {
		var info Watch
		_ = json.Unmarshal([]byte(s), &info)

		_ = w.cache.HDel(w.key, k)
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
	if locked {
		return errors.New("locked by cache")
	}
	defer w.unlock()

	all, err := w.cache.HGetAll(w.key)
	if err != nil {
		return err
	}

	for k, s := range all {
		var info Watch
		_ = json.Unmarshal([]byte(s), &info)

		_ = w.cache.HDel(w.key, k)

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
	err = w.cache.HSet(w.key, c.Field, string(bytes))
	if err != nil {
		return
	}
	timer := w.wheel.AfterFunc(t, func() {
		_ = w.cache.HDel(w.key, c.Field)
		f()
	})
	w.Timer[c.Field] = timer
	return timer, nil
}

func (w *TimeWatch) Stop(field string) {
	timer, ok := w.Timer[field]
	if !ok {
		return
	}
	_ = w.cache.HDel(w.key, field)
	timer.Stop()
}

func (w *TimeWatch) Reset(field string, d time.Duration) {
	timer, ok := w.Timer[field]
	if !ok {
		return
	}

	get, err := w.cache.HGet(w.key, field)
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
	err = w.cache.HSet(w.key, c.Field, string(bytes))
	if err != nil {
		return
	}

	timer.Reset(d)
}

const LockKey = "CheckLock"

func (w *TimeWatch) lock() (bool, error) {
	if w.key == "" {
		return false, errors.New("miss lock key")
	}
	return w.cache.SetNX(w.lockKey(), "LOCKED", 60*time.Second)
}

func (w *TimeWatch) unlock() {
	_ = w.cache.Del(w.lockKey())
}

func (w TimeWatch) lockKey() string {
	return w.key + ":" + LockKey
}
