package action

import "time"

// a timeoutAction is an Action that runs another action for a maximum amount of time. Once either the action
// completes or the timeout is reached, the timeout is completed.
type timeoutAction struct {
	ActionBase
	action          Action
	maxRunTime      time.Duration
	endTime         time.Time
	interruptCalled bool
}

// NewTimeoutAction creates an Action to run the provided action until it completes execution
// (`IsDone` returns `true`), or the maximum run time is exceeded, whichever comes first.
func NewTimeoutAction(action Action, maxRunTime time.Duration) *timeoutAction {
	timeoutAction := &timeoutAction{
		maxRunTime: maxRunTime,
	}
	return timeoutAction
}

// Initialize initializes the action, and starts the timeout for the action.
func (a *timeoutAction) Initialize() {
	a.action.Initialize()
	a.endTime = time.Now().Add(a.maxRunTime)
}

// Execute runs the action.
func (a *timeoutAction) Execute() {
	a.action.Execute()
}

// IsFinished returns true if the timeout for the action has been reached or if the action has
// completed execution, whichever comes first. If neither is true, then it returns `false`.
func (a *timeoutAction) IsDone() bool {
	if time.Now().After(a.endTime) {
		a.action.Interrupt()
		a.interruptCalled = true
		return true
	}
	return a.action.IsFinished()
}
