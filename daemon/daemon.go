package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/krbreyn/reminder"
)

const app_name string = "go_reminder"

var xdg_data_home string = os.Getenv("XDG_DATA_HOME")
var xdg_runtime_dir string = os.Getenv("XDG_RUNTIME_DIR")

var NotifsGlobal []reminder.Notification

func RunDaemon() {
	setDataDir()
	log.Println(data_file)

	loadData(true)
	log.Println("loaded data")

	/* acquire program lock */
	if xdg_runtime_dir == "" {
		log.Fatal("xdg_runtime_dir not set")
	}

	path := filepath.Join(xdg_runtime_dir, fmt.Sprintf("%s.lock", app_name))
	lockFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("cannot create lock file: %v", err)
	}
	defer lockFile.Close()

	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		log.Fatal("another instance is already running")
	}
	/* acquire program lock */

	log.Println("lock acquired")

	ticker := time.NewTicker(time.Second * 5)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("entering loop")

	for {
		select {
		case <-ticker.C:
			log.Println("ticker")
			loadData(false)
			now := time.Now()
			for _, n := range NotifsGlobal {
				if n.Handler.Verify(now) {
					err := Notify(n.Message)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		case <-sigChan:
			saveData()
			os.Exit(0)
		}
	}
}

var data_file string

func setDataDir() {
	var data_dir string
	if xdg_data_home != "" {
		data_dir = xdg_data_home
	} else {
		data_dir = filepath.Join(os.Getenv("HOME"), ".local/share")
	}

	data_file = filepath.Join(
		data_dir,
		fmt.Sprintf("%s.json", app_name),
	)
}

func OpenDataFile() *os.File {
	dataFile, err := os.OpenFile(data_file, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return dataFile
}

var lastModificationDate time.Time

func ifShouldUpdateNotifs() bool {
	stat, err := os.Stat(data_file)
	if err != nil {
		log.Fatal(err)
	}

	modTime := stat.ModTime()
	if modTime == lastModificationDate {
		return false
	}
	lastModificationDate = modTime
	return true
}

func loadData(startup bool) {
	if !startup {
		should := ifShouldUpdateNotifs()
		if !should {
			return
		}
	}

	log.Println("updating notifs")

	dataFile := OpenDataFile()
	defer dataFile.Close()

	bytes, err := io.ReadAll(dataFile)
	if err != nil {
		log.Fatal(err)
	}

	if len(bytes) == 0 {
		return
	}

	data := []reminder.NotifData{}

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		log.Fatal(err)
	}

	new_notifs := []reminder.Notification{}

	for _, n := range data {
		notif := reminder.Notification{}
		notif.Message = n.Message

		prereq := &reminder.DailyTimeHandler{
			Hour:         n.Hour,
			Minute:       n.Minute,
			WasDoneToday: n.WasDoneToday,
			WasDoneAt:    n.WasDoneAt,
		}

		switch n.N_type {
		case reminder.DailyNotif:
			notif.Handler = prereq
		case reminder.WeeklyNotif:
			notif.Handler = &reminder.WeeklyTimeHandler{
				Weekday: n.Weekday,
				Prereq:  prereq,
			}
		case reminder.MonthlyNotif:
			notif.Handler = &reminder.MonthlyTimeHandler{
				Day:    n.Day,
				Prereq: prereq,
			}
		}

		new_notifs = append(new_notifs, notif)
	}

	NotifsGlobal = new_notifs
}

func saveData() {
	data := []reminder.NotifData{}

	for _, n := range NotifsGlobal {
		notif := reminder.NotifData{}
		notif.Message = n.Message

		switch n := n.Handler.(type) {
		case *reminder.DailyTimeHandler:
			notif.N_type = reminder.DailyNotif
			notif.Hour = n.Hour
			notif.Minute = n.Minute
			notif.WasDoneToday = n.WasDoneToday
			notif.WasDoneAt = n.WasDoneAt
		case *reminder.WeeklyTimeHandler:
			notif.N_type = reminder.WeeklyNotif
			notif.Hour = n.Prereq.Hour
			notif.Minute = n.Prereq.Minute
			notif.Weekday = n.Weekday
			notif.WasDoneToday = n.Prereq.WasDoneToday
			notif.WasDoneAt = n.Prereq.WasDoneAt
		case *reminder.MonthlyTimeHandler:
			notif.N_type = reminder.MonthlyNotif
			notif.Hour = n.Prereq.Hour
			notif.Minute = n.Prereq.Minute
			notif.Day = n.Day
			notif.WasDoneToday = n.Prereq.WasDoneToday
			notif.WasDoneAt = n.Prereq.WasDoneAt
		}

		data = append(data, notif)
	}

	dataFile := OpenDataFile()
	defer dataFile.Close()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dataFile.Write(jsonData)
	if err != nil {
		log.Fatal(err)
	}
}

// add debugging info stuff
func LogError(err error) {

}

func Notify(msg string) error {
	send, err := exec.LookPath("kdialog")
	if err != nil {
		return err
	}

	c := exec.Command(send, "--title", "Reminder!", "--msgbox", msg)
	return c.Run()
}
