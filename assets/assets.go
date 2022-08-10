package assets

import _ "embed"

//go:embed doc-tempalte.go.tpl
var docTemplate []byte
var DocTemplate string = string(docTemplate)
