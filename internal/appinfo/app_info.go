package appinfo

import (
	_ "embed"
	"encoding/json"
)

var (
	CurrentInfo = AppInfo{}
)

type AppInfo struct {
	Name      string         `json:"name"`
	Version   string         `json:"version"`
	Author    string         `json:"author"`
	Copyright string         `json:"copyright"`
	Meta      map[string]any `json:"meta"`
}

//go:embed app_info.json
var currInfoJson string

func init() {
	err := json.Unmarshal([]byte(currInfoJson), &CurrentInfo)
	if err != nil {
		panic(err)
	}
}
