package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type state int

const (
	registered state = iota
	timeSet
	onLine
	notStarted
	notFinished
	runningMain
	firing
	leftFiringRange
	runningPenalty
	finished
)

type failStatus string

const (
	notStartedStatus  failStatus = "NotStarted"
	notFinishedStatus failStatus = "NotFinished"
)

const timeLayout = "15:04:05"

// Runner info
type Runner struct {
	totalRaceLaps int
	lapLen        int
	penaltyLapLen int
	startDelta    int
	runnerID      int

	state state

	drawStartTime int

	startDiff int

	lastFinishLineTime int

	laps       int
	lapTimes   []int
	avLapSpeed []float64

	lastPenaltyTime int
	penaltyLaps     int
	penaltyTime     int

	firingRange   int
	targetHit     int
	targetsAmount int
}

// NewRunner returns new Runner
func NewRunner(
	config Config,
	runnerID int,
) (*Runner, error) {
	startDeltaInt, err := formatTimeNoMill(config.StartDelta)
	if err != nil {
		return nil, err
	}
	return &Runner{
		totalRaceLaps: config.Laps,
		lapLen:        config.LapLen,
		penaltyLapLen: config.PenaltyLen,
		startDelta:    startDeltaInt,
		targetsAmount: config.TargetsAmount,
		runnerID:      runnerID,
		state:         registered,
	}, nil
}

// SetStartTime sets draw-start time for runner
func (r *Runner) SetStartTime(time string) error {
	if r.state == registered {
		timeInt, err := formatTime(time)
		if err != nil {
			return err
		}
		r.drawStartTime = timeInt
		r.state = timeSet
		return nil
	}
	return errSetDrawTimes
}

// OnLine starts run at the time
func (r *Runner) OnLine() error {
	if r.state == timeSet {
		r.state = onLine
		return nil
	}
	return errOnLine
}

// Start starts run at the time
func (r *Runner) Start(time string) (bool, error) {
	if r.state == onLine {
		timeInt, err := formatTime(time)
		if err != nil {
			return false, err
		}
		r.startDiff = timeInt - r.drawStartTime
		r.lastFinishLineTime = timeInt
		if r.startDiff > r.startDelta {
			r.state = notStarted
			return false, nil
		} else {
			r.state = runningMain
			return true, nil
		}
	}
	return false, errStart
}

// StartFiring sets runner on firing range
func (r *Runner) StartFiring(firingRange int) error {
	if r.state == runningMain {
		r.firingRange = firingRange
		r.state = firing
		return nil
	}
	return errNotRunningMainLap
}

// HitTarget hits the target
func (r *Runner) HitTarget(_ int) error {
	if r.state == firing {
		r.targetHit++
		return nil
	}
	return errNotOnFiringRange
}

// QuitFiring runner
func (r *Runner) QuitFiring() (int, error) {
	if r.state == firing {
		r.state = leftFiringRange
		return r.firingRange, nil
	}
	return 0, errNotOnFiringRange
}

// StartPenalty runner is on penalty lap
func (r *Runner) StartPenalty(time string) error {
	if r.state == leftFiringRange {
		timeInt, err := formatTime(time)
		if err != nil {
			return err
		}
		r.state = runningPenalty
		r.penaltyLaps++
		r.lastPenaltyTime = timeInt
		return nil
	}
	return errNotAfterFiringRange
}

// QuitPenalty runner quits penalty lap
func (r *Runner) QuitPenalty(time string) error {
	if r.state == runningPenalty {
		timeInt, err := formatTime(time)
		if err != nil {
			return err
		}
		r.state = runningMain
		r.penaltyTime += timeInt - r.lastPenaltyTime
		return nil
	}
	return errQuitPenalty
}

