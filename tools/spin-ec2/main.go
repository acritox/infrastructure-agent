package main

import (
	"fmt"
	"log"
	"math/rand"
	"os/user"
	"strconv"
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func main() {
	rand.Seed(time.Now().UnixNano())

	ansibleGroupVars, err := readAnsibleGroupVars()
	if err != nil {
		log.Fatal(err.Error())
	}

	opts, err := generateOptions(*ansibleGroupVars)
	if err != nil {
		log.Fatal(err.Error())
	}

	opts.print()

	chosenAmiNumber, err := strconv.Atoi(askUser(fmt.Sprintf("Select one of numbers (or %s to quit): ", colorizeRed("q"))))

	if err != nil {
		panic(err)
	}

	// request for prefix
	provisionHostPrefix := randStringRunes(4)

	userProvisionHostPrefix := askUser(fmt.Sprintf("Enter a prefix for the boxes (empty for random): [%s] ", colorizeYellow(provisionHostPrefix)))
	if userProvisionHostPrefix != "" {
		provisionHostPrefix = userProvisionHostPrefix
	}

	u, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	provisionHostPrefix = fmt.Sprintf("%s-%s", u.Username, provisionHostPrefix)
	// validate input

	chosenOption := opts[chosenAmiNumber]

	printVmInfo(chosenOption, provisionHostPrefix)
	confirm := askUser(fmt.Sprintf("Is this right [(%s)es / (%s)o / (%s)uit]: ", colorizeGreen("y"), colorizeYellow("n"), colorizeRed("q")))

	if !(confirm == "" || confirm == "yes" || confirm == "y") {
		exit()
	}

	prepareAnsibleConfig(chosenOption, provisionHostPrefix)

	executeAnsible()
}
