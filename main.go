package main

import (
	"fmt"
	"github.com/caseymrm/menuet"
	"math/rand"

	"time"
)

var lastPos = 0
var lastTime = time.Now()

var clockIn = time.Now()

func tracker() {
	for {
		duration := time.Now().Sub(clockIn)
		menuet.App().SetMenuState(&menuet.MenuState{
			Title: fmtDuration(duration),
		})
		time.Sleep(time.Second)
	}
}
func clockInNow() {
	clockIn = time.Now()
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d:%02d", h, m)
}
func active() bool {
	pos := rand.Int()
	if pos != lastPos {
		lastPos = pos
		lastTime = time.Now()
		return true
	}
	if time.Now().Add(-15 * time.Minute).Before(lastTime) {
		return true
	}
	//x, y := robotgo.GetMousePos()
	//	fmt.Println("pos:", x, y)
	return false
}

func menuItems() []menuet.MenuItem {
	items := []menuet.MenuItem{
		{
			Text:    "Clock in",
			Clicked: clockInNow,
		},
		{
			Text: "Clock out",
		},
	}
	return items
}

func main() {
	go tracker()
	app := menuet.App()
	app.Children = menuItems
	app.RunApplication()
}
