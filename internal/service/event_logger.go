// Package service is a business logic
package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"racingMetrics/internal/model"
	"sort"
	"strconv"
	"strings"
)

type rangeStatus int

const (
	registerRunner int = iota + 1
	setRunnerTime
	runnerOnStart
	startRunner
	runnerStartFire
	runnerHitTarget
	runnerQuitFire
	runnerEnterPenalty
	runnerLeftPenalty
	runnerEndMain
	runnerCantRun
)
const (
	rangeFree     rangeStatus = iota
	rangeOccupied rangeStatus = 1
)
const (
	timeInd = iota
	eventIDInd
	runnerIDInd
	extraParamInd
)

const totalTargetsAmount = 5

type runnerInterface interface {
	SetStartTime(time string) error
	OnLine() error
	Start(time string) (bool, error)
	StartFiring(firingRange int) error
	HitTarget(target int) error
	QuitFiring() (int, error)
	StartPenalty(time string) error
	QuitPenalty(time string) error
	FinishLap(time string) (bool, error)
	QuitRunning() error

	GetResult() (int, string)
}

// NewRunLog returns EventLogger
func NewRunLog(jsonConfigPath, eventsPath string, logger *log.Logger) *EventLogger {
	config := parseConfig(jsonConfigPath)
	if logger == nil {
		log.Fatalf("NewRunLog looger is nil")
	}
	return &EventLogger{
		logger:       logger,
		eventsPath:   eventsPath,
		config:       config,
		runners:      make(map[int]runnerInterface),
		firingRanges: make(map[int]rangeStatus),
	}
}

// EventLogger logs incoming events
type EventLogger struct {
	logger       *log.Logger
	eventsPath   string
	config       model.Config
	runners      map[int]runnerInterface
	firingRanges map[int]rangeStatus
}

func parseConfig(jsonConfigPath string) model.Config {
	jsonData, err := os.ReadFile(jsonConfigPath)
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}

	config := model.Config{}
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	config.TargetsAmount = totalTargetsAmount
	return config
}

// RunEvents runs given events
func (s *EventLogger) RunEvents(ctx context.Context) {
	file, err := os.Open(s.eventsPath)
	if err != nil {
		log.Fatalf("Error opening text file: %v", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatalf("Error closing text file: %v", err)
		}
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()
			args := strings.Fields(line)
			s.parseEvent(args)
		}
	}
}

// PrintResultingTable -
func (s *EventLogger) PrintResultingTable() {
	type Log struct {
		time int
		log  string
	}
	successfulRunners := []Log{}
	failedRunners := []Log{}

	for _, runner := range s.runners {
		result := Log{}
		result.time, result.log = runner.GetResult()
		if result.time == 0 {
			failedRunners = append(failedRunners, result)
		} else {
			successfulRunners = append(successfulRunners, result)
		}
	}

	sort.Slice(successfulRunners, func(i, j int) bool {
		return successfulRunners[i].time < successfulRunners[j].time
	})

	fmt.Println("Resulting table")

	for _, result := range successfulRunners {
		fmt.Println(result.log)
	}

	for _, result := range failedRunners {
		fmt.Println(result.log)
	}
}

func (s *EventLogger) parseEvent(args []string) {
	time := args[timeInd][1 : len(args[timeInd])-1]
	eventID, err := strconv.Atoi(args[eventIDInd])
	if err != nil {
		s.logger.Fatalf("Error parsing event ID: %v", err)
	}
	runnerID, err := strconv.Atoi(args[runnerIDInd])
	if err != nil {
		s.logger.Fatalf("Error parsing event ID: %v", err)
	}

	switch eventID {
	case registerRunner:
		s.handleRegisterRunner(time, runnerID)
	case setRunnerTime:
		s.handleSetRunnerTime(time, runnerID, args[extraParamInd])
	case runnerOnStart:
		s.handleRunnerOnStart(time, runnerID)
	case startRunner:
		s.handleStartRunner(time, runnerID)
	case runnerStartFire:
		s.handleRunnerStartFire(time, runnerID, args[extraParamInd])
	case runnerHitTarget:
		s.handleRunnerHitTarget(time, runnerID, args[extraParamInd])
	case runnerQuitFire:
		s.handleRunnerQuitFire(time, runnerID)
	case runnerEnterPenalty:
		s.handleRunnerEnterPenalty(time, runnerID)
	case runnerLeftPenalty:
		s.handleRunnerLeftPenalty(time, runnerID)
	case runnerEndMain:
		s.handleRunnerEndMain(time, runnerID)
	case runnerCantRun:
		s.handleRunnerCantRun(time, runnerID, args[extraParamInd])
	default:
		fmt.Printf("[%s] No such event for competitor(%d)\n", time, runnerID)
	}
}

