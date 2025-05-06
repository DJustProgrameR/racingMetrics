package model

// Err model
type Err string

// Error returns err text
func (e Err) Error() string {
	return string(e)
}

const (
	errSetDrawTimes        Err = "cannot reset draw time for runner"
	errInvalidTimeFormat   Err = "invalid time format"
	errOnLine              Err = "draw time ain't set"
	errStart               Err = "not on the start"
	errNotRunningMainLap   Err = "not running main lap"
	errNotOnFiringRange    Err = "not on firing range"
	errNotAfterFiringRange Err = "started penalty not exactly after firing"
	errQuitPenalty         Err = "not running penalty lap"
)
