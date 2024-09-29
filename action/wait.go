package action

import "time"

type waitAction struct {
	ActionBase
	wait    time.Duration
	endTime time.Time
}

// NewWaitAction is an action that executes until the provided wait duration has been reached.
func NewWaitAction(wait time.Duration) *waitAction {
	action := &waitAction{
		wait: wait,
	}
	return action
}

// Initialize is used to initialize the action.
func (a *waitAction) Initialize() {
	a.endTime = time.Now().Add(a.wait)
}

// Execute runs the action. This is a no-op for a base Action.
func (a *waitAction) Execute() {
	// No-Op
}

// IsDone always returns true, as the action has no function to perform in a base Action.
func (a *waitAction) IsDone() bool {
	return time.Now().After(a.endTime)
}
