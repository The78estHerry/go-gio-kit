package gioform

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Errors returned by ParseDuration and ParseDurationIn.
var (
	ErrDurationFormat   = errors.New("gioform: not a duration")
	ErrDurationTooLong  = errors.New("gioform: duration too long")
	ErrDurationTooShort = errors.New("gioform: duration too short")
)

// DefaultDurationMax is the longest time a Duration field accepts when
// Field.Max is zero.
const DefaultDurationMax = 4 * time.Hour

// ParseDuration reads a time the way people type it, in whole seconds:
//
//	90      90 minutes (a number without a unit is minutes)
//	10s     10 seconds
//	3m      3 minutes
//	3h      3 hours
//	2:30    2 hours 30 minutes (h:mm)
//	1m30s   1 minute 30 seconds
//	2h30    2 hours 30 minutes (a trailing number takes the next unit)
//
// Units go from large to small and may be separated by spaces; case does
// not matter. Go's own form ("1h0m0s") is accepted too.
func ParseDuration(text string) (time.Duration, error) {
	t := strings.ToLower(strings.TrimSpace(text))
	if t == "" {
		return 0, ErrDurationFormat
	}
	const limit = int64(math.MaxInt64 / int64(time.Second))
	add := func(total, n, unit int64) (int64, error) {
		if n > (limit-total)/unit {
			return 0, ErrDurationTooLong
		}
		return total + n*unit, nil
	}
	number := func(d string) (int64, error) {
		if d == "" {
			return 0, nil
		}
		for _, c := range d {
			if c < '0' || c > '9' {
				return 0, ErrDurationFormat
			}
		}
		n, err := strconv.ParseInt(d, 10, 64)
		if err != nil {
			return 0, ErrDurationTooLong
		}
		return n, nil
	}

	var total int64
	if h, m, ok := strings.Cut(t, ":"); ok {
		h, m = strings.TrimSpace(h), strings.TrimSpace(m)
		if h == "" && m == "" {
			return 0, ErrDurationFormat
		}
		hn, err := number(h)
		if err != nil {
			return 0, err
		}
		mn, err := number(m)
		if err != nil {
			return 0, err
		}
		if total, err = add(0, hn, 3600); err != nil {
			return 0, err
		}
		if total, err = add(total, mn, 60); err != nil {
			return 0, err
		}
		return time.Duration(total) * time.Second, nil
	}

	var digits string
	var last int64 // the last unit used, in seconds
	for _, c := range t {
		var unit int64
		switch {
		case c >= '0' && c <= '9':
			digits += string(c)
			continue
		case c == ' ' || c == '\t':
			if digits != "" {
				return 0, ErrDurationFormat // "1 30" is ambiguous
			}
			continue
		case c == 'h':
			unit = 3600
		case c == 'm':
			unit = 60
		case c == 's':
			unit = 1
		default:
			return 0, ErrDurationFormat
		}
		if digits == "" || (last != 0 && unit >= last) {
			return 0, ErrDurationFormat
		}
		n, err := number(digits)
		if err != nil {
			return 0, err
		}
		if total, err = add(total, n, unit); err != nil {
			return 0, err
		}
		digits, last = "", unit
	}
	if digits != "" {
		var unit int64
		switch last {
		case 0, 3600:
			unit = 60
		case 60:
			unit = 1
		default:
			return 0, ErrDurationFormat
		}
		n, err := number(digits)
		if err != nil {
			return 0, err
		}
		if total, err = add(total, n, unit); err != nil {
			return 0, err
		}
	}
	return time.Duration(total) * time.Second, nil
}

// ParseDurationIn is ParseDuration with a range check.
func ParseDurationIn(text string, min, max time.Duration) (time.Duration, error) {
	d, err := ParseDuration(text)
	switch {
	case err != nil:
		return 0, err
	case d > max:
		return 0, ErrDurationTooLong
	case d < min:
		return 0, ErrDurationTooShort
	}
	return d, nil
}

// FormatDuration writes d for people with the words in l, in whole
// seconds: "Now", "10 s", "5 min", "1 min 30 s", "2 h 30 min".
func FormatDuration(d time.Duration, l Labels) string {
	total := int64((d + time.Second/2) / time.Second)
	if total <= 0 {
		return l.DurationNow
	}
	h, m, s := total/3600, total%3600/60, total%60
	var parts []string
	if h > 0 {
		parts = append(parts, fmt.Sprintf(l.DurationHours, h))
	}
	if m > 0 {
		f := l.DurationMin
		if h == 0 && s == 0 {
			f = l.DurationMinutes
		}
		parts = append(parts, fmt.Sprintf(f, m))
	}
	if s > 0 {
		parts = append(parts, fmt.Sprintf(l.DurationSeconds, s))
	}
	return strings.Join(parts, " ")
}

// commonDurations are the stops of the ◀ ▶ arrows when a field sets no
// Steps.
var commonDurations = []time.Duration{
	0, time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second, 15 * time.Second, 30 * time.Second,
	time.Minute, 2 * time.Minute, 3 * time.Minute, 5 * time.Minute, 10 * time.Minute, 15 * time.Minute,
	20 * time.Minute, 30 * time.Minute, 45 * time.Minute,
	time.Hour, 2 * time.Hour, 3 * time.Hour, 4 * time.Hour, 6 * time.Hour, 8 * time.Hour, 12 * time.Hour, 24 * time.Hour,
}

// durationStops returns the sorted values the arrows step through: steps
// if given, otherwise min, the common times between min and max, and max.
func durationStops(min, max time.Duration, steps []time.Duration) []time.Duration {
	if len(steps) > 0 {
		s := slices.Clone(steps)
		slices.Sort(s)
		return slices.Compact(s)
	}
	s := []time.Duration{min}
	for _, v := range commonDurations {
		if v > min && v < max {
			s = append(s, v)
		}
	}
	if max > min {
		s = append(s, max)
	}
	return s
}

// stepDuration returns the next stop after v (dir > 0) or before it
// (dir < 0), or v itself when there is none.
func stepDuration(stops []time.Duration, v time.Duration, dir int) time.Duration {
	if dir > 0 {
		for _, s := range stops {
			if s > v {
				return s
			}
		}
		return v
	}
	for i := len(stops) - 1; i >= 0; i-- {
		if stops[i] < v {
			return stops[i]
		}
	}
	return v
}

// clampDuration rounds d to whole seconds and keeps it in [min, max].
func clampDuration(d, lo, hi time.Duration) time.Duration {
	return max(lo, min(d.Round(time.Second), hi))
}
