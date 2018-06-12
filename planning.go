package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/vatine/planning/models"
)

func help(prog string) {
	fmt.Printf("%s <inputspec>... <file>\n\n\tinputspec should be <input>=<number>\n", prog)
	fmt.Println()
	fmt.Println("\tThe model file should be a YAML-formatted list of server models")
	fmt.Println("\tEach model should follow the following format:")
	fmt.Println("\tname: <name>\n\tinputs:\n\t - <input>\n\t   ...")
	fmt.Println("\toutputs:\n\t - backend: <backend name>\n\t   input: <name of backend's input>\n\t   expression: <value>")
	fmt.Println("\tvariables:\n\t <varname>: <expression>\n\t  ...")
	fmt.Println("\tresources:\n\t ram: <expression>\n\t cpu: <expression>")
	fmt.Println("\t replicas: <expression>")
}

func main () {
	inputs := make(map[string]models.Expression)
	usage := make(map[string]*models.Model)
	var filename string

	for _, arg := range(os.Args[1:]) {
		if arg == "help" {
			help(path.Base(os.Args[0]))
			return
		}
		if strings.Index(arg, "=") != -1 {
			// we have an input!
			tmp := strings.Split(arg, "=")
			name := tmp[0]
			value, err := models.Parse(tmp[1])
			if err != nil {
				inputs[name] = value
			}
		} else {
			filename = arg
		}
	}

	base := path.Base(filename)
	switch {
	case strings.HasSuffix(base, ".yaml"):
		base = base[:len(base)-5]
	case strings.HasSuffix(base, ".yml"):
		base = base[:len(base)-4]
	}

	f, fErr := os.Open(filename)
	if fErr != nil {
		fmt.Printf("Error opening %s, %s", filename, fErr)
	}
	defer f.Close()
	usageModel, err := models.LoadExternalModels(f)
	if err != nil {
		fmt.Printf("Error loading models, %s", err)
	}
	for _, m := range usageModel {
		usage[m.Name] = models.ModelFromExternal(m)
	}
	models.Propagate(usage, base, inputs)
}
