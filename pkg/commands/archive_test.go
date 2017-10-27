package commands

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/Azure/acr-builder/pkg/constants"
	"github.com/Azure/acr-builder/pkg/domain"
	test_domain "github.com/Azure/acr-builder/tests/mocks/pkg/domain"
	"github.com/Azure/acr-builder/tests/testCommon"
	"github.com/stretchr/testify/assert"
)

type obtainTestCase struct {
	url           string
	targetDir     string
	getWdErr      *error
	expectedChdir test_domain.ChdirExpectations
	//expectedFSAccess  test_domain.FileSystemExpectations
	expectedCommands  []test_domain.CommandsExpectation
	expectedObtainErr string
	expectedExports   []domain.EnvVar
	expectedReturnErr string
}

func TestObtainFromKnownLocation(t *testing.T) {
	targetDir := filepath.Join(testCommon.Config.ProjectRoot, "tests", "workspace")
	testArchiveSource(t,
		obtainTestCase{
			url:       testCommon.Config.Archive.URL,
			targetDir: targetDir,
			expectedChdir: []test_domain.ChdirExpectation{
				{Path: targetDir},
				{Path: "home"},
			},
			getWdErr: &testCommon.NilError,
			expectedExports: []domain.EnvVar{
				{Name: constants.ExportsWorkingDir, Value: targetDir},
			},
		},
	)
}

func testArchiveSource(t *testing.T, tc obtainTestCase) {
	cleanup(tc.targetDir)
	defer cleanup(tc.targetDir)
	source := NewArchiveSource(tc.url, tc.targetDir)
	runner := test_domain.NewMockRunner()
	runner.PrepareCommandExpectation(tc.expectedCommands)
	fs := runner.GetFileSystem().(*test_domain.MockFileSystem)
	fs.PrepareChdir(tc.expectedChdir)
	if tc.getWdErr != nil {
		fs.On("Getwd").Return("home", *tc.getWdErr).Once()
	}
	err := source.Obtain(runner)
	if tc.expectedObtainErr != "" {
		assert.NotNil(t, err)
		assert.Regexp(t, regexp.MustCompile(tc.expectedObtainErr), err.Error())
		return
	}
	assert.Nil(t, err)

	var projectDockerComposeFile string
	if tc.targetDir == "" {
		projectDockerComposeFile = "docker-compose.yml"
	} else {
		projectDockerComposeFile = filepath.Join(tc.targetDir, "docker-compose.yml")
	}
	_, err = os.Stat(projectDockerComposeFile)
	assert.Nil(t, err)

	exports := source.Export()
	assert.Equal(t, tc.expectedExports, exports)
	err = source.Return(runner)
	if tc.expectedReturnErr != "" {
		assert.NotNil(t, err)
		assert.Regexp(t, regexp.MustCompile(tc.expectedReturnErr), err.Error())
		return
	}
	assert.Nil(t, err)
}

func cleanup(targetDir string) {
	if targetDir != "" {
		err := os.RemoveAll(targetDir)
		if err != nil {
			panic("Cleanup error: " + err.Error())
		}
	}
}
