package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const instancesFile = "../../test/automated/ansible/group_vars/localhost/main.yml"

type AnsibleGroupVars struct {
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

	options, err := generateOptions(yamlFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	for i := 0; i < len(options)/2+1; i++ {
		fmt.Print(options[i].name)
		if _, ok := options[i+len(options)/2]; ok {
			fmt.Printf("        %s\n", options[i+len(options)/2].name)
		}
	}

	fmt.Print("Select one of numbers: ")

	// get user input
	var userInput string

	fmt.Scanln(&userInput)

	chosenAmiNumber, err := strconv.Atoi(userInput)

	if err != nil {
		panic(err)
	}

	// validate input

	// confirm
	fmt.Printf("Chosen AMI %d - %s\nIs this right [(y)es/(n)o]: ", chosenAmiNumber, options[chosenAmiNumber].name)

	fmt.Scanln(&userInput)

	if userInput != "yes" && userInput != "y" {
		os.Exit(0)
	}

	// prepare ansible config (tmp list of hosts to create)
	fmt.Printf("Preparing config for %s\n", options[chosenAmiNumber].name)

	newConfig := AnsibleGroupVars{}
	newConfig.Instances = append(newConfig.Instances, options[chosenAmiNumber].instance)
	newConfigByte, err := yaml.Marshal(newConfig)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("./instances.yml", newConfigByte, 0644)
	if err != nil {
		panic(err)
	}

	// execute ansible
	fmt.Printf("Executing Ansible for %s\n", options[chosenAmiNumber].name)

	//cmd := exec.Command("ansible-playbook", "release.yml", "--extra-vars", "@")
	//cmd.Stdin = strings.NewReader("some input")
	//var out bytes.Buffer
	//cmd.Stdout = &out
	//err := cmd.Run()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("in all caps: %q\n", out.String())

}

type option struct {
	id       int
	name     string
	arch     string
	os       string
	instance instanceDef
}

type options map[int]option

func generateOptions(yamlContent []byte) (options, error) {

	options := options{}

	groupVars := AnsibleGroupVars{}
	err := yaml.Unmarshal(yamlContent, &groupVars)
	if err != nil {
		log.Fatal(err.Error())
	}

	optionFormat := "[%2d] %s%6s%s %18s"

	colorGreen := "\033[32m"
	colorBlue := "\033[34m"
	colorReset := "\033[0m"

	for i, instance := range groupVars.Instances {

		arch := instance.Name[:strings.Index(instance.Name, ":")]
		operationSystem := instance.Name[strings.Index(instance.Name, ":")+1:]

		osColor := colorGreen

		if arch == "amd64" {
			osColor = colorBlue
		}

		options[i] = option{
			id:       i,
			arch:     arch,
			os:       operationSystem,
			name:     fmt.Sprintf(optionFormat, i, osColor, arch, colorReset, operationSystem),
			instance: instance,
		}
	}

	return options, nil
}
