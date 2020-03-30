package main

import (
	"encoding/json"
	"fmt"
	"github.com/caseymrm/menuet"
	"github.com/jantb/robotgo"
	"io/ioutil"
	"log"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"time"
)

const sevenAndHalfHour = 450 * time.Minute

var lastPos = 0
var lastTime = time.Now()
var auto = true

var days []Day

type Day struct {
	date  string
	times []TimeStruct
}

var autoTimeTresh = -15 * time.Minute
var subTimeTresh = true
var endOfDayNotice = false

type TimeStruct struct {
	ClockIn  time.Time
	ClockOut time.Time
}

var times []TimeStruct

func tracker() {
	for {
		duration := hoursForToday()
		checkEndOfDayAndDisplayMessage(duration)
		if canClockOut() {
			menuet.App().SetMenuState(&menuet.MenuState{
				Title: fmtDuration(duration),
				Image: "clock.pdf",
			})
		} else {
			menuet.App().SetMenuState(&menuet.MenuState{
				Title: fmtDuration(duration),
			})
		}

		time.Sleep(200 * time.Millisecond)
		if auto {
			if !active() {
				if subTimeTresh && canClockOut() {
					clockOutNow()
					subAutoTresh()
					continue
				}
				clockOutNow()
			} else {
				clockInNow()
			}
		}
	}
}

func checkEndOfDayAndDisplayMessage(duration time.Duration) {

	if duration > sevenAndHalfHour && !endOfDayNotice {
		endOfDayNotice = true
		alert := menuet.App().Alert(menuet.Alert{
			MessageText:     "Worked 7.5 hours today",
			InformativeText: "Time to register hours",
			Buttons:         []string{"Ok", "Open AlvTime"},
			Inputs:          nil,
		})
		if alert.Button == 1 {
			openAlvTime()
		}
	}
}

func hoursForToday() time.Duration {
	duration := int64(0)
	for _, timeStruct := range times {
		if timeStruct.ClockOut.IsZero() {
			duration += time.Now().Sub(timeStruct.ClockIn).Nanoseconds()
			continue
		}
		duration += timeStruct.ClockOut.Sub(timeStruct.ClockIn).Nanoseconds()
	}
	return time.Duration(duration)
}

func clockInNow() {
	if len(times) == 0 || !times[len(times)-1].ClockOut.IsZero() {
		times = append(times, TimeStruct{
			ClockIn: time.Now(),
		})
		store(days)
	}
}

func clockInNowClicked() {
	auto = false
	clockInNow()
}

func clockOutNowClicked() {
	auto = false
	clockOutNow()
}
func clockOutNow() {
	if canClockOut() {
		times[len(times)-1].ClockOut = time.Now()
		store(days)
	}
}

func canClockOut() bool {
	return len(times) != 0 && times[len(times)-1].ClockOut.IsZero()
}

func toggleAuto() {
	auto = !auto
}

func toggleSubAutotresh() {
	subTimeTresh = !subTimeTresh
}

func reset() {
	times = times[:0]
	endOfDayNotice = false
}

func addDuration(dur time.Duration) {
	clockOutNow()
	times = append(times, TimeStruct{
		ClockIn:  time.Now(),
		ClockOut: time.Now().Add(dur),
	})
}
func subAutoTresh() {
	clockOutNow()
	times = append(times, TimeStruct{
		ClockIn:  time.Now(),
		ClockOut: time.Now().Add(autoTimeTresh),
	})
}

