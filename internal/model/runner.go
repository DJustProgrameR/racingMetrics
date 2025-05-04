package model

import (
	"fmt"
)

type state int

const (
	registered state = iota
	waiting
	notStarted
	notFinished
	runningMain
	firing
	runningPenalty
)

type Runner struct {
	totalRaceLaps int
	lapLen        int
	penaltyLapLen int
	startDelta    int
	runnerID      int

	state state

	drawStartTime int

	startDiff int64

	lastFinishLineTime int64

	laps       int
	lapTimes   []int
	avLapSpeed []float64

	lastPenaltyTime int
	penaltyLaps     int
	penaltyTime     int

	targetHit int
}

func NewRunner(
	totalRaceLaps int,
	lapLen int,
	penaltyLapLen int,
	startDelta int,
	runnerID int,
) *Runner {
	return &Runner{
		totalRaceLaps: totalRaceLaps,
		lapLen:        lapLen,
		penaltyLapLen: penaltyLapLen,
		startDelta:    startDelta,
		runnerID:      runnerID,
	}
}

func (r *Runner) GotStartTime(time int) error {
	if r.state == registered {
		r.drawStartTime = time
		r.state = waiting
		return nil
	}
	return fmt.Errorf("%s: %d", errSetDrawTimes.Error(), r.runnerID)
}

func (r *Runner) Start(time int) error        {}
func (r *Runner) StartFiring(time int) error  {}
func (r *Runner) HitTarget(time int) error    {}
func (r *Runner) QuitFiring(time int) error   {}
func (r *Runner) StartPenalty(time int) error {}
func (r *Runner) QuitPenalty(time int) error  {}
func (r *Runner) FinishLap(time int) error    {}
func (r *Runner) QuitRunning(time int) error  {}

func (r *Runner) GetLog() string {}
