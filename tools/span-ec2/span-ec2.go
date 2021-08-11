package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
)

//const instancesFile = "../../test/automated/ansible/group_vars/localhost/main.yml"
const instancesFile = "/Users/rruizdegauna/src/nr-pub/infrastructure-agent/test/automated/ansible/group_vars/localhost/main.yml"
const inventory = "test/automated/ansible/custom-instances.yml"

type AnsibleGroupVars struct {
	ProvisionHostPrefix string        `yaml:"provision_host_prefix"`
	Instances           []instanceDef `yaml:"instances"`
}

type instanceDef struct {
	Ami               string `yaml:"ami"`
	InstanceType      string `yaml:"type"`
	Name              string `yaml:"name"`
	Username          string `yaml:"username"`
	PythonInterpreter string `yaml:"python_interpreter"`
	LaunchTemplate    string `yaml:"launch_template"`
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func main() {
	rand.Seed(time.Now().UnixNano())

	yamlFile, err := ioutil.ReadFile(instancesFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	opts, err := generateOptions(yamlFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	for i := 0; i < len(opts)/2+1; i++ {
		fmt.Print(opts[i].name)
		if _, ok := opts[i+len(opts)/2]; ok {
			fmt.Printf("        %s\n", opts[i+len(opts)/2].name)
		}
	}

	fmt.Print("Select one of numbers (or q to quit): ")

	// get user input
	var userInput string

	fmt.Scanln(&userInput)

	if userInput == "q" {
		fmt.Println("Have a nice day!")
		os.Exit(0)
	}

	chosenAmiNumber, err := strconv.Atoi(userInput)

	if err != nil {
		panic(err)
	}

	// request for prefix
	provisionHostPrefix := randStringRunes(4)
	fmt.Printf("Enter a prefix for the boxes (empty for random): [%s] ", provisionHostPrefix)
	userInput = ""
	fmt.Scanln(&userInput)
	if userInput != "" {
		provisionHostPrefix = userInput
	}
	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	username := user.Username
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	provisionHostPrefix = fmt.Sprintf("%s-%s-%s", username, hostname, provisionHostPrefix)
	// validate input

	// confirm
	fmt.Printf("Chosen AMI\n")
	fmt.Printf("Os: %s\n", opts[chosenAmiNumber].os)
	fmt.Printf("Arch: %s\n", opts[chosenAmiNumber].arch)
	fmt.Printf("Prefix: %s\n", provisionHostPrefix)
	fmt.Printf("\n")
	fmt.Printf("Is this right [(y)es / (n)o / (q)uit]: ")
	userInput = ""
	fmt.Scanln(&userInput)

	if (userInput != "yes" && userInput != "y") || userInput == "q" {
		os.Exit(0)
	}

	// prepare ansible config (tmp list of hosts to create)
	fmt.Printf("Preparing config for %s\n", opts[chosenAmiNumber].name)

	newConfig := AnsibleGroupVars{}
	newConfig.ProvisionHostPrefix = provisionHostPrefix
	newConfig.Instances = append(newConfig.Instances, opts[chosenAmiNumber].instance)
	newConfigByte, err := yaml.Marshal(newConfig)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(inventory, newConfigByte, 0644)
	if err != nil {
		panic(err)
	}

	// execute ansible
	fmt.Printf("Executing Ansible for %s\n", opts[chosenAmiNumber].name)

	curPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(
		"ansible-playbook",
		"-i", path.Join(curPath, "test/automated/ansible/inventory.local"),
		"--extra-vars", "@"+path.Join(curPath, inventory),
		path.Join(curPath, "test/automated/ansible/provision.yml"),
	)

	fmt.Println("Executing command: " + cmd.String())

	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	if err != nil {
		log.Fatal(errOut.String())
	}
	fmt.Printf("in all caps: %q\n", out.String())

}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
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
