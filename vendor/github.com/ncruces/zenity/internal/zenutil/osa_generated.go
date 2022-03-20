// Code generated by zenity; DO NOT EDIT.
// +build darwin

package zenutil

import (
	"encoding/json"
	"text/template"
)

var scripts = template.Must(template.New("").Funcs(template.FuncMap{"json": func(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}}).Parse(`
{{define "color" -}}
var app=Application.currentApplication()
app.includeStandardAdditions=true
app.activate()
var res=app.chooseColor({defaultColor:{{json .}}})
{"rgb("+res.map(x=>Math.round(x*255))+")"}
{{- end}}
{{define "dialog" -}}
var app=Application.currentApplication()
app.includeStandardAdditions=true
app.activate()
ObjC.import("stdlib")
ObjC.import("stdio")
var res=app.{{.Operation}}({{json .Text}},{{json .Options}})
if(res.gaveUp){$.exit(5)}
if(res.buttonReturned==={{json .Extra}}){$.puts(res.buttonReturned)
$.exit(1)}
res.textReturned
{{- end}}
{{define "file" -}}
var app=Application.currentApplication()
app.includeStandardAdditions=true
app.activate()
var res=app.{{.Operation}}({{json .Options}})
if(Array.isArray(res)){res.join({{json .Separator}})}else{res.toString()}
{{- end}}
{{define "list" -}}
var app=Application.currentApplication()
app.includeStandardAdditions=true
var res=app.chooseFromList({{json .Items}},{{json .Options}})
if(res.length!==0)res.join({{json .Separator}})
{{- end}}
{{define "notify" -}}
var app=Application.currentApplication()
app.includeStandardAdditions=true
void app.displayNotification({{json .Text}},{{json .Options}})
{{- end}}
{{define "progress" -}}
var app=Application.currentApplication()
app.includeStandardAdditions=true
app.activate()
ObjC.import('stdlib')
ObjC.import('readline')
{{- if .Total}}
Progress.totalUnitCount={{.Total}}
{{- end}}
{{- if .Description}}
Progress.description={{json .Description}}
{{- end}}
while(true){var s
try{s=$.readline('')}catch(e){if(e.errorNumber===-128)$.exit(1)
break}
if(s.indexOf('#')===0){Progress.additionalDescription=s.slice(1)
continue}
var i=parseInt(s)
if(i>=0&&Progress.totalUnitCount>0){Progress.completedUnitCount=i
continue}}
{{- end}}`))