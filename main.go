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
var auto = true

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
		time.Sleep(200 * time.Millisecond)
		if auto {
			if !active() {
				clockOutNow()
			} else {
				clockInNow()
			}
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

func clockInNowClicked() {
	auto = false
	clockInNow()
}

func clockOutNowClicked() {
	auto = false
	clockOutNow()
}
func clockOutNow() {
	if len(times) != 0 && times[len(times)-1].ClockOut.IsZero() {
		times[len(times)-1].ClockOut = time.Now()
	}
}

func toggleAuto() {
	auto = !auto
}

func reset() {
	times = times[:0]
}

func add15() {
	times = append(times, TimeStruct{
		ClockIn:  time.Now(),
		ClockOut: time.Now().Add(15 * time.Minute),
	})
}
func add30() {
	times = append(times, TimeStruct{
		ClockIn:  time.Now(),
		ClockOut: time.Now().Add(30 * time.Minute),
	})
}
func sub15() {
	times = append(times, TimeStruct{
		ClockIn:  time.Now(),
		ClockOut: time.Now().Add(-15 * time.Minute),
	})
}
func sub30() {
	times = append(times, TimeStruct{
		ClockIn:  time.Now(),
		ClockOut: time.Now().Add(-30 * time.Minute),
	})
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
			Text:    "Reset",
			Clicked: reset,
		},
		{
			Text:    "Add 15",
			Clicked: add15,
		},
		{
			Text:    "Remove 15",
			Clicked: sub15,
		},
		{
			Text:    "Add 30",
			Clicked: add30,
		},
		{
			Text:    "Remove 30",
			Clicked: sub30,
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
