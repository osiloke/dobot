package bots

import (
	"errors"

	"github.com/valyala/fasttemplate"
)

// ActionStore used to get information for action
type ActionStore interface {
	Do(action *Action, args map[string]interface{}) (interface{}, error)
}

// Error error message template
type Error struct {
	Message string                 `json:"msg"`
	Data    map[string]interface{} `json:"data"`
}

// Action an action that can be performed
type Action struct {
	Method string                 `json:"method"`
	Store  string                 `json:"store"`
	Query  map[string]interface{} `json:"query"`
	Data   map[string]interface{} `json:"data"`
	Error  *Error                 `json:"error"`
	Next   *Action 				  `json:"next"`
}

// Actions defines certain actions a bot can process
type Actions map[string]*Action

// DoAction does an action
func DoAction(name string, args map[string]interface{}, actions Actions, store ActionStore) (interface{}, error) {
	action := actions[name]
	if action == nil {
		return nil, errors.New("unknown action")
	}
	res, err := store.Do(action, args)
	if err != nil {
		if action.Error != nil {
			t := fasttemplate.New(action.Error.Message, "{{", "}}")
			s := t.ExecuteString(substituteData(action.Error.Data, args))
			return nil, errors.New(s)
		}
	}
	return res, err
}
