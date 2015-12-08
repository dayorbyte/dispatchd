package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
)

var amqpPort int
var amqpPortDefault = 5672
var debugPort int
var debugPortDefault = 6060
var adminPort int
var adminPortDefault = 8080
var persistDir string
var persistDirDefault = "/data/dispatchd/"
var configFile string
var configFileDefault = ""

func init() {
	flag.IntVar(&amqpPort, "amqp-port", 0, "Port for amqp protocol messages. Default: 5672")
	flag.IntVar(&debugPort, "debug-port", 0, "Port for the golang debug handlers. Default: 6060")
	flag.IntVar(&adminPort, "admin-port", 0, "Port for admin server. Default: 8080")
	flag.StringVar(&persistDir, "persist-dir", "", "Directory for the server and message database files. Default: /data/dispatchd/")
	flag.StringVar(
		&configFile,
		"config-file",
		"",
		"Directory for the server and message database files. Default: do not read a config file",
	)
}

func configure() map[string]interface{} {
	// TODO: It's no great that this is manual. I should make/find a small library
	//       to automate this.
	var config = make(map[string]interface{})
	if configFile != "" {
		config = parseConfigFile(configFile)
	}
	configureIntParam(&amqpPort, amqpPortDefault, "amqp-port", config)
	configureIntParam(&debugPort, debugPortDefault, "debug-port", config)
	configureIntParam(&adminPort, adminPortDefault, "admin-port", config)
	configureStringParam(&persistDir, persistDirDefault, "persist-dir", config)
	return config
}

func configureIntParam(param *int, defaultValue int, configName string, config map[string]interface{}) {
	if *param != 0 {
		return
	}
	if len(configName) != 0 {
		value, ok := config[configName]
		if ok {
			*param = int(value.(float64))
			return
		}
	}
	*param = defaultValue
}

func configureStringParam(param *string, defaultValue string, configName string, config map[string]interface{}) {
	if *param != "" {
		return
	}
	if len(configName) != 0 {
		value, ok := config[configName]
		if ok {
			*param = value.(string)
			return
		}
	}
	*param = defaultValue
}

func parseConfigFile(path string) map[string]interface{} {
	ret := make(map[string]interface{})
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Could not read config file: " + err.Error())
	}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		panic("Could not parse config file: " + err.Error())
	}
	return ret
}
