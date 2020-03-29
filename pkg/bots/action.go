package bots

// ActionStore used to get information for action
type ActionStore interface {
	Do(action *Action, args map[string]interface{}) (interface{}, error)
}

// Action an action that can be performed
type Action struct {
	Method string                 `json:"method"`
	Store  string                 `json:"store"`
	Query  map[string]interface{} `json:"query"`
}

// Actions defines certain actions a bot can process
type Actions map[string]*Action

// DoAction does an action
func DoAction(name string, args map[string]interface{}, actions Actions, store ActionStore) (interface{}, error) {
	return store.Do(actions[name], args)
}
