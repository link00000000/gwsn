//go:build unix

package assets

import _ "embed"

//go:embed tray.png
var TrayIcon []byte
