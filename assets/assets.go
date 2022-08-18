package assets

import (
	_ "embed"
	"html/template"
	"strings"
)

//go:embed doc-tempalte.go.tpl
var docTemplate []byte
var DocTemplate string = string(docTemplate)

//go:embed show-template.go.tpl
var showTemplate []byte
var ShowTemplate string = string(showTemplate)

func GetDocTemplate() (*template.Template, error) {
	docTemplate, err := template.New("doc-template").Funcs(template.FuncMap{
		"mdlink": func(s *string) string {
			result := strings.ToLower(*s)
			result = strings.ReplaceAll(result, " ", "-")
			return result
		},
	}).Parse(DocTemplate)
	return docTemplate, err
}

func GetShowTemplate() (*template.Template, error) {
	showTemplate, err := template.New("doc-template").Funcs(template.FuncMap{
		"mdlink": func(s *string) string {
			result := strings.ToLower(*s)
			result = strings.ReplaceAll(result, " ", "-")
			return result
		},
	}).Parse(ShowTemplate)
	return showTemplate, err
}
