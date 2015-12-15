package main

import (
	"text/template"
)

var protoTemplate *template.Template

func init() {
	t, err := template.New("proto").Parse(`package amqp;

import "github.com/gogo/protobuf/gogoproto/gogo.proto";
import "amqp/amqp.proto";

option (gogoproto.marshaler_all) = true;
option (gogoproto.sizer_all) = true;
option (gogoproto.unmarshaler_all) = true;
{{range .Amqp.Classes}}
{{if .Fields}}
message {{.NormName}}ContentHeaderProperties {
  {{range $index, $field := .Fields}}
  optional{{.ProtoType}} {{.ProtoName}} = {{.ProtoIndex}}{{.Options}};{{end}}
}
{{end}}
{{range .Methods}}
message {{.NormName}} {
  option (gogoproto.goproto_unrecognized) = false;
  option (gogoproto.goproto_getters) = false;
  {{range $index, $field := .Fields}}
  optional{{.ProtoType}} {{.ProtoName}} = {{.ProtoIndex}}{{.Options}};{{end}}
}
{{end}}
{{end}}


  `)
	if err != nil {
		panic(err.Error())
	}
	protoTemplate = t
}
