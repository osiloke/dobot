package bots

import (
	"encoding/json"
	"errors"

	"github.com/osiloke/dostow-contrib/api"
)

// NewDostowActionStore create new store
func NewDostowActionStore(api *api.Client) *DostowActionStore {
	return &DostowActionStore{api}
}

// DostowActionStore an action store that uses dostow
type DostowActionStore struct {
	api *api.Client
}

type result struct {
	Data []*json.RawMessage `json:"data"`
}

func (d *DostowActionStore) query(action *Action, args map[string]interface{}) (*json.RawMessage, error) {
	query := substituteData(action.Query, args)
	raw, err := d.api.Store.Search(action.Store, api.QueryParams(query, 1, 0))
	if err == nil {
		var rows result
		err = json.Unmarshal(*raw, &rows)
		if err == nil {
			return rows.Data[0], nil
		}
	}
	return nil, err
}

// Do perform an action
func (d *DostowActionStore) Do(action *Action, args map[string]interface{}) (interface{}, error) {
	switch action.Method {
	case "GetOne":
		return d.query(action, args)
	case "UpdateOne":
		raw, err := d.query(action, args)
		if err != nil {
			return nil, err
		}
		result := struct {
			ID string `json:"id"`
		}{}
		err = json.Unmarshal(*raw, &result)
		if err != nil {
			return nil, err
		}
		return d.api.Store.Update(action.Store, result.ID, substituteData(action.Data, args))
	case "CreateOne":
		return d.api.Store.Create(action.Store, substituteData(action.Data, args))
	}
	return nil, errors.New("unknown store action")
}
