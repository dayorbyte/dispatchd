package server

import (
	"text/template"
)

var protoTemplate *template.Template
var readWriteTemplate *template.Template

func init() {
	//
	//         PROTOCOL BUFFER TEMPLATE
	//
	t, err := template.New("proto").Parse(`package amqp;

import "github.com/gogo/protobuf/gogoproto/gogo.proto";
import "github.com/jeffjenkins/dispatchd/amqp/amqp.proto";

option (gogoproto.marshaler_all) = true;
option (gogoproto.sizer_all) = true;
option (gogoproto.unmarshaler_all) = true;
{{range .Amqp.Classes}}
{{if .Fields}}
message {{.NormName}}ContentHeaderProperties {
  {{range $index, $field := .Fields}}
  optional {{.ProtoType}} {{.ProtoName}} = {{.ProtoIndex}}{{.Options}};{{end}}
}
{{end}}
{{range .Methods}}
message {{.NormName}} {
  option (gogoproto.goproto_unrecognized) = false;
  option (gogoproto.goproto_getters) = false;
  {{range $index, $field := .Fields}}
  optional {{.ProtoType}} {{.ProtoName}} = {{.ProtoIndex}}{{.Options}};{{end}}
}
{{end}}
{{end}}


  `)
	if err != nil {
		panic(err.Error())
	}
	protoTemplate = t
}
func init() {
	//
	//         READ/WRITE TEMPLATE
	//
	t, err := template.New("proto").Parse(`
package amqp

import (
  "io"
  "errors"
  "fmt"
)

{{range $class := .Amqp.Classes}}
var ClassId{{.NormName}} uint16 = {{.Index}}
{{if .Fields}}{{range $index, $field := .Fields}}
var Mask{{.NormName}} uint16 = 0x{{.MaskIndex}};{{end}}
{{end}}
{{range .Methods}}
// ************************
// {{.NormName}}
// ************************
var MethodId{{.NormName}} uint16 = {{.Index}};

func (f *{{.NormName}}) MethodIdentifier() (uint16, uint16) {
  return {{$class.Index}}, {{.Index}}
}

func (f *{{.NormName}}) MethodName() string {
  return "{{.NormName}}"
}

func (f *{{.NormName}}) FrameType() byte {
  return 1
}

// Reader
func (f *{{.NormName}}) Read(reader io.Reader) (err error) {
{{range $index, $field := .Fields}}
  {{if eq .BitOffset -1}}
  f.{{.NormName}}, err = Read{{.Serializer}}(reader)
  if err != nil {
    return errors.New("Error reading field {{.NormName}}: " + err.Error())
  }
  {{else}}
    {{if eq .BitOffset 0}}
  bits, err := ReadOctet(reader)
  if err != nil {
    return errors.New("Error reading bit fields" + err.Error())
  }
    {{end}}
  f.{{.NormName}} = (bits & (1 << {{.BitOffset}}) > 0)
  {{end}}
{{end}}{{/* end fields */}}
  return
}

// Writer
func (f *{{.NormName}}) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, {{$class.Index}}); err != nil {
    return err
  }
  if err = WriteShort(writer, {{.Index}}); err != nil {
    return err
  }

  {{range $index, $field := .Fields}}
  {{if eq .GoType "bool"}}
      {{if eq .BitOffset 0}}
  var bits byte
      {{end}}
  if f.{{.NormName}} {
    bits |= 1 << {{.BitOffset}}
  }
  {{else}}{{/* else go type is not bool */}}
    {{if .PreviousBit}}
  err = WriteOctet(writer, bits)
  if err != nil {
    return errors.New("Error writing bit fields")
  }
    {{end}}
  err = Write{{.Serializer}}(writer, f.{{.NormName}})
  if err != nil {
    return errors.New("Error writing field {{.NormName}}")
  }
  {{end}}{{/* end go type bool */}}
  {{end}}{{/* end fields */}}
  {{if .BitsAtEnd}}
  err = WriteOctet(writer, bits)
  if err != nil {
    return errors.New("Error writing bit fields")
  }
  {{end}}
  return
}

{{end}}{{/* end methods */}}
{{end}}{{/* end amqp.classes */}}

// ********************************
// METHOD READER
// ********************************

func ReadMethod(reader io.Reader) (MethodFrame, error) {
  classIndex, err := ReadShort(reader)
  if err != nil {
    return nil, err
  }
  methodIndex, err := ReadShort(reader)
  if err != nil {
    return nil, err
  }
  switch {
{{range $class := .Amqp.Classes}}
  case classIndex == {{.Index}}:
    switch {
{{range .Methods}}
      case methodIndex == {{.Index}}:
        var method = &{{.NormName}}{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
{{end}}
    }
{{end}}
  }

  return nil, errors.New(fmt.Sprintf("Bad method or class Id! classId: %d, methodIndex: %d", classIndex, methodIndex))
}
  `)
	if err != nil {
		panic(err.Error())
	}
	readWriteTemplate = t
}
