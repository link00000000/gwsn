assets := $(wildcard internal/**/assets/*.png) $(wildcard internal/**/assets/*.ico)
entrypoint := main.go
sources := $(entrypoint) $(wildcard internal/**/*.go)
module := go.mod go.sum

default : build/google-workspace-notify/x86_64-windows/google-workspace-notify.exe

# Linux builds
build/google-workspace-notify/x86_64-linux/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=amd64 go build -o $@ $(entrypoint)

build/google-workspace-notify/arm64-linux/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=arm64 go build -o $@ $(entrypoint)

build/google-workspace-notify/386-linux/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=386 go build -o $@ $(entrypoint)

# macOS builds
build/google-workspace-notify/x86_64-darwin/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=darwin GOARCH=amd64 go build -o $@ $(entrypoint)

build/google-workspace-notify/arm64-darwin/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=darwin GOARCH=arm64 go build -o $@ $(entrypoint)

# FreeBSD builds
build/google-workspace-notify/x86_64-freebsd/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=freebsd GOARCH=amd64 go build -o $@ $(entrypoint)

# Windows builds
build/google-workspace-notify/x86_64-windows/google-workspace-notify.exe : $(sources) $(assets) $(module)
	GOOS=windows GOARCH=amd64 go build -o $@ $(entrypoint)

build/google-workspace-notify/386-windows/google-workspace-notify.exe : $(sources) $(assets) $(module)
	GOOS=windows GOARCH=386 go build -o $@ $(entrypoint)
