//go:build windows

package assets

import _ "embed"

//go:embed tray.ico
var TrayIcon []byte
