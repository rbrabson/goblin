package action

import "time"

// An Action is used to run a piece of code
type Action interface {
	// Initialize is used to initialize the action.
	Initialize()

	// Execute runs the action
	Execute()

	// IsDone returns an indication as to whether the action has completed execution.
	IsDone() bool

	// AlongWith decorates this action with a set of actions to run parallel to it, ending when the last
	// action ends. Often more convenient/less-verbose than constructing a new ParallelAction explicitly.
	AlongWith(actions ...Action) Action

	// AndThen decorates this action with a set of actions to run after it in sequence. Often more
	// convenient/less-verbose than constructing a new SequentialAction explicitly.
	AndThen(actions ...Action) Action

	// DelayFor decorates this action with a delay before executing the action. It is often more
	// convenient/less-verbose than construction a new SequentialAction that contains a WaitAction and
	// this action.
	DelayFor(duration time.Duration) Action

	// DelayFor decorates this action with a timeout. It is often more convenient/less-verbose than
	// construction a new TimeoutAction for this action.
	WithTimeout(duration time.Duration) Action

	// WithWait decorates this action that waits for a delay once completing this action It is often more
	// convenient/less-verbose than construction a new SequentialAction that contains this action and a
	// WaitAction.
	WithWait(duration time.Duration) Action
}

// Base action implementation. This allows implementers of the Action interface to override those methods
// that they need to, such as `Initialize`, `Execute` and `Done`, without having to implement the rest.
type ActionBase struct{}

// Initialize is used to initialize the action.
func (a *ActionBase) Initialize() {}

// Execute runs the action. This is a no-op for a base Action.
func (a *ActionBase) Execute() {}

// IsDone always returns true, as the action has no function to perform in a base Action.
func (a *ActionBase) IsDone() bool { return true }

// AlongWith decorates this action with a set of actions to run parallel to it, ending when the last
// action ends. Often more convenient/less-verbose than constructing a new ParallelAction explicitly.
func (a *ActionBase) AlongWith(actions ...Action) Action {
	parallelAction := NewParallelAction(a)
	return parallelAction.AlongWith(actions...)
}

// AndThen decorates this action with a set of actions to run after it in sequence. Often more
// convenient/less-verbose than constructing a new SequentialAction explicitly.
func (a *ActionBase) AndThen(actions ...Action) Action {
	sequentialAction := NewSequentialAction(a)
	return sequentialAction.AndThen(actions...)
}

// DelayFor decorates this action with a delay before executing the action. It is often more
// convenient/less-verbose than construction a new SequentialAction that contains a WaitAction and
// this action.
func (a *ActionBase) DelayFor(duration time.Duration) Action { return a }

// DelayFor decorates this action with a timeout. It is often more convenient/less-verbose than
// construction a new TimeoutAction for this action.
func (a *ActionBase) WithTimeout(duration time.Duration) Action { return a }

// WithWait decorates this action that waits for a delay once completing this action It is often more
// convenient/less-verbose than construction a new SequentialAction that contains this action and a
// WaitAction.
func (a *ActionBase) WithWait(duration time.Duration) Action { return a }
