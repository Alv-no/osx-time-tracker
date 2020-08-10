package main

import (
	"fmt"
	alvtimeClient "github.com/Alv-no/alvtime-go-client"
	"github.com/caseymrm/menuet"
	"github.com/jantb/robotgo"
	"log"
	"os/exec"
	"runtime"
	"strconv"
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
	if tracking.AlvTimeKey == "" {
		alert := menuet.App().Alert(menuet.Alert{
			MessageText:     "Need accesskey to continue",
			InformativeText: "Please enter it below",
			Buttons:         []string{"Set"},
			Inputs:          []string{"Access key"},
		})
		tracking.setAlvTimeKey(alert.Inputs[0])
	}
	c, err := alvtimeClient.New("https://alvtime-api-prod.azurewebsites.net", tracking.AlvTimeKey)
	t, err := c.GetTasks()
	if err != nil {
		fmt.Println(err)
	}
	tasks = t
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
	openbrowser("https://alvtime.no/")
}

func setAlvTime() {
	value := getTodaysHoursRoundedTo15Minutes()

	c, err := alvtimeClient.New("https://alvtime-api-prod.azurewebsites.net", tracking.AlvTimeKey)
	if err != nil {
		log.Fatal(err)
	}
	_, err = c.EditTimeEntries([]alvtimeClient.TimeEntrie{{Date: time.Now().Format("2006-01-02"), Value: float32(value), TaskID: tracking.TaskId}})
	if err != nil {
		log.Fatal(err)
	}
}

func getTodaysHoursRoundedTo15Minutes() float64 {
	duration := tracking.hoursForToday()
	d := duration.Round(15 * time.Minute)

	value, err := strconv.ParseFloat(fmt.Sprintf("%.2f", d.Hours()), 32)
	if err != nil {
		log.Fatal(err)
	}
	return value
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
	openbrowser("https://mytime.experis.no/")
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
		}, {
			Text: "Set default task",
			Children: func() []menuet.MenuItem {

				var items []menuet.MenuItem
				for i, task := range tasks {
					if task.Favorite {
						clickedTask := tasks[i]
						items = append(items, menuet.MenuItem{
							Text: fmt.Sprintf("%s %s %s %s", task.Name, task.Description, task.Project.Customer.Name, task.Project.Name),
							Clicked: func() {
								tracking.setTaskId(clickedTask.ID)
							},
							State: tracking.TaskId == tasks[i].ID,
						})
					}

				}
				return items
			},
		},
		{
			Type: menuet.Separator,
		},
		{
			Text:    "AlvTime Website",
			Clicked: openAlvTime,
		},
		{
			Text: fmt.Sprintf("Set %.2f on %s %s %s %s in Alvtime",
				getTodaysHoursRoundedTo15Minutes(), getDefaultTask().Name, getDefaultTask().Description, getDefaultTask().Project.Customer.Name, getDefaultTask().Project.Name),
			Clicked: setAlvTime,
		},
		{
			Text:    "Experis",
			Clicked: openExperis,
		},
	}
	return items
}

func getDefaultTask() alvtimeClient.Task {
	for _, task := range tasks {
		if task.ID == tracking.TaskId {
			return task
		}
	}
	return alvtimeClient.Task{}
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

var tasks = []alvtimeClient.Task{}

func main() {
	tracking.load()
	go tracker()
	app := menuet.App()
	app.Children = menuItems
	app.Label = "AlvMenuTime"
	app.AutoUpdate.Version = "v0.4"
	app.AutoUpdate.Repo = "jantb/time"
	app.RunApplication()
}
