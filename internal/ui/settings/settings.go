package settings

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed settings.html
var f embed.FS

type viewModel struct {
}

func Handle(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(f, "settings.html"))
	vm := viewModel{}
	tmpl.Execute(w, vm)
}
