package action

// sequentialAction is an Action that runs a slice of actions in sequence.
type sequentialAction struct {
	ActionBase
	Actions     []Action
	ActionIndex int
}

// NewSequentialAction creates an action that returns the list of provided actions in sequence.
func NewSequentialAction(actions ...Action) *sequentialAction {
	action := &sequentialAction{
		Actions: actions,
	}
	return action
}

// Initialize is used to initialize the action.
func (a *sequentialAction) Initialize() {
	if len(a.Actions) > 0 {
		a.Actions[0].Initialize()
	}
}

// Execute runs the current action in the sequence. Once an action in the sequence has completed
// execution, then the next action in the sequence will be executed.
func (a *sequentialAction) Execute() {
	// If the current action has completed, move to the next action in the list
	if a.ActionIndex < len(a.Actions) {
		action := a.Actions[a.ActionIndex]
		if action.IsFinished() {
			a.ActionIndex++
			if a.ActionIndex < len(a.Actions) {
				a.Actions[a.ActionIndex].Initialize()
			}
		}
	}
	// Already reached the end of the actions to run in paralle, so this is a no-op
	if a.ActionIndex >= len(a.Actions) {
		return
	}
	//
	action := a.Actions[a.ActionIndex]
	action.Execute()
}

// IsDone returns `true` once the last action in the sequence has completed execution.
// Otherwise, it returns `false`.
func (a *sequentialAction) IsDone() bool {
	// If not executing the last action, then the list of sequentail actions are not done executing
	if a.ActionIndex < len(a.Actions)-1 {
		return false
	}
	// Executing the last action, so just check to see if it is done executing
	action := a.Actions[a.ActionIndex]
	return action.IsFinished()
}
