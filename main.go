package main

import (
	"fmt"
	"github.com/caseymrm/menuet"
	"github.com/jantb/robotgo"
	"log"
	"os/exec"
	"runtime"
	"time"
)

const sevenAndHalfHour = 450 * time.Minute

var lastPos = 0
var lastTime = time.Now()
var auto = true

var tracking Tracking

var autoTimeTresh = -15 * time.Minute
var subTimeTresh = true
var endOfDayNotice = false

func tracker() {
	tracking.reset()
	for {
		hoursForToday := tracking.hoursForToday()
		tracking.updateTotalForToday(hoursForToday)
		checkEndOfDayAndDisplayMessage(hoursForToday)
		if tracking.canClockOut() {
			menuet.App().SetMenuState(&menuet.MenuState{
				Title: fmtDuration(hoursForToday),
				Image: "clock.pdf",
			})
		} else {
			menuet.App().SetMenuState(&menuet.MenuState{
				Title: fmtDuration(hoursForToday),
			})
		}

		time.Sleep(200 * time.Millisecond)
		tracking.store()
		if auto {
			if !active() {
				if subTimeTresh && tracking.canClockOut() {
					tracking.clockOutNow()
					tracking.subAutoTresh(autoTimeTresh)
					continue
				}
				tracking.clockOutNow()
			} else {
				tracking.clockInNow()
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

func clockInNowClicked() {
	auto = false
	tracking.clockInNow()
}

func clockOutNowClicked() {
	auto = false
	tracking.clockOutNow()
}

func toggleAuto() {
	auto = !auto
}

func toggleSubAutotresh() {
	subTimeTresh = !subTimeTresh
}

func openAlvTime() {
	duration := tracking.hoursForToday()
	d := duration.Round(15 * time.Minute)
	err := robotgo.WriteAll(fmt.Sprintf("%.2f", d.Hours()))
	if err != nil {
		log.Fatal(err)
	}
	openbrowser("https://alvtime-vue-pwa-prod.azurewebsites.net/")
}
func openExperis() {
	duration := tracking.hoursForToday()
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

	if tracking.canClockOut() {
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
			State:   len(tracking.getTimes()) != 0 && tracking.getTimes()[len(tracking.getTimes())-1].ClockOut.IsZero(),
		},
		{
			Text:    "Clock out",
			Clicked: clockOutNowClicked,
			State:   len(tracking.getTimes()) == 0 || !tracking.getTimes()[len(tracking.getTimes())-1].ClockOut.IsZero(),
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
						Text: "Reset",
						Clicked: func() {
							tracking.reset()
							endOfDayNotice = false
						},
					},
					{
						Text: "Add 15",
						Clicked: func() {
							tracking.addDuration(15 * time.Minute)
						},
					},
					{
						Text: "Add 30",
						Clicked: func() {
							tracking.addDuration(30 * time.Minute)
						},
					},
					{
						Text: "Remove 15",
						Clicked: func() {
							tracking.addDuration(-15 * time.Minute)
						},
					},
					{
						Text: "Remove 30",
						Clicked: func() {
							tracking.addDuration(-30 * time.Minute)
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

func main() {
	tracking.load()
	go tracker()
	app := menuet.App()
	app.Children = menuItems
	app.Label = "AlvMenuTime"
	app.AutoUpdate.Version = "v0.3"
	app.AutoUpdate.Repo = "jantb/time"
	app.RunApplication()
}
