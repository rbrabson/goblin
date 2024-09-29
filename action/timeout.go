package action

import "time"

type timeoutAction struct {
	ActionBase
	Action     Action
	maxRunTime time.Duration
	endTime    time.Time
}

// NewTimeoutAction creates an Action to run the provided action until it completes execution
// (`IsDone` returns `true`), or the maximum run time is exceeded, whichever comes first.
func NewTimeoutAction(action Action, maxRunTime time.Duration) Action {
	timeoutAction := &timeoutAction{
		maxRunTime: maxRunTime,
	}
	return timeoutAction
}

// Initialize initializes the action, and starts the timeout for the action.
func (a *timeoutAction) Initialize() {
	a.Action.Initialize()
	a.endTime = time.Now().Add(a.maxRunTime)
}

// Execute runs the action.
func (a *timeoutAction) Execute() {
	a.Action.Execute()
}

// IsFinished returns true if the timeout for the action has been reached or if the action has
// completed execution, whichever comes first. If neither is true, then it returns `false`.
func (a *timeoutAction) IsDone() bool {
	if time.Now().After(a.endTime) {
		return true
	}
	return a.Action.IsFinished()
}
