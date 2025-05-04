package model

type Err string

func (e Err) Error() string {
	return string(e)
}

const (
	errSetDrawTimes Err = "cannot reset draw time for runner"
)
