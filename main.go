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

var lastPos = 0
var lastTime = time.Now()

type TimeStruct struct {
	ClockIn  time.Time
	ClockOut time.Time
}

var times []TimeStruct

func tracker() {
	for {
		duration := hoursForToday()
		menuet.App().SetMenuState(&menuet.MenuState{
			Title: fmtDuration(duration),
		})
		time.Sleep(time.Second)
		if !active() {
			clockOutNow()
		} else {
			clockInNow()
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
	}
}

func clockOutNow() {
	if len(times) != 0 && times[len(times)-1].ClockOut.IsZero() {
		times[len(times)-1].ClockOut = time.Now()
	}
}

func openAlvTime() {
	openbrowser("https://alvtime-vue-pwa-prod.azurewebsites.net/")
}

func openExperis() {
	openbrowser("https://mytime.experis.no//")
}
func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
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
	if time.Now().Add(-15 * time.Minute).Before(lastTime) {
		return true
	}
	return false
}

func menuItems() []menuet.MenuItem {
	items := []menuet.MenuItem{
		{
			Text:    "Clocked in",
			Clicked: clockInNow,
			State:   len(times) != 0 && times[len(times)-1].ClockOut.IsZero(),
		},
		{
			Text:    "Clocked out",
			Clicked: clockOutNow,
			State:   len(times) == 0 || !times[len(times)-1].ClockOut.IsZero(),
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
	go tracker()
	app := menuet.App()
	app.Children = menuItems
	app.RunApplication()
}
