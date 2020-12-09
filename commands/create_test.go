package commands

import (
	"testing"
)

func TestMavenProject(t *testing.T) {
	stages := make([]string, 4, 4)
	stages[0] = "dev"
	stages[1] = "qa"
	stages[2] = "uat"
	stages[3] = "prod"

	//	createRepositories("ninja", "maven", stages, true)
}
