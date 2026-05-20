// Package timehelper
// функции для работы со временем
// стандартный часовой пояс везде +5
package timehelper

import (
	"time"
)

const (
	defaultTimeOffset = 5 * 60 * 60 // Часы/Минуты/Секунды
)

func Now() time.Time {
	tz := time.FixedZone("timezone", defaultTimeOffset)

	return time.Now().In(tz)
}

// GetDateInServerTZ договорились, что время в сервисе всегда будет в часовом поясе екб.
func GetDateInServerTZ(t time.Time) time.Time {
	tz := time.FixedZone("timezone", defaultTimeOffset)
	y, m, d := t.Date()
	t = time.Date(y, m, d, 0, 0, 0, 0, tz)

	return t
}

func CorrectTimezone(t time.Time) time.Time {
	tz := time.FixedZone("timezone", defaultTimeOffset)

	return t.In(tz)
}

// ClockDuration переводит строку времени в формате HH:MM в duration, т.е. "11:00" превратится в 11h.
func ClockDuration(clock string) (time.Duration, error) {
	c, err := time.Parse("15:04", clock)
	if err != nil {
		return 0, err
	}

	h, m, _ := c.Clock()
	d := time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute

	return d, nil
}

// Dow - День недели с 0=ПН, 6=ВСКР.
func Dow(t time.Time) (dow time.Weekday) {
	dow = t.Weekday() - 1
	if dow < 0 {
		dow = 6
	}

	return dow
}
