package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
	"time"
)

type Tracking struct {
	Days       []Day
	TaskId     int
	AlvTimeKey string
}

type Day struct {
	Date  string
	Total string
	Times []TimeStruct
}

type TimeStruct struct {
	ClockIn  time.Time
	ClockOut time.Time
}

func (tracking Tracking) getTimes() []TimeStruct {
	return tracking.Days[len(tracking.Days)-1].Times
}

func (tracking Tracking) hoursForToday() time.Duration {
	duration := int64(0)
	for _, timeStruct := range tracking.getTimes() {
		if timeStruct.ClockOut.IsZero() {
			duration += time.Now().Sub(timeStruct.ClockIn).Nanoseconds()
			continue
		}
		duration += timeStruct.ClockOut.Sub(timeStruct.ClockIn).Nanoseconds()
	}
	return time.Duration(duration)
}

func (tracking *Tracking) updateTotalForToday(duration time.Duration) {
	tracking.Days[len(tracking.Days)-1].Total = duration.String()
}

func (tracking *Tracking) clockInNow() {
	if len(tracking.getTimes()) == 0 || !tracking.getTimes()[len(tracking.getTimes())-1].ClockOut.IsZero() {
		tracking.Days[len(tracking.Days)-1].Times = append(tracking.Days[len(tracking.Days)-1].Times, TimeStruct{
			ClockIn: time.Now(),
		})
		tracking.store()
	}
}

func (tracking *Tracking) clockOutNow() {
	if tracking.canClockOut() {
		tracking.getTimes()[len(tracking.getTimes())-1].ClockOut = time.Now()
		tracking.store()
	}
}

func (tracking Tracking) canClockOut() bool {
	return len(tracking.getTimes()) != 0 && tracking.getTimes()[len(tracking.getTimes())-1].ClockOut.IsZero()
}

func (tracking *Tracking) reset() {
	tracking.Days = append(tracking.Days, Day{
		Date:  time.Now().Format("02.01.2006"),
		Total: "",
		Times: []TimeStruct{},
	})
}

func (tracking *Tracking) addDuration(dur time.Duration) {
	tracking.clockOutNow()
	tracking.Days[len(tracking.Days)-1].Times = append(tracking.getTimes(), TimeStruct{
		ClockIn:  time.Now(),
		ClockOut: time.Now().Add(dur),
	})
}

func (tracking *Tracking) subAutoTresh(autoTimeTresh time.Duration) {
	tracking.clockOutNow()
	tracking.Days[len(tracking.Days)-1].Times = append(tracking.getTimes(), TimeStruct{
		ClockIn:  time.Now(),
		ClockOut: time.Now().Add(autoTimeTresh),
	})
}

func (tracking *Tracking) setTaskId(taskId int) {
	tracking.TaskId = taskId
}

func (tracking *Tracking) setAlvTimeKey(key string) {
	tracking.AlvTimeKey = key
}

func (tracking Tracking) store() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := json.MarshalIndent(tracking, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(filepath.Join(usr.HomeDir, ".alvTimeTracker.json"), bytes, 0600)
}

func (tracking *Tracking) load() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadFile(filepath.Join(usr.HomeDir, ".alvTimeTracker.json"))
	if err != nil {
		bytes, err := json.MarshalIndent(tracking, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(filepath.Join(usr.HomeDir, ".alvTimeTracker.json"), bytes, 0600)
	}

	json.Unmarshal(bytes, &tracking)
}
