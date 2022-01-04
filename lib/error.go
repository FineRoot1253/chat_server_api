package lib


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