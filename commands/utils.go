package commands

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands"

	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/log"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
)

type commonConfiguration struct {
	details *config.ArtifactoryDetails
	path    string
}

func createCommonConfiguration(c *components.Context) (*commonConfiguration, error) {
	if err := checkInputs(c); err != nil {
		return nil, err
	}

	confDetails, err := getRtDetails(c)
	if err != nil {
		return nil, err
	}

	// Increase log level to avoid search command logs
	increaseLogLevel()

	conf := &commonConfiguration{
		details: confDetails,
		path:    c.Arguments[0],
	}
	return conf, nil
}

func checkInputs(c *components.Context) error {
	if len(c.Arguments) != 1 {
		return errors.New("Wrong number of arguments. Expected: 1, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}
	if strings.Contains(c.Arguments[0], "*") {
		return errors.New("Wildcards are not supported in paths")
	}
	return nil
}

// Returns the Artifactory Details of the provided server-id, or the default one.
func getRtDetails(c *components.Context) (*config.ArtifactoryDetails, error) {
	details, err := commands.GetConfig(c.GetStringFlagValue("server-id"), false)
	if err != nil {
		return nil, err
	}
	if details.Url == "" {
		return nil, errors.New("no server-id was found, or the server-id has no url")
	}
	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	err = config.CreateInitialRefreshableTokensIfNeeded(details)
	if err != nil {
		return nil, err
	}
	return details, nil
}

// Set the log level to ERROR to avoid the following outputs:
// [Info] Searching artifacts...
// [Info] Found 1 artifact.
func increaseLogLevel() {
	if log.GetCliLogLevel() == clientlog.INFO {
		clientlog.SetLogger(clientlog.NewLogger(clientlog.ERROR, nil))
	}
}
