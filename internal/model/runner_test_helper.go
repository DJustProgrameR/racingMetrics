package model

import (
	"testing"
)

func newTestRunner(t *testing.T) (*Runner, error) {
	return NewRunner(Config{
		Laps:          3,
		LapLen:        1000,
		PenaltyLen:    200,
		StartDelta:    "00:00:30",
		TargetsAmount: 5,
	}, 1)
}

func mustSetStartTime(t *testing.T, r *Runner, timeStr string) {
	if err := r.SetStartTime(timeStr); err != nil {
		t.Fatalf("SetStartTime failed: %v", err)
	}
}

func mustOnLine(t *testing.T, r *Runner) {
	if err := r.OnLine(); err != nil {
		t.Fatalf("OnLine failed: %v", err)
	}
}

func mustStart(t *testing.T, r *Runner, timeStr string) bool {
	started, err := r.Start(timeStr)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	return started
}
