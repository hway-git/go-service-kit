package tool

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
)

const (
	// 全局定义时间格式化
	TIME_FORMAT_LAYOUT string = "2006-01-02 15:04:05"
)

// ##########  时间  ############
// 当前时间：2024-07-20 13:32:00
func GetCurrTime(tz ...string) string {
	if len(tz) == 0 {
		return time.Now().Format(TIME_FORMAT_LAYOUT)
	}
	loc, err := time.LoadLocation(tz[0])
	if err != nil {
		return ""
	}
	currTime := time.Now()
	time := currTime.In(loc).Format(TIME_FORMAT_LAYOUT)
	return time
}

// 时间戳 转 时间
// 1722215506 => 2024-07-29 09:11:46
func GetTimestampToTimeZone(timestamp int64, tz string) string {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return ""
	}
	TimeStr := time.Unix(timestamp, 0).In(loc)
	time := TimeStr.Format(TIME_FORMAT_LAYOUT)
	return time
}

// 时间 转 时间戳
// 2024-07-29 09:11:46 => 1722215506
func GetTimeZoneToTimestamp(timeStr string, tz string) (int64, error) {
	loc, err := time.LoadLocation(tz) // 转换成时区的时间对象
	if err != nil {
		return 0, err
	}
	t, err := time.ParseInLocation(TIME_FORMAT_LAYOUT, timeStr, loc) // 解析时间字符串
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil // 转换成时间戳
}

// 获取 明日的 00:00:00 时间戳（秒）
func GetTomorrowMidnightTimestamp(tz string) (int64, error) {
	// 加载指定时区
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return 0, err
	}
	// 获取当前时间，并转换为指定时区
	now := time.Now().In(loc)
	// 计算明日的 00:00:00
	tomorrow := now.AddDate(0, 0, 1)
	midnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)
	// 返回时间戳（秒）
	return midnight.Unix(), nil
}

func GetMonthsBetweenTimestamp(start, end int64) []string {
	var months []string
	startTime := time.Unix(start, 0).UTC()
	endTime := time.Unix(end, 0).UTC()
	for t := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, time.UTC); !t.After(endTime); t = t.AddDate(0, 1, 0) {
		months = append(months, t.Format("2006-01"))
	}
	return months
}

// 获取指定月份的日期列表 month: 2021-01
// 获取指定月份的日期列表 month: 2021-01
func GetMonthDates(month string) []string {
	// 格式化月份字符串
	month = strings.TrimSpace(month)
	if len(month) != 7 && len(month) != 6 {
		return []string{}
	}

	// 补上 -
	if len(month) == 6 {
		month = month[:4] + "-" + month[4:]
	}

	var dates []string
	startTime, err := time.Parse("2006-01", month)
	if err != nil {
		return dates
	}

	// 获取该月的第一天和最后一天
	firstDay := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, startTime.Location())
	lastDay := firstDay.AddDate(0, 1, -1) // 下个月的第一天减一天

	// 生成该月所有日期
	for current := firstDay; !current.After(lastDay); current = current.AddDate(0, 0, 1) {
		dates = append(dates, fmt.Sprintf("%d-%02d-%02d", current.Year(), current.Month(), current.Day()))
	}

	return dates
}

// 获取该月的第一天 00:00:00
func GetMonthFirstTimestamp(month string) (int64, error) {
	month = strings.ReplaceAll(month, "-", "") // 去掉-
	// 解析年份和月份
	t, err := time.Parse("200601", month)
	if err != nil {
		return time.Now().Unix(), err
	}

	startTime := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return startTime.Unix(), nil
}

// 获取该月的最后一天 23:59:59
func GetMonthLastTimestamp(month string) (int64, error) {
	month = strings.ReplaceAll(month, "-", "") // 去掉-
	// 解析年份和月份
	t, err := time.Parse("200601", month)
	if err != nil {
		return time.Now().Unix(), err
	}
	startTime := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC) // 获取该月的第一天 00:00:00
	endTime := startTime.AddDate(0, 1, 0).Add(-time.Second)              // 计算下个月的第一天，再减去 1 秒得到该月的最后一秒
	return endTime.Unix(), nil
}

