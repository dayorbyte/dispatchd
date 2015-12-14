package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
)

type Root struct {
	Amqp Amqp `xml:"amqp"`
}

type Amqp struct {
	Constants []Constant `xml:"constant"`
	Domains   []Domain   `xml:"domain"`
	Classes   []Class    `xml:"class"`
}

type Domain struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type Class struct {
	Methods []Method `xml:"method"`
	Name    string   `xml:"name,attr"`
	Handler string   `xml:"handler,attr"`
	Index   string   `xml:"index,attr"`
	Fields  []Field  `xml:"field"`
}

type Constant struct {
	Name  string `xml:"name,attr"`
	Value uint16 `xml:"value,attr"`
	Class string `xml:"class,attr"`
}

type Method struct {
	Name        string  `xml:"name,attr"`
	Synchronous string  `xml:"synchronous,attr"`
	Index       string  `xml:"index,attr"`
	Fields      []Field `xml:"field"`
}

type Field struct {
	Name   string `xml:"name,attr"`
	Domain string `xml:"domain,attr"`
}

var specFile string

func main() {
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
	// fmt.Println(root)
	for _, c := range root.Amqp.Classes {
		fmt.Println(c.Name)
		for _, m := range c.Fields {
			fmt.Print(m.Name, ",")
		}
		fmt.Println()
		for _, m := range c.Methods {
			fmt.Println("\t", m.Name)
		}
	}
}
