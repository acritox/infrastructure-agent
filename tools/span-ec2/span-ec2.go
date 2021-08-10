package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

const instancesFile = "../../test/automated/ansible/group_vars/localhost/main.yml"

type Config struct {
	Instances []instanceDef `yaml:"instances"`
}

type instanceDef struct {
	Ami               string `yaml:"ami"`
	InstanceType      string `yaml:"type"`
	Name              string `yaml:"name"`
	Username          string `yaml:"username"`
	PythonInterpreter string `yaml:"python_interpreter"`
	LaunchTemplate    string `yaml:"launch_template"`
}

func main() {
	yamlFile, err := ioutil.ReadFile(instancesFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	config := Config{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	optionFormat := "[%2d] %25s"

	fmt.Printf("Choose AMI to create\n\n")
	for i, instance := range config.Instances{
		fmt.Println(fmt.Sprintf(optionFormat,i, instance.Name))
	}

	fmt.Print("Select one of numbers: ")

	// get user input
	var userInput string

	fmt.Scanln(&userInput)

	chosenAmiNumber, err := strconv.Atoi(userInput)

	if err != nil{
		panic(err)
	}

	// validate input

	// confirm
	fmt.Printf("Chosen AMI %d - %s\nIs this right [(y)es/(n)o]: ", chosenAmiNumber, config.Instances[chosenAmiNumber].Name)

	fmt.Scanln(&userInput)

	if userInput != "yes" && userInput != "y"{
		os.Exit(0)
	}

	// prepare ansible config (tmp list of hosts to create)
	fmt.Printf("Preparing config for %s\n", config.Instances[chosenAmiNumber].Name)


	// execute ansible
	fmt.Printf("Executing Ansible for %s\n", config.Instances[chosenAmiNumber].Name)

}


