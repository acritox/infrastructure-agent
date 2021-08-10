package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

const instancesFile = "../../test/automated/ansible/group_vars/localhost/main.yml"

type instanceDef struct {
	ami               string `yaml:"ami"`
	instanceType      string `yaml:"type"`
	name              string `yaml:"name"`
	username          string `yaml:"username"`
	pythonInterpreter string `yaml:"python_interpreter"`
	launchTemplate    string `yaml:"launch_template"`
}

func main() {
	yamlFile, err := ioutil.ReadFile(instancesFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	var instances []instanceDef
	err = yaml.Unmarshal(yamlFile, &instances)
	if err != nil {
		log.Fatal(err.Error())
	}



}
