package bots

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

func substituteData(vars, args map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{}
	j, _ := json.Marshal(args)
	jsonString := string(j)
	for k, v := range vars {
		data[k] = gjson.Get(jsonString, v.(string)).String()
	}
	return data
}
