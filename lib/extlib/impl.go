package extlib

import (
	"strings"
	"time"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib"
)

var Std = exp.Builtins(make(lib.Specs).AddMap(lib.Std).AddMap(MustLib(
	Str, Time, UUID,
)))

var Str = FuncMap{
	"index":    strings.Index,
	"last":     strings.LastIndex,
	"prefix":   strings.HasPrefix,
	"suffix":   strings.HasSuffix,
	"contains": strings.Contains,
	"upper":    strings.ToUpper,
	"lower":    strings.ToLower,
	"trim":     strings.TrimSpace,
	"like":     func(t, p string) bool { return Like(t, p, false) },
	"ilike":    func(t, p string) bool { return Like(t, p, true) },
}

var Time = FuncMap{
	"add_span":    time.Time.Add,
	"add_days":    AddDays,
	"add_date":    time.Time.AddDate,
	"sub_time":    time.Time.Sub,
	"year":        time.Time.Year,
	"month":       time.Time.Month,
	"weekday":     time.Time.Weekday,
	"yearday":     time.Time.YearDay,
	"day_start":   DayStart,
	"day_end":     DayEnd,
	"time_format": time.Time.Format,
	"fmt_date":    FmtDate,
	"fmt_time":    FmtTime,
	"fmt_human":   FmtTime,
}

var UUID = FuncMap{
	"new_uuid": cor.NewUUID,
}

func AddDays(t time.Time, days int) time.Time { return t.AddDate(0, 0, days) }
func DayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
func DayEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, -1, 0, t.Location())
}
func FmtDate(t time.Time) string  { return t.Format("2006-01-02") }
func FmtTime(t time.Time) string  { return t.Format("15:04:05") }
func FmtHuman(t time.Time) string { return t.Format("06-01-02 15:04") }
