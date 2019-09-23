package ci

import (
	"fmt"
	"github.com/cvbarros/go-teamcity/teamcity"
	"github.com/liamg/tml"
	"net/http"
	"os"
	"testing"
)

func TestPeter(t *testing.T) {
	parentProjectName := "GOT1"
	name := "testing2"
	gitUrl := "git@bitbucket.org:easypark-ondemand/peter_test.git"
	uploadedKeyName := "tcity@ezp-stg-bld-001"

	client, err := teamcity.NewClientWithAddress(teamcity.TokenAuth("eyJ0eXAiOiAiVENWMiJ9.UTE5S0NjdVVLenFSdW1FT3FxZXhCeGNGNFBZ.MjllZjU3MDYtYjQ0Yy00NjUxLTg4ZTItZmNkMmU3OGVmMDA4"), "https://teamcity.easyparkinfra.net", http.DefaultClient)
	if err != nil {
		tml.Println(fmt.Sprintf("<red>Failed to create client</red>\n%v", err))
		os.Exit(-1)
	}

	parentProject, err := client.Projects.GetByName(parentProjectName)
	if err != nil {
		tml.Println(fmt.Sprintf("<red>Failed to find project</red> %s\n%v", parentProjectName, err))
		os.Exit(-1)
	}

	tml.Println(fmt.Sprintf("Found parent project named <green>%s</green>", parentProjectName))

	newProject, err := client.Projects.Create(&teamcity.Project{
		Name:            name,
		ParentProject:   parentProject.ProjectReference(),
		Parameters:      teamcity.NewParametersEmpty(),
		ProjectFeatures: teamcity.NewProjectFeaturesEmpty(),
	})
	if err != nil {
		tml.Println(fmt.Sprintf("<red>Failed to create project%s</red>\n%v", name, err))
		os.Exit(-1)
	}

	options, err := teamcity.NewGitVcsRootOptions(
		"master",
		gitUrl,
		"",
		teamcity.GitAuthSSHUploadedKey,
		"git",
		"",
	)

	if err != nil {
		tml.Println(fmt.Sprintf("<red>Failed to create VCS Options</red>\n%v", err))
		os.Exit(-1)
	}

	options.PrivateKeySource = uploadedKeyName
	vcsRootName := fmt.Sprintf("%s-GitVcsRoot", name)
	vcsRoot, err := teamcity.NewGitVcsRoot(newProject.ID, vcsRootName, options)

	if err != nil {
		tml.Println(fmt.Sprintf("<red>Failed to create VCS Root</red>\n%v", err))
		os.Exit(-1)
	}
	fmt.Println(vcsRootName)
	vcs, err := client.VcsRoots.Create(newProject.ID, vcsRoot)

	if err != nil {
		tml.Println(fmt.Sprintf("<red>Failed to add VCS Root</red>\n%v", err))
		os.Exit(-1)
	}

	newProject.ProjectFeatures = teamcity.NewProjectFeatures(
		teamcity.VersionedSettings(
			teamcity.Kotlin,
			teamcity.PreferVcs,
			vcs.ID),
	)

	updatedProject, err := client.Projects.Update(newProject)
	if err != nil {
		tml.Println(fmt.Sprintf("<red>Failed to add Versioned settings</red>\n%v", err))
		os.Exit(-1)
	}


	fmt.Println(updatedProject)
}