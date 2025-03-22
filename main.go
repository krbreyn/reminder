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
			Prereq: DailyTimeHandler{
				Hour:   time.Now().Hour(),
				Minute: time.Now().Minute(),
			},
		},
	}

	new_weekly := Notification{
		Message: "This is my weekly notification!",
		Handler: WeeklyTimeHandler{
			Day: time.Now().Weekday(),
			Prereq: DailyTimeHandler{
				Hour:   time.Now().Hour(),
				Minute: time.Now().Minute(),
			},
		},
	}

	new_daily := Notification{
		Message: "This is my daily notification!",
		Handler: DailyTimeHandler{
			Hour:   time.Now().Hour(),
			Minute: time.Now().Minute(),
		},
	}

	ticker := time.NewTicker(time.Minute * 1)

	notifs := []Notification{new_daily, new_weekly, new_monthly}

	for {
		<-ticker.C
		now := time.Now()
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
	Hour   int
	Minute int
}

func (h DailyTimeHandler) Verify(t time.Time) bool {
	return t.Hour() >= h.Hour && t.Minute() >= h.Minute
}

type WeeklyTimeHandler struct {
	Day    time.Weekday
	Prereq DailyTimeHandler
}

func (h WeeklyTimeHandler) Verify(t time.Time) bool {
	return t.Weekday() == h.Day && h.Prereq.Verify(t)
}

type MonthlyTimeHandler struct {
	Day    int
	Prereq DailyTimeHandler
}

func (h MonthlyTimeHandler) Verify(t time.Time) bool {
	return t.Day() == h.Day && h.Prereq.Verify(t)
}