// 获取今天的 00:00:00 到 23:59:59 时间戳（秒）
func GetTodayTimestamps(timezone string) (startSec, endSec int64, err error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, 0, err
	}
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)
	return startOfDay.Unix(), endOfDay.Unix(), nil
}

// 获取某日的 00:00:00 到 23:59:59 时间戳（秒）
func GetDateTimestamps(date string, timezone string) (startSec, endSec int64, err error) {
	if len(date) != 10 {
		if len(date) == 8 {
			date = date[:4] + "-" + date[4:6] + "-" + date[6:]
		} else {
			return 0, 0, fmt.Errorf("invalid date format: %s", date)
		}
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, 0, err
	}
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return 0, 0, err
	}
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)
	return startOfDay.Unix(), endOfDay.Unix(), nil
}

// 获取下个 setHour 点的 00:00:00 时间差
func GetDurationToNextTargetTime(targerHour int, loc *time.Location) (time.Time, time.Time, time.Duration) {
	now := time.Now().In(loc)
	target := time.Date(now.Year(), now.Month(), now.Day(), targerHour, 0, 0, 0, now.Location())
	if now.After(target) {
		target = target.Add(24 * time.Hour) // 如果当前时间已经过了 setHour 点，就设置为第二天的 setHour 点
	}
	// 计算时间差
	duration := target.Sub(now)
	return now, target, duration
}

// ############################   定时器    ################

type (
	Func   func()
	Ticker struct {
		fn       Func
		closeCh  chan bool
		closed   int32
		duration time.Duration // 基准间隔
		jitterMs int64         // 抖动范围（毫秒），0 表示不抖动
		timer    *time.Timer
	}
)

func pexec(fn Func) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%v\n", err)
			time.Sleep(1 * time.Second)
		}
	}()
	fn()
}

func (r *Ticker) interval() time.Duration {
	if r.jitterMs <= 0 {
		return r.duration
	}
	jitter := rand.Int63n(r.jitterMs*2+1) - r.jitterMs
	return r.duration + time.Duration(jitter)*time.Millisecond
}

func (r *Ticker) loop() {
	defer func() {
		r.timer.Stop()
		if err := recover(); err != nil {
			fmt.Printf("Recovered from panic: %v\n", err)
			time.Sleep(1 * time.Second)
		}
	}()

	r.timer = time.NewTimer(r.interval())
	for {
		select {
		case <-r.timer.C:
			if atomic.LoadInt32(&r.closed) == 0 {
				pexec(r.fn)
			}
			if atomic.LoadInt32(&r.closed) == 0 {
				if !r.timer.Stop() {
					select {
					case <-r.timer.C:
					default:
					}
				}
				r.timer.Reset(r.interval())
			} else {
				return // 退出 loop，避免错误的 Reset()
			}
		case <-r.closeCh:
			return
		}
	}
}

func (r *Ticker) Stop() {
	if atomic.CompareAndSwapInt32(&r.closed, 0, 1) {
		close(r.closeCh) // 关闭通道，通知 goroutine 退出
	}
}

// NewTimer 创建一个定时器，每隔 ms 毫秒调用 fn 函数。
func NewTimer(ms int64, fn Func) *Ticker {
	return newTicker(ms, 0, fn)
}

// NewTimerJitter 创建带随机抖动的定时器，实际间隔为 ms ± jitterMs，避免多进程同时刷新。
func NewTimerJitter(ms int64, jitterMs int64, fn Func) *Ticker {
	return newTicker(ms, jitterMs, fn)
}

func newTicker(ms int64, jitterMs int64, fn Func) *Ticker {
	ticker := &Ticker{
		fn:       fn,
		closed:   0,
		closeCh:  make(chan bool, 2),
		duration: time.Duration(ms) * time.Millisecond,
		jitterMs: jitterMs,
	}
	go ticker.loop()
	return ticker
}
