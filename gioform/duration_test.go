package gioform

import (
	"errors"
	"slices"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	const s, m, h = time.Second, time.Minute, time.Hour
	for in, want := range map[string]time.Duration{
		"90":      90 * m, // no unit: minutes
		"10s":     10 * s,
		"3m":      3 * m,
		"3h":      3 * h,
		"2:30":    2*h + 30*m,
		":45":     45 * m,
		"2:":      2 * h,
		"1m30s":   m + 30*s,
		"1m30":    m + 30*s, // trailing number: the next unit down
		"2h30":    2*h + 30*m,
		"2h30m":   2*h + 30*m,
		"1h0m0s":  h, // Go's own form
		" 1M 5S ": m + 5*s,
		"0":       0,
		"0s":      0,
	} {
		got, err := ParseDuration(in)
		if err != nil || got != want {
			t.Errorf("ParseDuration(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, bad := range []string{"", " ", ":", "abc", "1x", "m", "1s30m", "1m1m", "1s5", "1 30", "-5", "1.5m", "２m"} {
		if _, err := ParseDuration(bad); !errors.Is(err, ErrDurationFormat) {
			t.Errorf("ParseDuration(%q) error = %v, want ErrDurationFormat", bad, err)
		}
	}
	for _, huge := range []string{"99999999999999999999", "9999999999999h"} {
		if _, err := ParseDuration(huge); !errors.Is(err, ErrDurationTooLong) {
			t.Errorf("ParseDuration(%q) error = %v, want ErrDurationTooLong", huge, err)
		}
	}
}

func TestParseDurationIn(t *testing.T) {
	if d, err := ParseDurationIn("30s", 5*time.Second, time.Minute); err != nil || d != 30*time.Second {
		t.Errorf("in range: %v %v", d, err)
	}
	if _, err := ParseDurationIn("2m", 0, time.Minute); !errors.Is(err, ErrDurationTooLong) {
		t.Errorf("too long: %v", err)
	}
	if _, err := ParseDurationIn("1s", 5*time.Second, time.Minute); !errors.Is(err, ErrDurationTooShort) {
		t.Errorf("too short: %v", err)
	}
}

func TestFormatDuration(t *testing.T) {
	zh := DefaultLabels
	zh.DurationNow, zh.DurationHours, zh.DurationMinutes, zh.DurationMin, zh.DurationSeconds =
		"立即", "%d 小時", "%d 分鐘", "%d 分", "%d 秒"
	for _, c := range []struct {
		d      time.Duration
		en, zh string
	}{
		{0, "Now", "立即"},
		{400 * time.Millisecond, "Now", "立即"},
		{10 * time.Second, "10 s", "10 秒"},
		{5 * time.Minute, "5 min", "5 分鐘"},
		{90 * time.Second, "1 min 30 s", "1 分 30 秒"},
		{2*time.Hour + 30*time.Minute, "2 h 30 min", "2 小時 30 分"},
		{2 * time.Hour, "2 h", "2 小時"},
	} {
		if got := FormatDuration(c.d, DefaultLabels); got != c.en {
			t.Errorf("FormatDuration(%v) = %q, want %q", c.d, got, c.en)
		}
		if got := FormatDuration(c.d, zh); got != c.zh {
			t.Errorf("FormatDuration(%v, zh) = %q, want %q", c.d, got, c.zh)
		}
	}
}

func TestDurationStops(t *testing.T) {
	s := durationStops(5*time.Second, 10*time.Minute, nil)
	if s[0] != 5*time.Second || s[len(s)-1] != 10*time.Minute || !slices.IsSorted(s) {
		t.Fatalf("stops = %v", s)
	}
	if slices.Contains(s, 2*time.Second) || slices.Contains(s, time.Hour) {
		t.Fatalf("stops outside the range: %v", s)
	}
	// An odd minimum is a stop of its own.
	if s := durationStops(7*time.Second, time.Minute, nil); s[0] != 7*time.Second || s[1] != 10*time.Second {
		t.Fatalf("stops = %v", s)
	}
	custom := durationStops(0, time.Hour, []time.Duration{time.Minute, 0, time.Minute, 30 * time.Second})
	if !slices.Equal(custom, []time.Duration{0, 30 * time.Second, time.Minute}) {
		t.Fatalf("custom stops = %v", custom)
	}

	if got := stepDuration(s, 30*time.Second, 1); got != time.Minute {
		t.Errorf("step up = %v", got)
	}
	if got := stepDuration(s, 40*time.Second, -1); got != 30*time.Second {
		t.Errorf("step down from a typed value = %v", got)
	}
	if got := stepDuration(s, 10*time.Minute, 1); got != 10*time.Minute {
		t.Errorf("step past the end = %v", got)
	}
	if got := stepDuration(s, 5*time.Second, -1); got != 5*time.Second {
		t.Errorf("step before the start = %v", got)
	}
}

func TestDurationField(t *testing.T) {
	f := New(NewTheme(false), []Field{
		{Key: "clear", Label: "Clear clipboard after", Type: Duration, Default: "30s", Min: 5 * time.Second, Max: 10 * time.Minute},
		{Key: "lock", Label: "Lock after", Type: Duration, Default: "not a time", Min: time.Minute},
		{Key: "big", Label: "Big", Type: Duration, Default: "9h"},
	})
	if got := f.Duration("clear"); got != 30*time.Second {
		t.Errorf("default = %v", got)
	}
	if got := f.Duration("lock"); got != time.Minute {
		t.Errorf("invalid default should fall back to Min, got %v", got)
	}
	if got := f.Duration("big"); got != DefaultDurationMax {
		t.Errorf("default above Max should be clamped, got %v", got)
	}

	fs := f.field("clear")
	for _, c := range []struct {
		typed string
		want  time.Duration
		err   string
	}{
		{"2m", 2 * time.Minute, ""},
		{"1h", 2 * time.Minute, "At most 10 min"},
		{"1s", 2 * time.Minute, "At least 5 s"},
		{"soon", 2 * time.Minute, DefaultLabels.DurationFormat},
		{"  ", 2 * time.Minute, ""}, // blank: keep the value
	} {
		fs.editing = true
		fs.ed.SetText(c.typed)
		f.commitDuration(fs)
		if fs.dur != c.want || fs.err != c.err || fs.editing {
			t.Errorf("typed %q: dur %v err %q editing %v", c.typed, fs.dur, fs.err, fs.editing)
		}
	}

	f.SetText("clear", "1h") // clamped
	f.trySubmit()
	v, ok := f.Submitted()
	if !ok || v.Duration("clear") != 10*time.Minute || v.Text("clear") != "10m0s" {
		t.Fatalf("submitted %v %v %q", ok, v.Duration("clear"), v.Text("clear"))
	}
	f.Reset()
	if f.Duration("clear") != 30*time.Second {
		t.Errorf("Reset = %v", f.Duration("clear"))
	}
}
