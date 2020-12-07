package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/repository"

	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"gopkg.in/yaml.v2"
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
	var projectsToInit Projects

	if dryRun {
		log.Output("DRY RUN")
	} else {
		log.Output("REAL RUN")
	}

	// read our opened yaml file as a byte array.
	byteValue, _ := ioutil.ReadFile(configFile)

	err := yaml.Unmarshal(byteValue, &projectsToInit)

	if err != nil {
		return errors.New("Errors occured when reading Yaml File ")
	}

	for _, v := range projectsToInit.ArrProj {
		log.Output(v.Name)
		log.Output(v.RepoType)
		log.Output(v.Stages)
		CreateRepositories(v.Name, v.RepoType, v.Stages, c)
	}

	return nil
}

// CreateRepositories generates all the repositories
func CreateRepositories(projectName string, repoType string, stages []string, c *commonConfiguration) {
	CreateLocalRepositories(projectName, repoType, stages, c)
	CreateRemoteRepositories(projectName, repoType, stages, c)
	CreateVirtualRepositories(projectName, repoType, stages, c)
}

// CreateLocalRepositories generates all the local repositories
func CreateLocalRepositories(projectName string, repoType string, stages []string, c *commonConfiguration) error {
	log.Output("Create locals for ", projectName)
	LocalRepositoryTemplate := LocalRepositoryTemplate{
		key:         "local-repo-name",
		packageType: "maven",
		rclass:      "local",
	}

	file, _ := json.MarshalIndent(LocalRepositoryTemplate, "", " ")
	err := ioutil.WriteFile("templates/local-repo-template-autocreated.json", file, 0644)
	if err != nil {
		fmt.Printf("Unable to write file: %v", err)
	}
	repositoryCreateCmd := repository.NewRepoCreateCommand().SetTemplatePath("templates/local-repo-template-autocreated.json").SetRtDetails(c.details)

	if err := commands.Exec(repositoryCreateCmd); err != nil {
		return err
	}
	return nil
}

// CreateRemoteRepositories generates all the remote repository
func CreateRemoteRepositories(projectName string, repoType string, stages []string, c *commonConfiguration) error {
	log.Output("Create remote for ", projectName)

	repositoryCreateCmd := repository.NewRepoCreateCommand().SetTemplatePath("templates/remote-repo-template.json").SetRtDetails(c.details)
	if err := commands.Exec(repositoryCreateCmd); err != nil {
		return err
	}
	return nil
}

// CreateVirtualRepositories generates all the remote repository
func CreateVirtualRepositories(projectName string, repoType string, stages []string, c *commonConfiguration) error {
	log.Output("Create virtual for ", projectName)

	repositoryCreateCmd := repository.NewRepoCreateCommand().SetTemplatePath("templates/virtual-repo-template.json").SetRtDetails(c.details)
	if err := commands.Exec(repositoryCreateCmd); err != nil {
		return err
	}
	return nil
}
