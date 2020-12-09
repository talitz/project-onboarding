package commands

import (
	"bufio"
	"encoding/json"
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

type repoInfo struct {
	Name, Layout, RemoteURL string
}

var repoRefInfo = map[string]repoInfo{
	"maven":  repoInfo{"jcenter", "maven-2-default", "https://jcenter.bintray.io"},
	"nuget":  repoInfo{"nugetorg", "nuget-default", "https://nuget.org"},
	"docker": repoInfo{"dockerub", "simple-default", "https://dockerhub.io"},
	"npm":    repoInfo{"npmjs", "simple-default", "https://www.npmjs.org"},
}

var tempFolder = "./tmpConfigFiles/"
var lstDefaultProfiles = []string{"dev", "delivery", "ops", "sec"}

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
			DefaultValue: false,
		},
	}
}

// Profile will be used to create a permission Target
// type ProjectProfiles struct {
// 	type Profile struct {

// 	} `yaml:"name"`
// 	Name      string   `yaml:"name"`
// 	RepoOwner []string `yaml:"repo-owner"`
// }

// Stage describes a runtime env
type Stage struct {
	Name  string `yaml:"name"`
	Owner string `yaml:"owner"`
}

// Project describes a project to create
type Project struct {
	Name     string `yaml:"name"`
	RepoType string `yaml:"repoType"`
	//	Stages   []string `yaml:"stages"`
	Stages   []Stage `yaml:"stages"`
	Profiles []string
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

// RtGroup is used for the Create Group Rest API
type RtGroup struct {
	Desc string `json:"description"`
	//	Realm     string `json:"realm"`
	WatchMgr  bool `json:"watchManager"`
	PolicyMgr bool `json:"policyManager"`
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

	doCreate(c.Arguments[0], c.GetBoolFlagValue("dry-run"), conf, c)
	return nil
}

func doCreate(configFile string, dryRun bool, c *commonConfiguration, ctxt *components.Context) error {

	if dryRun {
		log.Output("DRY RUN Mode : No change will be done to your Artifactory !!!")
	}

	//var profiles map[string][]Profile
	profiles := make(map[string][]string)

	ParseOnboardingTemplate(c, configFile, profiles)

	for pn, profiles := range profiles {
		generateJSONFiles(profiles, tempFolder+pn)
		if !dryRun {
			createGroups(pn, profiles, c)
			// createPermissionTargets(c)
		}
	}

	return nil
}

// ParseOnboardingTemplate reads the template
func ParseOnboardingTemplate(c *commonConfiguration, configFile string, profiles map[string][]string) error {

	var projectsToInit Projects
	var lstRepo []Repository

	// read our opened yaml file as a byte array.
	byteValue, _ := ioutil.ReadFile(configFile)

	err := yaml.Unmarshal(byteValue, &projectsToInit)

	if err != nil {
		return errors.New("Errors occured when reading Yaml File ")
	}

	log.Output("Parsing " + configFile + "...")

	for _, project := range projectsToInit.ArrProj {

		initTmpFolder(project.Name)

		lstRepo = []Repository{}
		var aggregatedRepo []string
		var name string

		// get profiles
		profiles[project.Name] = []string{}

		for _, prof := range project.Profiles {
			//			log.Output("loop: " + prof)
			profiles[project.Name] = append(profiles[project.Name], prof)
		}

		// get repo name for local
		for _, stage := range project.Stages {
			name = project.Name + "-" + project.RepoType + "-" + stage.Name + "-local"
			//			log.Output(name)
			lstRepo = append(lstRepo,
				Repository{
					name,
					"local",
					project.RepoType,
					repoRefInfo[project.RepoType].Layout, "", nil, ""})
			aggregatedRepo = append(aggregatedRepo, name)
		}

		// get repo name for remote based on package type
		name = project.Name + "-" + repoRefInfo[project.RepoType].Name + "-remote"
		//log.Output(name)

		lstRepo = append(lstRepo, Repository{
			name,
			"remote",
			project.RepoType,
			repoRefInfo[project.RepoType].Layout,
			repoRefInfo[project.RepoType].RemoteURL, nil, ""})
		aggregatedRepo = append(aggregatedRepo, name)

		// get repo name for virtual
		name = project.Name + "-" + project.RepoType
		//		log.Output(name)

		lstRepo = append(lstRepo, Repository{
			name,
			"virtual",
			project.RepoType,
			repoRefInfo[project.RepoType].Layout, "", aggregatedRepo, ""})

		log.Output("Parsing done !!")

		yamlFile := tempFolder + project.Name + "/yaml/repo.yml"
		generateYamlFile(yamlFile, lstRepo, profiles[project.Name])
		PatchConfigurationFile(c, yamlFile)

	}

	return nil
}

func generateYamlFile(yamlFile string, r []Repository, lstProfile []string) error {

	const indent1 = "  "
	const indent2 = "    "
	const indent3 = "      "

	log.Output("Generating " + yamlFile + "...")

	// For more granular writes, open a file for writing.
	f, err := os.Create(yamlFile)

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
	log.Output("Generated " + yamlFile + " !!!")

	return nil
}

// PatchConfigurationFile executing the configuration.yml file changes to the artifactory instance
func PatchConfigurationFile(c *commonConfiguration, yamlFile string) error {
	arguments := []string{"-XPATCH", "/api/system/configuration", "-H", "\"Content-Type: application/yaml\"", "-T", yamlFile}
	curlCmd := curl.NewCurlCommand().SetArguments(arguments).SetRtDetails(c.details)

	log.Output("Patching Artifactory configuration ... ")

	if err := commands.Exec(curlCmd); err != nil {
		return err
	}

	log.Output("Patched Artifactory !!!")

	return nil
}

func generateJSONFiles(lstProfiles []string, tempFolder string) error {

	for _, profile := range lstProfiles {

		log.Output("Generating " + profile + "...")

		// For more granular writes, open a file for writing.
		f, err := os.Create(tempFolder + "/json/" + profile + ".json")

		if err != nil {

			return errors.New("Errors occured when writing JSON file ")
		}

		defer f.Close()

		w := bufio.NewWriter(f)

		//		var arrByte []byte
		arrByte, _ := json.Marshal(RtGroup{"Auto generated", false, false})

		w.WriteString(string(arrByte))
		w.Flush()
		f.Sync()
		log.Output("Generated json/" + profile + ".json")

	}
	return nil
}

func createGroups(projectName string, lstProfiles []string, c *commonConfiguration) error {
	endPoint := "/api/security/groups/"

	for _, profile := range lstProfiles {
		jsonData := tempFolder + projectName + "/json/" + profile + ".json"

		arguments := []string{"-XPUT", endPoint + projectName + "-" + profile, "-H", "\"Content-Type: application/json\"", "-T", jsonData}
		curlCmd := curl.NewCurlCommand().SetArguments(arguments).SetRtDetails(c.details)

		if err := commands.Exec(curlCmd); err != nil {
			return err
		}
	}

	return nil
}

// POTENTIAL BUG on the CLI Core :
//  - cannot create permission on NON E+ instances !
//  - cannot create permission target as repo is never detected !
func createPermissionTargets(c *commonConfiguration) error {

	// // profiles dev, release manager, ops, sec
	// log.Output(c.details.ServerId)

	// // for _, profile := range lstProfiles {
	// // }

	// log.Output("createPermissionTargets")
	// params := services.NewPermissionTargetParams()

	// params.Name = "java-developers"
	// params.Repo.Repositories = []string{"android-sdk"}
	// params.Repo.Actions.Groups = map[string][]string{
	// 	"dev": {"manage", "read", "annotate"},
	// }

	// servicesManager, err := utils.CreateServiceManager(c.details, false)
	// if err != nil {
	// 	return errors.New("Errors occured when declaring serviceManager ")
	// }
	// log.Output(params)

	// log.Output(servicesManager.GetRepository("android-sdk")

	// err = servicesManager.CreatePermissionTarget(params)

	// if err != nil {
	// 	log.Output(err)
	// 	return errors.New("Errors occured when creating permission ")
	// }

	// log.Output("after create permission")

	return nil
}

func initTmpFolder(projectName string) {

	//Create a folder/directory at a full qualified path
	// if _, err := os.Stat(tempFolder); os.IsNotExist(err) {
	// 	log.Output(tempFolder + " created")
	// 	_ = os.Mkdir(tempFolder, 0755)
	// }

	if _, err := os.Stat(tempFolder + projectName + "/yaml"); os.IsNotExist(err) {
		_ = os.MkdirAll(tempFolder+projectName+"/yaml", 0755)
		log.Output(tempFolder + projectName + "/yaml folder created")
	}

	if _, err := os.Stat(tempFolder + projectName + "/json"); os.IsNotExist(err) {
		_ = os.MkdirAll(tempFolder+projectName+"/json", 0755)
		log.Output(tempFolder + projectName + "/json folder created")
	}

}
