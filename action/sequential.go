package action

// sequentialAction is an Action that runs a slice of actions in sequence.
type sequentialAction struct {
	ActionBase
	actions     []Action
	actionIndex int
}

// NewSequentialAction creates an action that returns the list of provided actions in sequence.
func NewSequentialAction(actions ...Action) *sequentialAction {
	action := &sequentialAction{
		actions: actions,
	}
	return action
}

// Initialize is used to initialize the action.
func (a *sequentialAction) Initialize() {
	if len(a.actions) > 0 {
		a.actions[0].Initialize()
	}
}

// Execute runs the current action in the sequence. Once an action in the sequence has completed
// execution, then the next action in the sequence will be executed.
func (a *sequentialAction) Execute() {
	// If the current action has completed, move to the next action in the list
	if a.actionIndex < len(a.actions) {
		action := a.actions[a.actionIndex]
		if action.IsFinished() {
			a.actionIndex++
			if a.actionIndex < len(a.actions) {
				a.actions[a.actionIndex].Initialize()
			}
		}
	}
	// Already reached the end of the actions to run in paralle, so this is a no-op
	if a.actionIndex >= len(a.actions) {
		return
	}
	//
	action := a.actions[a.actionIndex]
	action.Execute()
}

// IsDone returns `true` once the last action in the sequence has completed execution.
// Otherwise, it returns `false`.
func (a *sequentialAction) IsDone() bool {
	// If not executing the last action, then the list of sequentail actions are not done executing
	if a.actionIndex < len(a.actions)-1 {
		return false
	}
	// Executing the last action, so just check to see if it is done executing
	action := a.actions[a.actionIndex]
	return action.IsFinished()
}
