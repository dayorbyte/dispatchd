package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var amqpToGo = map[string]string{
	"bit":       "bool",
	"octet":     "byte",
	"short":     "uint16",
	"long":      "uint32",
	"longlong":  "uint64",
	"table":     "Table",
	"timestamp": "uint64",
	"shortstr":  "string",
	"longstr":   "[]byte",
}

var amqpToProto = map[string]string{
	"bit":       "bool",
	"octet":     "uint32",
	"short":     "uint32",
	"long":      "uint32",
	"longlong":  "uint64",
	"table":     "Table",
	"timestamp": "uint64",
	"shortstr":  "string",
	"longstr":   "bytes",
}

type Root struct {
	Amqp Amqp `xml:"amqp"`
}

type Amqp struct {
	Constants []*Constant `xml:"constant"`
	Domains   []*Domain   `xml:"domain"`
	Classes   []*Class    `xml:"class"`
}

type Domain struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type Class struct {
	Methods  []*Method `xml:"method"`
	Name     string    `xml:"name,attr"`
	NormName string
	Handler  string   `xml:"handler,attr"`
	Index    string   `xml:"index,attr"`
	Fields   []*Field `xml:"field"`
}

type Constant struct {
	Name     string `xml:"name,attr"`
	NormName string
	Value    uint16 `xml:"value,attr"`
	Class    string `xml:"class,attr"`
}

type Method struct {
	Name        string `xml:"name,attr"`
	NormName    string
	Synchronous string   `xml:"synchronous,attr"`
	Index       string   `xml:"index,attr"`
	Fields      []*Field `xml:"field"`
}

type Field struct {
	Name       string `xml:"name,attr"`
	Domain     string `xml:"domain,attr"`
	NormName   string
	ProtoType  string
	ProtoIndex int
	ProtoName  string
	Options    string
}

var specFile string

func main() {
	fmt.Println(protoTemplate)
	flag.StringVar(&specFile, "spec", "", "Spec XML file")
	flag.Parse()
	var bytes, err = ioutil.ReadFile(specFile)
	if err != nil {
		panic(err)
	}
	var root Root
	err = xml.Unmarshal(bytes, &root.Amqp)
	if err != nil {
		panic(err.Error())
	}
	transform(&root)
	protoTemplate.Execute(os.Stdout, &root)
}

func transform(r *Root) {
	transformConstants(r.Amqp.Constants)
	domainTypes := transformDomains(r.Amqp.Domains)
	transformClasses(r.Amqp.Classes, domainTypes)
}

func transformConstants(cs []*Constant) {
	for _, c := range cs {
		c.NormName = normalizeName(c.Name)
	}
}

func transformClasses(cs []*Class, domainTypes map[string]string) {
	for _, c := range cs {
		c.NormName = normalizeName(c.Name)
		transformFields(c.Fields, domainTypes, true, 1)
		transformMethods(c.NormName, c.Methods, domainTypes)
	}
}

func transformDomains(ds []*Domain) map[string]string {
	domainTypes := make(map[string]string)
	for _, d := range ds {
		domainTypes[d.Name] = d.Type
	}
	return domainTypes
}

func transformMethods(className string, ms []*Method, domainTypes map[string]string) {
	for _, m := range ms {
		m.NormName = className + normalizeName(m.Name)
		transformFields(m.Fields, domainTypes, false, 1)
	}
}

func transformFields(fs []*Field, domainTypes map[string]string, nullable bool, offset int) {
	for index, f := range fs {
		f.NormName = normalizeName(f.Name)
		f.ProtoName = protoName(f.Name)
		f.Options = fieldOptions(f, domainTypes, nullable)
		f.ProtoIndex = index + offset
	}
}

func fieldOptions(f *Field, domainTypes map[string]string, nullable bool) string {
	var options = make([]string, 0)
	domain := f.Domain
	amqpType := domainTypes[domain]
	goType := amqpToGo[amqpType]
	protoType := amqpToProto[amqpType]
	if goType != protoType && protoType != "bytes" {
		options = append(options, fmt.Sprintf("(gogoproto.casttype) = \"%s\"", goType))
	}
	if protoType != "Table" && protoType != "bytes" && !nullable {
		options = append(options, "(gogoproto.nullable) = false")
	}
	if len(options) > 0 {
		return " [" + strings.Join(options, ", ") + "]"
	}
	return ""

}

func normalizeName(s string) string {
	parts := strings.Split(s, "-")
	ret := ""
	for _, p := range parts {
		ret += upperFirst(p)
	}
	return ret
}
func protoName(s string) string {
	return strings.Join(strings.Split(s, "-"), "_")
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}

	return string(bytes.ToUpper([]byte(s[0:1]))) + s[1:]
}
