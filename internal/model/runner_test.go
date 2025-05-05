package model

import (
	"log"
	"strings"
	"testing"
)

func TestNewRunner(t *testing.T) {
	r, err := newTestRunner(t)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if r.state != registered {
		t.Errorf("Expected state registered, got %v", r.state)
	}
	if r.runnerID != 1 {
		t.Errorf("Expected runnerID 1, got %d", r.runnerID)
	}
	if r.totalRaceLaps != 3 {
		t.Errorf("Expected 3 laps, got %d", r.totalRaceLaps)
	}
}

func TestHappyPathScenario(t *testing.T) {
	r, err := newTestRunner(t)
	if err != nil {
		log.Fatalf("%v", err)
	}

	mustSetStartTime(t, r, "10:00:00.000")
	if r.state != timeSet {
		t.Errorf("Expected state timeSet, got %v", r.state)
	}

	mustOnLine(t, r)
	if r.state != onLine {
		t.Errorf("Expected state onLine, got %v", r.state)
	}

	started := mustStart(t, r, "10:00:10.000")
	if !started {
		t.Error("Expected runner to start successfully")
	}
	if r.state != runningMain {
		t.Errorf("Expected state runningMain, got %v", r.state)
	}
	if r.startDiff != 10000 {
		t.Errorf("Expected startDiff 10000, got %d", r.startDiff)
	}

	r.lastFinishLineTime = 36010000
	finishedRun, err := r.FinishLap("10:01:10.000")
	if err != nil {
		t.Fatalf("FinishLap failed: %v", err)
	}
	if finishedRun {
		t.Error("Race should not be finished after first lap")
	}
	if r.state != runningMain {
		t.Errorf("Expected state runningMain, got %v", r.state)
	}
	if len(r.lapTimes) != 1 || r.lapTimes[0] != 60000 {
		t.Errorf("Expected lap time 60000, got %v", r.lapTimes)
	}

	if err := r.StartFiring(2); err != nil {
		t.Fatalf("StartFiring failed: %v", err)
	}
	if r.state != firing {
		t.Errorf("Expected state firing, got %v", r.state)
	}

	for i := 0; i < 3; i++ {
		if err := r.HitTarget(i); err != nil {
			t.Fatalf("HitTarget failed: %v", err)
		}
	}
	if r.targetHit != 3 {
		t.Errorf("Expected 3 targets hit, got %d", r.targetHit)
	}

	rangeID, err := r.QuitFiring()
	if err != nil {
		t.Fatalf("QuitFiring failed: %v", err)
	}
	if rangeID != 2 {
		t.Errorf("Expected rangeID 2, got %d", rangeID)
	}
	if r.state != leftFiringRange {
		t.Errorf("Expected state leftFiringRange, got %v", r.state)
	}

	if err := r.StartPenalty("10:02:10.000"); err != nil {
		t.Fatalf("StartPenalty failed: %v", err)
	}
	if r.state != runningPenalty {
		t.Errorf("Expected state runningPenalty, got %v", r.state)
	}
	if r.penaltyLaps != 1 {
		t.Errorf("Expected 1 penalty lap, got %d", r.penaltyLaps)
	}

	if err := r.QuitPenalty("10:02:30.000"); err != nil {
		t.Fatalf("QuitPenalty failed: %v", err)
	}
	if r.state != runningMain {
		t.Errorf("Expected state runningMain, got %v", r.state)
	}
	if r.penaltyTime != 20000 {
		t.Errorf("Expected penaltyTime 20000, got %d", r.penaltyTime)
	}

	finishedRun, err = r.FinishLap("10:03:30.000")
	if err != nil {
		t.Fatalf("FinishLap failed: %v", err)
	}
	if finishedRun {
		t.Error("Race should not be finished after second lap")
	}
	if r.state != runningMain {
		t.Errorf("Expected state runningMain, got %v", r.state)
	}
	if len(r.lapTimes) != 2 || r.lapTimes[1] != 140000 {
		t.Errorf("Expected second lap time 140000, got %v", r.lapTimes)
	}

	finishedRun, err = r.FinishLap("10:04:40.000")
	if err != nil {
		t.Fatalf("FinishLap failed: %v", err)
	}
	if !finishedRun {
		t.Error("Race should be finished after third lap")
	}
	if r.state != finished {
		t.Errorf("Expected state finished, got %v", r.state)
	}

	totalTime, result := r.GetResult()
	if totalTime != 280000 {
		t.Errorf("Expected total time 280000, got %d", totalTime)
	}
	if !strings.Contains(result, "[00:04:40.000]") {
		t.Errorf("Result string unexpected: %s", result)
	}
}

func TestLateStartScenario(t *testing.T) {
	r, err := newTestRunner(t)
	if err != nil {
		log.Fatalf("%v", err)
	}

	mustSetStartTime(t, r, "10:00:00.000")
	mustOnLine(t, r)

	started := mustStart(t, r, "10:00:40.000")
	if started {
		t.Error("Expected runner to fail start")
	}
	if r.state != notStarted {
		t.Errorf("Expected state notStarted, got %v", r.state)
	}

	_, result := r.GetResult()
	if !strings.Contains(result, string(notStartedStatus)) {
		t.Errorf("Expected not started status, got %s", result)
	}
}

