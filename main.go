package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gen2brain/beeep"
)

func main() {
	err := beeep.Alert("test notif", "test notif desc", "")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	new_monthly := Notification{
		Message: "This is my monthly notification!",
		Handler: MonthlyTimeHandler{
			Day: 22,
			Prereq: &DailyTimeHandler{
				Hour:         time.Now().Hour(),
				Minute:       time.Now().Minute(),
				wasDoneToday: false,
				wasDoneAt:    time.Time{},
			},
		},
	}

	new_weekly := Notification{
		Message: "This is my weekly notification!",
		Handler: WeeklyTimeHandler{
			Day: time.Now().Weekday(),
			Prereq: &DailyTimeHandler{
				Hour:         time.Now().Hour(),
				Minute:       time.Now().Minute(),
				wasDoneToday: false,
				wasDoneAt:    time.Time{},
			},
		},
	}

	new_daily := Notification{
		Message: "This is my daily notification!",
		Handler: &DailyTimeHandler{
			Hour:         time.Now().Hour(),
			Minute:       time.Now().Minute(),
			wasDoneToday: false,
			wasDoneAt:    time.Time{},
		},
	}

	ticker := time.NewTicker(time.Second * 5)

	notifs := []Notification{new_daily, new_weekly, new_monthly}

	now := time.Now()
	for {
		<-ticker.C

		now = now.Add(time.Minute * 2)
		for _, n := range notifs {
			if n.Handler.Verify(now) {
				err := beeep.Notify("Reminder!", n.Message, "")

				if err != nil {
					panic(err)
				}
			}
		}
	}
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
	wasDoneToday bool
	wasDoneAt    time.Time
}

func (h *DailyTimeHandler) Verify(t time.Time) bool {
	if h.wasDoneToday && t.Day() != h.wasDoneAt.Day() {
		h.wasDoneToday = false
	}

	if !h.wasDoneToday && t.Hour() >= h.Hour && t.Minute() >= h.Minute {
		h.wasDoneToday = true
		h.wasDoneAt = t
		return true
	}

	return false
}

type WeeklyTimeHandler struct {
	Day    time.Weekday
	Prereq *DailyTimeHandler
}

func (h WeeklyTimeHandler) Verify(t time.Time) bool {
	return t.Weekday() == h.Day && h.Prereq.Verify(t)
}

type MonthlyTimeHandler struct {
	Day    int
	Prereq *DailyTimeHandler
}

func (h MonthlyTimeHandler) Verify(t time.Time) bool {
	return t.Day() == h.Day && h.Prereq.Verify(t)
}
