package action

// parallelAction is an Action that runs a set of Actions at the same time
type parallelAction struct {
	ActionBase
	actions []Action
}

// NewParallelAction creates a new action that runs all the provided actions at the same time.
func NewParallelAction(actions ...Action) *parallelAction {
	action := &parallelAction{
		actions: actions,
	}
	return action
}

// Initialize is used to initialize the action. This calls the `Initialize` method for each
// parallel action.
func (a *parallelAction) Initialize() {
	for _, action := range a.actions {
		action.Initialize()
	}
}

// Execute runs the action. This calls the `Execute` method for each parallel action that has
// not completed processing (i.e., the action's `IsDone` method returns `faslse`).
func (a *parallelAction) Execute() {
	for _, action := range a.actions {
		if !action.IsFinished() {
			action.Execute()
		}
	}
}

// End is run when an action completes. The action may run to completion or be interrupted, as defined
// by the provided parameter.
func (a *parallelAction) End(interrupted bool) {}

// IsFinished returns an indication as to whether the action has completed execution.
func (a *parallelAction) IsFinished() bool {
	isFinished := true
	for _, action := range a.actions {
		if !action.IsFinished() {
			isFinished = false
			break
		}
	}
	return isFinished
}
