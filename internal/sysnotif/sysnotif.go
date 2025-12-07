package sysnotif

import (
	"github.com/gen2brain/beeep"
)

func init() {
	beeep.AppName = "Google Workspace Notify"
}

func ShowNotification(title, message string) {
	beeep.Notify(title, message, "")
}

func ShowNotificationWithIcon(title, message string, icon []byte) {
	beeep.Notify(title, message, icon)
}
