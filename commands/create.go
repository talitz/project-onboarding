package commands

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/curl"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"gopkg.in/yaml.v2"
)

var repoRefInfo = map[string]repoInfo{
	"maven":  repoInfo{"jcenter", "maven-2-default", "https://jcenter.bintray.io"},
	"nuget":  repoInfo{"nugetorg", "nuget-default", "https://nuget.org"},
	"docker": repoInfo{"dockerub", "simple-default", "https://dockerhub.io"},
	"npm":    repoInfo{"npmjs", "simple-default", "https://www.npmjs.org"},
}

type repoInfo struct {
	Name, Layout, RemoteURL string
}

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

// Repository define a virtual, remote, local repo
type Repository struct {
	// common to all repo types
	Name       string
	Type       string
	PkgType    string
	RepoLayout string

	// only for remote
	URL string

	// only for virtual
	RepoList   []string
	DeployRepo string
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

	if dryRun {
		log.Output("DRY RUN")
	} else {
		log.Output("REAL RUN")
	}
	ParseOnboardingTemplate(c, configFile)
	//	BuildConfigurationFile(c)

	return nil
}

// ParseOnboardingTemplate reads the template
func ParseOnboardingTemplate(c *commonConfiguration, configFile string) error {

	var projectsToInit Projects
	var lstRepo []Repository

	// read our opened yaml file as a byte array.
	byteValue, _ := ioutil.ReadFile(configFile)

	err := yaml.Unmarshal(byteValue, &projectsToInit)

	if err != nil {
		return errors.New("Errors occured when reading Yaml File ")
	}

	for _, project := range projectsToInit.ArrProj {

		lstRepo = []Repository{}
		//			len(project.Stages)+2)
		var aggregatedRepo []string
		var name string

		// get repo name for local
		for _, stage := range project.Stages {
			name = project.Name + "-" + project.RepoType + "-" + stage + "-local"
			log.Output(name)
			lstRepo = append(lstRepo,
				Repository{
					name,
					"local",
					project.RepoType,
					repoRefInfo[project.RepoType].Layout, "", nil, ""})
			aggregatedRepo = append(aggregatedRepo, name)
		}

		// get repo name for remote based on package type
		name = repoRefInfo[project.RepoType].Name + "-remote"
		log.Output(name)

		lstRepo = append(lstRepo, Repository{
			name,
			"remote",
			project.RepoType,
			repoRefInfo[project.RepoType].Layout,
			repoRefInfo[project.RepoType].RemoteURL, nil, ""})
		aggregatedRepo = append(aggregatedRepo, name)

		// get repo name for virtual
		name = project.Name + "-" + project.RepoType
		log.Output(name)

		lstRepo = append(lstRepo, Repository{
			name,
			"virtual",
			project.RepoType,
			repoRefInfo[project.RepoType].Layout, "", aggregatedRepo, ""})

		genrateYamlFile(project.Name, lstRepo)
		PatchConfigurationFile(c, project.Name)
	}

	return nil
}

func genrateYamlFile(projectName string, r []Repository) error {

	const indent1 = "  "
	const indent2 = "    "
	const indent3 = "      "

	// For more granular writes, open a file for writing.
	f, err := os.Create(projectName + ".yml")

	if err != nil {
		return errors.New("Errors occured when writing YAML file ")
	}

	defer f.Close()

	w := bufio.NewWriter(f)

	// inject local repo
	w.WriteString("localRepositories:\n")
	for _, repo := range r {
		if repo.Type == "local" {
			w.WriteString(indent1 + repo.Name + ":\n")
			w.WriteString(indent2 + "type: " + repo.PkgType + "\n")
			w.WriteString(indent2 + "repoLayout: " + repo.RepoLayout + "\n")
		}
	}

	// inject remote repo
	w.WriteString("remoteRepositories:\n")
	for _, repo := range r {
		if repo.Type == "remote" {
			w.WriteString(indent1 + repo.Name + ":\n")
			w.WriteString(indent2 + "type: " + repo.PkgType + "\n")
			w.WriteString(indent2 + "repoLayout: " + repo.RepoLayout + "\n")
			w.WriteString(indent2 + "url: " + repo.URL + "\n")
		}
	}

	// inject virtual repo
	w.WriteString("virtualRepositories:\n")
	for _, repo := range r {
		if repo.Type == "virtual" {
			w.WriteString(indent1 + repo.Name + ":\n")
			w.WriteString(indent2 + "type: " + repo.PkgType + "\n")
			w.WriteString(indent2 + "repoLayout: " + repo.RepoLayout + "\n")
			w.WriteString(indent2 + "repositories:\n")
			for _, aggRepo := range repo.RepoList {
				w.WriteString(indent3 + "- " + aggRepo + "\n")
			}
			w.WriteString(indent2 + "defaultDeploymentRepo: " + r[0].Name + "\n")
		}
	}
	w.Flush()
	f.Sync()
	return nil
}

// PatchConfigurationFile executing the configuration.yml file changes to the artifactory instance
func PatchConfigurationFile(c *commonConfiguration, projectName string) error {
	arguments := []string{"-XPATCH", "/api/system/configuration", "-H", "\"Content-Type: application/yaml\"", "-T", projectName + ".yml"}
	curlCmd := curl.NewCurlCommand().SetArguments(arguments).SetRtDetails(c.details)

	if err := commands.Exec(curlCmd); err != nil {
		return err
	}
	return nil
}
