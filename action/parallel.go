package action

// parallelAction is an Action that runs a set of Actions at the same time
type parallelAction struct {
	ActionBase
	Actions []Action
}

// NewParallelAction creates a new action that runs all the provided actions at the same time.
func NewParallelAction(actions ...Action) Action {
	action := &parallelAction{
		Actions: actions,
	}
	return action
}

// Initialize is used to initialize the action. This calls the `Initialize` method for each
// parallel action.
func (a *parallelAction) Initialize() {
	for _, action := range a.Actions {
		action.Initialize()
	}
}

// Execute runs the action. This calls the `Execute` method for each parallel action that has
// not completed processing (i.e., the action's `IsDone` method returns `faslse`).
func (a *parallelAction) Execute() {
	for _, action := range a.Actions {
		if !action.IsDone() {
			action.Execute()
		}
	}
}

// IsDone returns an indication as to whether the action has completed execution.
func (a *parallelAction) IsDone() bool {
	isDone := true
	for _, action := range a.Actions {
		if !action.IsDone() {
			isDone = false
			break
		}
	}
	return isDone
}