// FinishLap runner finished another lap
func (r *Runner) FinishLap(time string) (bool, error) {
	if r.state == runningMain || r.state == leftFiringRange {
		r.state = runningMain
		timeInt, err := formatTime(time)
		if err != nil {
			return false, err
		}
		lapTime := timeInt - r.lastFinishLineTime
		r.lapTimes = append(r.lapTimes, lapTime)
		r.avLapSpeed = append(r.avLapSpeed, float64(r.lapLen)*1000.0/float64(lapTime))
		r.laps++
		finishRunning := r.totalRaceLaps == r.laps
		if finishRunning {
			r.state = finished
		}
		r.lastFinishLineTime = timeInt
		return finishRunning, nil
	}
	return false, errNotRunningMainLap
}

// QuitRunning runner quit running for some reason
func (r *Runner) QuitRunning() error {
	r.state = notFinished
	return nil
}

func formatTime(timeStr string) (int, error) {
	timeSep := strings.Split(timeStr, ".")
	if len(timeSep) == 1 || len(timeSep[1]) != 3 {
		return 0, errInvalidTimeFormat
	}
	milliseconds, err := strconv.Atoi(timeSep[1])
	if err != nil {
		return 0, errInvalidTimeFormat
	}
	t, err := time.Parse(timeLayout, timeSep[0])
	if err != nil {
		return 0, errInvalidTimeFormat
	}

	totalMilliseconds := t.Hour()*3600*1000 + t.Minute()*60*1000 + t.Second()*1000 + milliseconds
	return totalMilliseconds, nil
}

func formatTimeNoMill(timeStr string) (int, error) {
	t, err := time.Parse(timeLayout, timeStr)
	if err != nil {
		return 0, errInvalidTimeFormat
	}

	totalMilliseconds := t.Hour()*3600*1000 + t.Minute()*60*1000 + t.Second()*1000
	return totalMilliseconds, nil
}

func parseTime(time int) string {
	hours := time / (3600 * 1000)
	minutes := (time % (3600 * 1000)) / (60 * 1000)
	seconds := (time % (60 * 1000)) / 1000
	milliseconds := time % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}

// GetResult returns run results
func (r *Runner) GetResult() (int, string) {
	var result string
	var totalTime int

	lapResults := []string{}
	for i := 0; i < len(r.lapTimes); i++ {
		lapResult := fmt.Sprintf("{%s,%f}", parseTime(r.lapTimes[i]), r.avLapSpeed[i])
		lapResults = append(lapResults, lapResult)
	}
	for i := 0; i < r.totalRaceLaps-len(r.lapTimes); i++ {
		lapResults = append(lapResults, "{,}")
	}
	lapResultsS := strings.Join(lapResults, ", ")
	if r.state == finished {
		totalTime = r.lastFinishLineTime - r.drawStartTime
		totalTimeS := parseTime(totalTime)
		penaltyTime := parseTime(r.penaltyTime)
		var avPenaltySpeed float64
		if r.penaltyTime > 0 {
			avPenaltySpeed = float64(r.penaltyLaps*r.penaltyLapLen*1000) / float64(r.penaltyTime)
		}
		result = fmt.Sprintf("[%s] %d [%s] {%s, %f} %d/%d", totalTimeS, r.runnerID, lapResultsS, penaltyTime, avPenaltySpeed, r.targetHit, r.laps*r.targetsAmount)
	} else if r.state == notStarted {
		result = fmt.Sprintf("[%s] %d [%s] {,} 0/0", notStartedStatus, r.runnerID, lapResultsS)
	} else if r.state == notFinished {
		penaltyTime := parseTime(r.penaltyTime)
		avPenaltySpeed := float64(r.penaltyLaps*r.penaltyLapLen*1000) / float64(r.penaltyTime)
		result = fmt.Sprintf("[%s] %d [%s] {%s, %f} %d/%d", notFinishedStatus, r.runnerID, lapResultsS, penaltyTime, avPenaltySpeed, r.targetHit, r.laps*5)
	}

	return totalTime, result
}
