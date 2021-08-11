package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	instancesFile = "test/automated/ansible/group_vars/localhost/main.yml"
	inventory     = "test/automated/ansible/custom-instances.yml"
	colorArm64    = "\033[32m"
	colorAmd64    = "\033[34m"
	colorReset    = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

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
		fmt.Print(opts[i].Option())
		if _, ok := opts[i+len(opts)/2]; ok {
			fmt.Printf("        %s\n", opts[i+len(opts)/2].Option())
		}
	}

	fmt.Printf("Select one of numbers (or %s to quit): ", colorizeRed("q"))

	// get user input
	var userInput string

	fmt.Scanln(&userInput)

	if userInput == "q" {
		exit()
	}

	chosenAmiNumber, err := strconv.Atoi(userInput)

	if err != nil {
		panic(err)
	}

	// request for prefix
	provisionHostPrefix := randStringRunes(4)
	fmt.Printf("Enter a prefix for the boxes (empty for random): [%s] ", colorizeYellow(provisionHostPrefix))
	userInput = ""
	fmt.Scanln(&userInput)
	if userInput != "" {
		provisionHostPrefix = userInput
	}
	u, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	provisionHostPrefix = fmt.Sprintf("%s-%s", u.Username, provisionHostPrefix)
	// validate input

	// confirm
	fmt.Printf("Chosen AMI\n")
	fmt.Printf("Os: %s%s%s\n", colorPurple, opts[chosenAmiNumber].os, colorReset)
	fmt.Printf("Arch: %s%s%s\n", opts[chosenAmiNumber].arch.color(), opts[chosenAmiNumber].arch, colorReset)
	fmt.Printf("Prefix: %s%s%s\n", colorCyan, provisionHostPrefix, colorReset)
	fmt.Printf("\n")
	fmt.Printf("Is this right [(%s)es / (%s)o / (%s)uit]: ",colorizeGreen("y"),colorizeYellow("n"), colorizeRed("q"))
	userInput = ""
	fmt.Scanln(&userInput)

	if (userInput != "" && userInput != "yes" && userInput != "y") || userInput == "q" {
		exit()
	}

	// prepare ansible config (tmp list of hosts to create)
	fmt.Printf("Preparing config\n")

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

	executeAnsible()
}

func executeAnsible(){
	fmt.Printf("Executing Ansible\n")

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

	var errStdout, errStderr error

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		fmt.Println("Printing output: " + cmd.String())
		errStdout = copyAndCapture(os.Stdout, stdoutIn)

		wg.Done()
	}()
	go func() {
		fmt.Println("Printing error output: " + cmd.String())
		errStderr = copyAndCapture(os.Stderr, stderrIn)

		wg.Done()
	}()

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type architecture string

func (a architecture) color() string {
	if a == "amd64" {
		return colorAmd64
	}

	return colorArm64
}

type option struct {
	id       int
	arch     architecture
	os       string
	instance instanceDef
}

func (o option) FullName() string {
	return ""
}
func (o option) Option() string {
	optionFormat := "[%2d] %s%6s%s %18s"
	return fmt.Sprintf(optionFormat, o.id, o.arch.color(), o.arch, colorReset, o.os)
}

type options map[int]option

func generateOptions(yamlContent []byte) (options, error) {

	options := options{}

	groupVars := AnsibleGroupVars{}
	err := yaml.Unmarshal(yamlContent, &groupVars)
	if err != nil {
		log.Fatal(err.Error())
	}

	for i, instance := range groupVars.Instances {

		arch := instance.Name[:strings.Index(instance.Name, ":")]
		opSystem := instance.Name[strings.Index(instance.Name, ":")+1:]

		options[i] = option{
			id:       i,
			arch:     architecture(arch),
			os:       opSystem,
			instance: instance,
		}
	}

	return options, nil
}

func colorizeRed(s string) string{
	return fmt.Sprintf("%s%s%s", colorRed, s, colorReset)
}

func colorizeGreen(s string) string{
	return fmt.Sprintf("%s%s%s", colorGreen, s, colorReset)
}
func colorizeYellow(s string) string{
	return fmt.Sprintf("%s%s%s", colorYellow, s, colorReset)
}

func exit() {
	fmt.Println("Have a nice day!")
	os.Exit(0)
}

func copyAndCapture(w io.Writer, r io.Reader) error {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return err
		}
	}
}