func (s *EventLogger) handleRegisterRunner(time string, runnerID int) {
	if _, ok := s.runners[runnerID]; ok {
		s.logger.Fatalf("Can't register same competitor twice: %d", runnerID)
	}
	runner, err := model.NewRunner(
		s.config,
		runnerID)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	s.runners[runnerID] = runner

	fmt.Printf("[%s] The competitor(%d) registered\n", time, runnerID)
}

func (s *EventLogger) handleSetRunnerTime(time string, runnerID int, drawTime string) {
	s.checkRunnerExist(runnerID)

	err := s.runners[runnerID].SetStartTime(drawTime)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	fmt.Printf("[%s] The start time for the competitor(%d) was set by a draw to %s\n", time, runnerID, drawTime)
}

func (s *EventLogger) handleRunnerOnStart(time string, runnerID int) {
	s.checkRunnerExist(runnerID)

	err := s.runners[runnerID].OnLine()
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}
	fmt.Printf("[%s] The competitor(%d) is on the start line\n", time, runnerID)
}

func (s *EventLogger) handleStartRunner(time string, runnerID int) {
	s.checkRunnerExist(runnerID)

	started, err := s.runners[runnerID].Start(time)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	fmt.Printf("[%s] The competitor(%d) has started\n", time, runnerID)
	if !started {
		fmt.Printf("[%s] The competitor(%d) is disqualified\n", time, runnerID)
	}
}

func (s *EventLogger) handleRunnerStartFire(time string, runnerID int, firingRangeStr string) {
	s.checkRunnerExist(runnerID)

	firingRange, err := strconv.Atoi(firingRangeStr)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	if status := s.firingRanges[firingRange]; status == rangeOccupied {
		s.logger.Fatalf("Err at event %s: %s", time, "range occupied")
	}
	s.firingRanges[firingRange] = rangeOccupied

	err = s.runners[runnerID].StartFiring(firingRange)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	fmt.Printf("[%s] The competitor(%d) is on the firing range(%d)\n", time, runnerID, firingRange)
}

func (s *EventLogger) handleRunnerHitTarget(time string, runnerID int, targetStr string) {
	s.checkRunnerExist(runnerID)

	target, err := strconv.Atoi(targetStr)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	err = s.runners[runnerID].HitTarget(target)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	fmt.Printf("[%s] The target(%d) has been hit by competitor(%d)\n", time, target, runnerID)
}

func (s *EventLogger) handleRunnerQuitFire(time string, runnerID int) {
	s.checkRunnerExist(runnerID)

	firingRange, err := s.runners[runnerID].QuitFiring()
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}
	s.firingRanges[firingRange] = rangeFree

	fmt.Printf("[%s] The competitor(%d) left the firing range\n", time, runnerID)
}

func (s *EventLogger) handleRunnerEnterPenalty(time string, runnerID int) {
	s.checkRunnerExist(runnerID)

	err := s.runners[runnerID].StartPenalty(time)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	fmt.Printf("[%s] The competitor(%d) entered the penalty laps\n", time, runnerID)
}

func (s *EventLogger) handleRunnerLeftPenalty(time string, runnerID int) {
	s.checkRunnerExist(runnerID)

	err := s.runners[runnerID].QuitPenalty(time)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	fmt.Printf("[%s] The competitor(%d) left the penalty laps\n", time, runnerID)
}

func (s *EventLogger) handleRunnerEndMain(time string, runnerID int) {
	s.checkRunnerExist(runnerID)

	finished, err := s.runners[runnerID].FinishLap(time)
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	fmt.Printf("[%s] The competitor(%d) ended the main lap\n", time, runnerID)

	if finished {
		fmt.Printf("[%s] The competitor(%d) has finished\n", time, runnerID)
	}
}

func (s *EventLogger) handleRunnerCantRun(time string, runnerID int, comment string) {
	s.checkRunnerExist(runnerID)

	err := s.runners[runnerID].QuitRunning()
	if err != nil {
		s.logger.Fatalf("Err at event %s: %v", time, err)
	}

	fmt.Printf("[%s] The competitor(%d) can`t continue: %s\n", time, runnerID, comment)
}

func (s *EventLogger) checkRunnerExist(runnerID int) {
	if _, ok := s.runners[runnerID]; !ok {
		s.logger.Fatalf("No such runner registered: %d", runnerID)
	}
}
