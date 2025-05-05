package service

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"racingMetrics/internal/model"
	"strconv"
	"strings"
)

type event int

const (
	registerRunner event = iota + 1
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

	timeInd       = 0
	eventIDInd    = 1
	runnerIDInd   = 2
	extraParamInd = 3
)

type Runner interface {
	SetStartTime(time string) error

	Start(time string) error
	StartFiring(time string) error
	HitTarget(time string) error
	QuitFiring(time string) error
	StartPenalty(time string) error
	QuitPenalty(time string) error
	FinishLap(time string) error
	QuitRunning(time string) error

	GetLog() string
}

func NewRunLog() *RunLogService {
	return &RunLogService{
		runners: make(map[int]Runner),
	}
}

type RunLogService struct {
	config  model.Config
	runners map[int]Runner
}

func (s *RunLogService) ParseFiles(jsonConfigPath, eventsPath string) {
	jsonData, err := os.ReadFile(jsonConfigPath)
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}

	config := model.Config{}
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	s.config = config

	file, err := os.Open(eventsPath)
	if err != nil {
		log.Fatalf("Error opening text file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		args := strings.Fields(line)
		s.parseEvent(args)
	}
}

func (s *RunLogService) parseEvent(args []string) {
	time := args[timeInd]
	eventID, err := strconv.Atoi(args[eventIDInd])
	if err != nil {
		log.Fatalf("Error parsing event ID: %v", err)
	}
	runnerID, err := strconv.Atoi(args[runnerIDInd])
	if err != nil {
		log.Fatalf("Error parsing event ID: %v", err)
	}
	switch event(eventID) {
	case registerRunner:
		if _, ok := s.runners[runnerID]; ok {
			log.Fatalf("Can't register same competitor twice: %d", runnerID)
		}
		runner := model.NewRunner(
			s.config.Laps,
			s.config.LapLen,
			s.config.PenaltyLen,
			s.config.StartDelta,
			runnerID)
		s.runners[runnerID] = runner
	case setRunnerTime:
		if _, ok := s.runners[runnerID]; ok {
			log.Fatalf("Can't register same runner twice: %d", runnerID)
		}
		drawTime := args[extraParamInd]
		err := s.runners[runnerID].SetStartTime(drawTime)
		if err != nil {
			log.Fatalf("Err at event %s: %v", time, err)
		}
	case runnerOnStart:
		s.checkRunnerExist(runnerID)
	case startRunner:
		s.checkRunnerExist(runnerID)

		err := s.runners[runnerID].Start(time)
		if err != nil {
			log.Fatalf("Err at event %s: %v", time, err)
		}
	case runnerStartFire:
		s.checkRunnerExist(runnerID)

		firingRange, err := strconv.Atoi(args[extraParamInd])
		if err != nil {
			log.Fatalf("Err at event %s: %v", time, err)
		}

		err = s.runners[runnerID].StartFiring(time)
		if err != nil {
			log.Fatalf("Err at event %s: %v", time, err)
		}
	case runnerHitTarget:
		s.checkRunnerExist(runnerID)

		err := s.runners[runnerID].SetStartTime(time)
		if err != nil {
			log.Fatalf("Err at event %s: %v", time, err)
		}
	case runnerQuitFire:
	case runnerEnterPenalty:
	case runnerLeftPenalty:
	case runnerEndMain:
	case runnerCantRun:
	default:
	}
}

func (s *RunLogService) checkRunnerExist(runnerID int) {
	if _, ok := s.runners[runnerID]; ok {
		log.Fatalf("Can't register same competitor twice: %d", runnerID)
	}
}