func openAlvTime() {
	duration := hoursForToday()
	d := duration.Round(15 * time.Minute)
	err := robotgo.WriteAll(fmt.Sprintf("%.2f", d.Hours()))
	if err != nil {
		log.Fatal(err)
	}
	openbrowser("https://alvtime-vue-pwa-prod.azurewebsites.net/")
}
func openExperis() {
	duration := hoursForToday()
	d := duration.Round(15 * time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	err := robotgo.WriteAll(fmt.Sprintf("%d:%d", h, m))
	if err != nil {
		log.Fatal(err)
	}
	openbrowser("https://mytime.experis.no//")
}
func fmtDuration(dur time.Duration) string {
	d := dur.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	doneBy := time.Now().Add(sevenAndHalfHour).Add(-dur)

	if canClockOut() {
		return fmt.Sprintf("%02d:%02d %s", h, m, doneBy.Format("15:04"))
	}

	return fmt.Sprintf("%02d:%02d", h, m)
}

func active() bool {
	x, y := robotgo.GetMousePos()
	pos := x + y
	if pos != lastPos {
		lastPos = pos
		lastTime = time.Now()
		return true
	}
	if time.Now().Add(autoTimeTresh).Before(lastTime) {
		return true
	}
	return false
}

func menuItems() []menuet.MenuItem {
	items := []menuet.MenuItem{
		{
			Text:    "Clock in",
			Clicked: clockInNowClicked,
			State:   len(times) != 0 && times[len(times)-1].ClockOut.IsZero(),
		},
		{
			Text:    "Clock out",
			Clicked: clockOutNowClicked,
			State:   len(times) == 0 || !times[len(times)-1].ClockOut.IsZero(),
		},
		{
			Text:    "Auto",
			Clicked: toggleAuto,
			State:   auto,
		},
		{
			Type: menuet.Separator,
		},
		{
			Text: "Adjust AutoTime",
			Children: func() []menuet.MenuItem {
				return []menuet.MenuItem{
					{
						Text: "5m",
						Clicked: func() {
							autoTimeTresh = -5 * time.Minute
						},
						State: autoTimeTresh == -5*time.Minute,
					},
					{
						Text: "10m",
						Clicked: func() {
							autoTimeTresh = -10 * time.Minute
						},
						State: autoTimeTresh == -10*time.Minute,
					},
					{
						Text: "15m",
						Clicked: func() {
							autoTimeTresh = -15 * time.Minute
						},
						State: autoTimeTresh == -15*time.Minute,
					},
					{
						Text: "30m",
						Clicked: func() {
							autoTimeTresh = -30 * time.Minute
						},
						State: autoTimeTresh == -30*time.Minute,
					},
					{
						Text: "1h",
						Clicked: func() {
							autoTimeTresh = -60 * time.Minute
						},
						State: autoTimeTresh == -60*time.Minute,
					},
					{
						Text:    "Subtract time after idle",
						Clicked: toggleSubAutotresh,
						State:   subTimeTresh,
					},
				}
			},
		},
		{
			Text: "Adjust time",
			Children: func() []menuet.MenuItem {
				return []menuet.MenuItem{
					{
						Text:    "Reset",
						Clicked: reset,
					},
					{
						Text: "Add 15",
						Clicked: func() {
							addDuration(15 * time.Minute)
						},
					},
					{
						Text: "Add 30",
						Clicked: func() {
							addDuration(30 * time.Minute)
						},
					},
					{
						Text: "Remove 15",
						Clicked: func() {
							addDuration(-15 * time.Minute)
						},
					},
					{
						Text: "Remove 30",
						Clicked: func() {
							addDuration(-30 * time.Minute)
						},
					},
				}
			},
		},
		{
			Type: menuet.Separator,
		},
		{
			Text:    "AlvTime",
			Clicked: openAlvTime,
		},
		{
			Text:    "Experis",
			Clicked: openExperis,
		},
	}
	return items
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func store(days []Day) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := json.MarshalIndent(days, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(filepath.Join(usr.HomeDir, ".days.json"), bytes, 0600)
}

func load() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadFile(filepath.Join(usr.HomeDir, ".jira.conf"))
	if err != nil {
		bytes, err := json.MarshalIndent(days, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(filepath.Join(usr.HomeDir, ".jira.conf"), bytes, 0600)
	}

	json.Unmarshal(bytes, &days)
}

func main() {
	load()
	go tracker()
	app := menuet.App()
	app.Children = menuItems
	app.Label = "AlvMenuTime"
	app.AutoUpdate.Version = "v0.3"
	app.AutoUpdate.Repo = "jantb/time"
	app.RunApplication()
}
