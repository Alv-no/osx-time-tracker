package main

import (
	"fmt"
	"github.com/caseymrm/menuet"
	"github.com/jantb/robotgo"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const sevenAndHalfHour = 27000000000000

var lastPos = 0
var lastTime = time.Now()
var auto = true
var week = Week{}

type Week struct {
	Monday    int
	Tuesday   int
	Wednesday int
	Thursday  int
	Friday    int
	Saturday  int
	Sunday    int
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

		menuet.App().SetMenuState(&menuet.MenuState{
			Title: fmtDuration(duration),
		})
		time.Sleep(200 * time.Millisecond)
		if auto {
			if !active() {
				if subTimeTresh && canClockOut() {
					subAutoTresh()
				}
				clockOutNow()
			} else {
				clockInNow()
			}
		}
	}
}

func checkEndOfDayAndDisplayMessage(duration time.Duration) {

	if duration.Nanoseconds() > sevenAndHalfHour && !endOfDayNotice {
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

func subAutoTresh() {
	times = append(times, TimeStruct{
		ClockIn:  time.Now(),
		ClockOut: time.Now().Add(autoTimeTresh * time.Minute),
	})
}

func openAlvTime() {
	openbrowser("https://alvtime-vue-pwa-prod.azurewebsites.net/")
}

func copyAlvTime() {
	robotgo.KeyTap("a", "command")
	robotgo.KeyTap("c", "command")
	time.Sleep(100 * time.Millisecond)
	field, err := robotgo.ReadAll()
	if err != nil {
		return
	}
	week = Week{
		Monday:    parseDecimalStringHoursToMinutes(field),
		Tuesday:   readNextField(),
		Wednesday: readNextField(),
		Thursday:  readNextField(),
		Friday:    readNextField(),
		Saturday:  readNextField(),
		Sunday:    readNextField(),
	}
}

func pasteExperis() {
	robotgo.PasteStr(fmt.Sprintf("%d:%d", week.Monday/60, week.Monday%60))
	time.Sleep(1000 * time.Second)
	robotgo.KeyTap("tab")
	time.Sleep(1000 * time.Second)
	robotgo.PasteStr(fmt.Sprintf("%d:%d", week.Tuesday/60, week.Tuesday%60))
	time.Sleep(1000 * time.Second)
	robotgo.KeyTap("tab")
	time.Sleep(1000 * time.Second)
	robotgo.PasteStr(fmt.Sprintf("%d:%d", week.Wednesday/60, week.Wednesday%60))
	time.Sleep(1000 * time.Second)
	robotgo.KeyTap("tab")
	time.Sleep(1000 * time.Second)
	robotgo.PasteStr(fmt.Sprintf("%d:%d", week.Thursday/60, week.Thursday%60))
	time.Sleep(1000 * time.Second)
	robotgo.KeyTap("tab")
	time.Sleep(1000 * time.Second)
	robotgo.PasteStr(fmt.Sprintf("%d:%d", week.Friday/60, week.Friday%60))
	time.Sleep(1000 * time.Second)
	robotgo.KeyTap("tab")
	time.Sleep(1000 * time.Second)
	robotgo.PasteStr(fmt.Sprintf("%d:%d", week.Saturday/60, week.Saturday%60))
	time.Sleep(1000 * time.Second)
	robotgo.KeyTap("tab")
	time.Sleep(1000 * time.Second)
	robotgo.PasteStr(fmt.Sprintf("%d:%d", week.Sunday/60, week.Sunday%60))
}

func readNextField() int {
	robotgo.KeyTap("tab")
	robotgo.KeyTap("c", "command")
	time.Sleep(100 * time.Millisecond)
	field, err := robotgo.ReadAll()
	if err != nil {
		return 0
	}

	return parseDecimalStringHoursToMinutes(field)
}

func parseDecimalStringHoursToMinutes(field string) int {
	contains := strings.Contains(field, ",")
	if contains {
		split := strings.Split(field, ",")
		if len(split) == 2 {
			hours, err := strconv.Atoi(split[0])
			if err != nil {
				return 0
			}
			minutes, err := strconv.Atoi(split[1])
			if err != nil {
				return 0
			}

			return minutes + (hours * 60)
		} else {
			return 0
		}
	}
	hours, err := strconv.Atoi(field)
	if err != nil {
		return 0
	}
	return hours * 60
}

func openExperis() {
	openbrowser("https://mytime.experis.no//")
}
func fmtDuration(dur time.Duration) string {
	d := dur.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	duration := sevenAndHalfHour * time.Nanosecond

	doneBy := time.Now().Add(duration).Add(-dur)

	if canClockOut() {
		return fmt.Sprintf("􀐱 %02d:%02d %s", h, m, doneBy.Format("15:04"))
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
	if time.Now().Add(autoTimeTresh * time.Minute).Before(lastTime) {
		return true
	}
	return false
}

func autoTres5() {
	autoTimeTresh = -5 * time.Minute
}
func autoTres10() {
	autoTimeTresh = -10 * time.Minute
}
func autoTres15() {
	autoTimeTresh = -15 * time.Minute
}
func autoTres30() {
	autoTimeTresh = -30 * time.Minute
}
func autoTres60() {
	autoTimeTresh = -60 * time.Minute
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
						Text:    "5m",
						Clicked: autoTres5,
						State:   autoTimeTresh == -5*time.Minute,
					},
					{
						Text:    "10m",
						Clicked: autoTres10,
						State:   autoTimeTresh == -10*time.Minute,
					},
					{
						Text:    "15m",
						Clicked: autoTres15,
						State:   autoTimeTresh == -15*time.Minute,
					},
					{
						Text:    "30m",
						Clicked: autoTres30,
						State:   autoTimeTresh == -30*time.Minute,
					},
					{
						Text:    "1h",
						Clicked: autoTres60,
						State:   autoTimeTresh == -60*time.Minute,
					},
					{
						Text:    "Substract time after inactive",
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
						Text:    "Add 15",
						Clicked: add15,
					},
					{
						Text:    "Add 30",
						Clicked: add30,
					},
					{
						Text:    "Remove 15",
						Clicked: sub15,
					},
					{
						Text:    "Remove 30",
						Clicked: sub30,
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
	go tracker()
	app := menuet.App()
	app.Children = menuItems
	app.Label = "AlvMenuTime"
	app.AutoUpdate.Version = "v0.3"
	app.AutoUpdate.Repo = "jantb/time"
	app.RunApplication()
}
