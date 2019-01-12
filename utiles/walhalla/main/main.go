package main

import (
	"flag"
	"html/template"
	"path"
	"strings"
	"sync"

	"github.com/bomkvilt/tech-db-ap/utiles/walhalla/assets"
)

func generateTemplates() *template.Template {
	templates := []string{
		"templates/main.tmpl",
	}
	data, err := assets.Asset(templates[0])
	check(err)

	return template.Must(template.New(path.Base(templates[0])).
		Funcs(template.FuncMap{
			"join": func(data []string) string {
				return strings.Join(data, ",")
			},
		}).Parse(string(data)))
}

func main() {
	flag.Parse()
	template := generateTemplates()

	wg := sync.WaitGroup{}
	for _, pack := range flag.Args() {
		wg.Add(1)
		go func(pack string) {
			defer wg.Done()
			walhalla(template, pack)
		}(pack)
	}
	wg.Wait()
}
