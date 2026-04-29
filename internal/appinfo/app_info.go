package appinfo

var (
	CurrentInfo = AppInfo{
		Name:      "kzhikcn-api",
		Version:   "0.0.0-alpha",
		Author:    "kzhik",
		Copyright: "© 2025 kzhik All rights reserved.",
		Meta:      map[string]any{},
	}
)

type AppInfo struct {
	Name      string
	Version   string
	Author    string
	Copyright string
	Meta      map[string]any
}
