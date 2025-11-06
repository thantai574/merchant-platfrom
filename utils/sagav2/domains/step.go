package domains

import "context"

type Step struct {
	Name           string
	Index          int
	Func           func(ctx context.Context) error
	CompensateFunc func(ctx context.Context) error
}

type Result struct {
	ExecutionError   error
	CompensateErrors []error
}
