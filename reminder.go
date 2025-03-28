package reminder

import "time"

type NotifType int

const (
	DailyNotif NotifType = iota + 1
	WeeklyNotif
	MonthlyNotif
)

type NotifData struct {
	Message      string       `json:"message"`
	N_type       NotifType    `json:"n_type"`
	Hour         int          `json:"hour"`
	Minute       int          `json:"minute"`
	Weekday      time.Weekday `json:"weekday"`
	Day          int          `json:"day"`
	WasDoneToday bool         `json:"wasDoneToday"`
	WasDoneAt    time.Time    `json:"wasDoneAt"`
}

type Notification struct {
	Message string
	Handler TimeHandler
}

type TimeHandler interface {
	Verify(time.Time) bool
}

type DailyTimeHandler struct {
	Hour         int
	Minute       int
	WasDoneToday bool
	WasDoneAt    time.Time
}

func (h *DailyTimeHandler) Verify(t time.Time) bool {
	if h.WasDoneToday && t.Day() != h.WasDoneAt.Day() {
		h.WasDoneToday = false
	}

	if !h.WasDoneToday && t.Hour() >= h.Hour && t.Minute() >= h.Minute {
		h.WasDoneToday = true
		h.WasDoneAt = t
		return true
	}

	return false
}

type WeeklyTimeHandler struct {
	Weekday time.Weekday
	Prereq  *DailyTimeHandler
}

func (h *WeeklyTimeHandler) Verify(t time.Time) bool {
	return h.Prereq.Verify(t) && t.Weekday() == h.Weekday
}

type MonthlyTimeHandler struct {
	Day    int
	Prereq *DailyTimeHandler
}

func (h *MonthlyTimeHandler) Verify(t time.Time) bool {
	return h.Prereq.Verify(t) && t.Day() == h.Day
}
