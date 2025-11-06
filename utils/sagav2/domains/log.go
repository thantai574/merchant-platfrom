package domains

import "time"

//noinspection ALL
const (
	LogTypeStartSaga SagaState = iota
	LogTypeSagaStepExec
	LogTypeSagaAbort
	LogTypeSagaStepCompensate
	LogTypeSagaComplete
)

type SagaState int

func (st SagaState) ToString() string {
	var data = []string{"LogTypeStartSaga", "LogTypeSagaStepExec", "LogTypeSagaAbort", "LogTypeSagaStepCompensate", "LogTypeSagaComplete"}
	return data[st]
}

type Log struct {
	ExecutionID  string
	State        SagaState
	Time         time.Time
	StepNumber   int
	StepError    string
	StepDuration time.Duration
	Data         interface{}
}
