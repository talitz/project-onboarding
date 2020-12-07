package commands

import (
	"errors"
	"os"
	"strconv"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/curl"
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

// Project describes a project to create
type Project struct {
	Name     string   `yaml:"name"`
	RepoType string   `yaml:"repoType"`
	Stages   []string `yaml:"stages"`
}

// Projects describes a list of projects
type Projects struct {
	ArrProj []Project `yaml:"projects"`
}

// LocalRepositoryTemplate Defines a local repository template
type LocalRepositoryTemplate struct {
	key, packageType, rclass string
}

func createCmd(c *components.Context) error {
	conf, err := createCommonConfiguration(c)

	if len(c.Arguments) != 1 {
		return errors.New("Wrong number of arguments. Expected: 1, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}

	log.Output("[INFO] Config file : " + c.Arguments[0])

	_, err = os.Open(c.Arguments[0])

	if err != nil {
		return errors.New("Cannot open " + c.Arguments[0])
	}

	doCreate(c.Arguments[0], c.GetBoolFlagValue("dry-run"), conf)
	return nil
}

func doCreate(configFile string, dryRun bool, c *commonConfiguration) error {
	// var projectsToInit Projects

	if dryRun {
		log.Output("DRY RUN")
	} else {
		log.Output("REAL RUN")
	}

	// read our opened yaml file as a byte array.
	// byteValue, _ := ioutil.ReadFile(configFile)

	// err := yaml.Unmarshal(byteValue, &projectsToInit)

	// if err != nil {
	// 	return errors.New("Errors occured when reading Yaml File ")
	// }

	// for _, v := range projectsToInit.ArrProj {
	// 	log.Output(v.Name)
	// 	log.Output(v.RepoType)
	// 	log.Output(v.Stages)
	// 	// CreateRepositories(v.Name, v.RepoType, v.Stages, c)
	// }
	BuildConfigurationFile(c)
	PatchConfigurationFile(c)
	return nil
}

// BuildConfigurationFile is creating the relevant configuration.yml file based on the onboarding template that ran with the plugin
func BuildConfigurationFile(c *commonConfiguration) error {
	return nil
}

// PatchConfigurationFile is executing the configuration.yml file changes to the artifactory instance
func PatchConfigurationFile(c *commonConfiguration) error {
	arguments := []string{"-XPATCH", "/api/system/configuration", "-H", "\"Content-Type: application/yaml\"", "-T", "configuration.yml"}
	curlCmd := curl.NewCurlCommand().SetArguments(arguments).SetRtDetails(c.details)

	if err := commands.Exec(curlCmd); err != nil {
		return err
	}
	return nil
}

// CreateRepositories generates all the repositories
// func CreateRepositories(projectName string, repoType string, stages []string, c *commonConfiguration) {

// 	// CreateLocalRepositories(projectName, repoType, stages, c)
// 	// CreateRemoteRepositories(projectName, repoType, stages, c)
// 	// CreateVirtualRepositories(projectName, repoType, stages, c)
