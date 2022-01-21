package utils


type ErrorInt interface {
	CurrentError() ErrorStateType
}

func New(state ErrorStateType) ErrorInt {
	return &errorIntModel{state}
}

type errorIntModel struct {
	errorState ErrorStateType
}

func (e *errorIntModel) CurrentError() ErrorStateType{
	return e.errorState
}

type CommonError struct {
	Func string // the failing functions
	Data  string // the input
	Err  error  // the reason the conversion failed (e.g. ErrRange, ErrSyntax, etc.)
}

func (e *CommonError) Error() string {
	return e.Func + ": " + "parsing " + e.Data + ": " + e.Err.Error()
}

func (e *CommonError) Unwrap() error { return e.Err }