func TestNotFinishedScenario(t *testing.T) {
	r, err := newTestRunner(t)
	if err != nil {
		log.Fatalf("%v", err)
	}

	mustSetStartTime(t, r, "10:00:00.000")
	mustOnLine(t, r)
	mustStart(t, r, "10:00:10.000")

	r.lastFinishLineTime = 10000
	_, err = r.FinishLap("10:01:10.000")
	if err != nil {
		t.Fatal(err)
	}

	if err := r.QuitRunning(); err != nil {
		t.Fatal(err)
	}
	if r.state != notFinished {
		t.Errorf("Expected state notFinished, got %v", r.state)
	}

	_, result := r.GetResult()
	if !strings.Contains(result, string(notFinishedStatus)) {
		t.Errorf("Expected not finished status, got %s", result)
	}
}

func TestInvalidStateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Runner)
		operation func(*Runner) error
		expected  error
	}{
		{
			name: "OnLine before SetStartTime",
			setup: func(r *Runner) {

			},
			operation: func(r *Runner) error {
				return r.OnLine()
			},
			expected: errOnLine,
		},
		{
			name: "Start before OnLine",
			setup: func(r *Runner) {
				mustSetStartTime(t, r, "10:00:00.000")
			},
			operation: func(r *Runner) error {
				_, err := r.Start("10:00:10.000")
				return err
			},
			expected: errStart,
		},
		{
			name: "StartFiring not in runningMain",
			setup: func(r *Runner) {
				mustSetStartTime(t, r, "10:00:00.000")
				mustOnLine(t, r)
				mustStart(t, r, "10:00:10.000")

				r.state = onLine
			},
			operation: func(r *Runner) error {
				return r.StartFiring(1)
			},
			expected: errNotRunningMainLap,
		},
		{
			name: "HitTarget not in firing",
			setup: func(r *Runner) {
				mustSetStartTime(t, r, "10:00:00.000")
				mustOnLine(t, r)
				mustStart(t, r, "10:00:10.000")
			},
			operation: func(r *Runner) error {
				return r.HitTarget(1)
			},
			expected: errNotOnFiringRange,
		},
		{
			name: "QuitFiring not in firing",
			setup: func(r *Runner) {
				mustSetStartTime(t, r, "10:00:00.000")
				mustOnLine(t, r)
				mustStart(t, r, "10:00:10.000")
			},
			operation: func(r *Runner) error {
				_, err := r.QuitFiring()
				return err
			},
			expected: errNotOnFiringRange,
		},
		{
			name: "StartPenalty not after firing",
			setup: func(r *Runner) {
				mustSetStartTime(t, r, "10:00:00.000")
				mustOnLine(t, r)
				mustStart(t, r, "10:00:10.000")
			},
			operation: func(r *Runner) error {
				return r.StartPenalty("10:01:00.000")
			},
			expected: errNotAfterFiringRange,
		},
		{
			name: "QuitPenalty not in penalty",
			setup: func(r *Runner) {
				mustSetStartTime(t, r, "10:00:00.000")
				mustOnLine(t, r)
				mustStart(t, r, "10:00:10.000")
			},
			operation: func(r *Runner) error {
				return r.QuitPenalty("10:01:00.000")
			},
			expected: errQuitPenalty,
		},
		{
			name: "FinishLap in invalid state",
			setup: func(r *Runner) {
				mustSetStartTime(t, r, "10:00:00.000")
				mustOnLine(t, r)
				mustStart(t, r, "10:00:10.000")
				r.state = firing
			},
			operation: func(r *Runner) error {
				_, err := r.FinishLap("10:01:00.000")
				return err
			},
			expected: errNotRunningMainLap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := newTestRunner(t)
			if err != nil {
				log.Fatalf("%v", err)
			}
			if tt.setup != nil {
				tt.setup(r)
			}
			err = tt.operation(r)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if err.Error() != tt.expected.Error() {
				t.Errorf("Expected error %v, got %v", tt.expected, err)
			}
		})
	}
}

func TestTimeFormatting(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"10:00:00.000", 10 * 3600 * 1000, false},
		{"10:00:00.500", 10*3600*1000 + 500, false},
		{"00:01:00.000", 60 * 1000, false},
		{"10:00:00", 0, true},
		{"10:00:00.abc", 0, true},
		{"25:00:00.000", 0, true},
		{"10:60:00.000", 0, true},
		{"10:00:60.000", 0, true},
		{"10:00:00.1000", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := formatTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("formatTime(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{10*3600*1000 + 500, "10:00:00.500"},
		{60 * 1000, "00:01:00.000"},
		{12345, "00:00:12.345"},
		{0, "00:00:00.000"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := parseTime(tt.input)
			if got != tt.expected {
				t.Errorf("parseTime(%d) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
