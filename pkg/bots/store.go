package bots

import (
	"encoding/json"
	"errors"

	"github.com/osiloke/dostow-contrib/api"
	"github.com/tidwall/gjson"
)

// NewDostowActionStore create new store
func NewDostowActionStore(api *api.Client) *DostowActionStore {
	return &DostowActionStore{api}
}

// DostowActionStore an action store that uses dostow
type DostowActionStore struct {
	api *api.Client
}

// Do perform an action
func (d *DostowActionStore) Do(action *Action, args map[string]interface{}) (interface{}, error) {
	var err error
	var raw *json.RawMessage
	switch action.Method {
	case "GetOne":
		query := map[string]interface{}{}
		j, _ := json.Marshal(args)
		jsonString := string(j)
		for k, v := range action.Query {
			query[k] = gjson.Get(jsonString, v.(string)).String()
		}
		raw, err = d.api.Store.Search(action.Store, api.QueryParams(query, 1, 0))
		if err == nil {
			return raw, nil
		}
		return nil, err
	}
	return nil, errors.New("not valid")
}
