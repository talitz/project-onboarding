package commands

import (
	"errors"
	"os"
	"strconv"

	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// GetCreateCommand will create the repo and permissions
func GetCreateCommand() components.Command {
	return components.Command{
		Name:        "create",
		Description: "create repo structure and permissions",
		Aliases:     []string{"run"},
		Arguments:   getCreateArguments(),
		Flags:       getCreateFlags(),
		//		EnvVars:     geCreateFlags(),
		Action: func(c *components.Context) error {
			return createCmd(c)
		},
	}
}

func getCreateArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "pathToYamlFile ",
			Description: "location of the config YAML file",
		},
	}
}

func getCreateFlags() []components.Flag {
	return []components.Flag{
		components.BoolFlag{
			Name:         "dry-run",
			Description:  "show what will be done",
			DefaultValue: true,
		},
	}
}

// func geCreateFlags() []components.EnvVar {
// 	return []components.EnvVar{
// 		{
// 			Name:        "HELLO_FROG_GREET_PREFIX",
// 			Default:     "A new greet from your plugin template: ",
// 			Description: "Adds a prefix to every greet.",
// 		},
// 	}
// }

// type helloConfiguration struct {
// 	addressee string
// 	shout     bool
// 	repeat    int
// 	prefix    string
// }

func createCmd(c *components.Context) error {

	if len(c.Arguments) != 1 {
		return errors.New("Wrong number of arguments. Expected: 1, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}

	log.Output("using path : " + c.Arguments[0])
	_, err := os.Open(c.Arguments[0])

	if err != nil {
		return err
	}

	doCreate(c.GetStringFlagValue("config"), c.GetBoolFlagValue("dry-run"))
	return nil
}

func doCreate(configFile string, dryRun bool) {

	if dryRun {
		log.Output("DRY RUN")
	} else {
		log.Output("REAL RUN")
	}

	log.Output("PARSE FILE !")

}
