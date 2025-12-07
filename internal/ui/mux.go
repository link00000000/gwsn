package ui

import (
	"net/http"

	ui_index "github.com/link00000000/google-workspace-notify/internal/ui/index"
	ui_settings "github.com/link00000000/google-workspace-notify/internal/ui/settings"
)

func NewHandler() http.Handler {
	m := http.NewServeMux()

	m.HandleFunc("/", ui_index.Handle)
	m.HandleFunc("/settings", ui_settings.Handle)

	return m
}
