// Package model is a container for domain models
package model

// Config is a run config
type Config struct {
	Laps          int    `json:"laps"`
	LapLen        int    `json:"lapLen"`
	PenaltyLen    int    `json:"penaltyLen"`
	FiringLines   int    `json:"firingLines"`
	Start         string `json:"start"`
	StartDelta    string `json:"startDelta"`
	TargetsAmount int
}
