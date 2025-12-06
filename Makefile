assets := $(wildcard cmd/**/assets/*.png) $(wildcard cmd/**/assets/*.ico)
sources := $(wildcard src/**/*.go) $(wildcard cmd/**/*.go)
module := go.mod go.sum

.PHONY : all
all : build/google-workspace-notify/x86_64-linux/google-workspace-notify build/google-workspace-notify/x86_64-windows/google-workspace-notify.exe

build/google-workspace-notify/x86_64-linux/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=amd64 go build -o $@ ./cmd/google-workspace-notify

build/google-workspace-notify/x86_64-windows/google-workspace-notify.exe : $(sources) $(assets) $()
	GOOS=windows GOARCH=amd64 go build -o $@ ./cmd/google-workspace-notify
