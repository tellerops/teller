package e2e

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"testing"

	"github.com/spectralops/teller/e2e/register"
	_ "github.com/spectralops/teller/e2e/tests"
	"github.com/spectralops/teller/e2e/testutils"
	"github.com/stretchr/testify/assert"
)

const (
	replaceToStaticPath      = "DYNAMIC-FULL-PATH"
	replaceToShortStaticPath = "DYNAMIC-SHORT-PATH"
	removeBinaryPlaceholder  = "<binary-path>"
	testsFolder              = "tests"
)

func TestE2E(t *testing.T) { //nolint
	t.Parallel()

	// validate given binary path
	binaryPath, err := getBinaryPath()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	snapshotSuites, err := testutils.GetYmlSnapshotSuites(testsFolder)
	assert.Nil(t, err)

	// consider to replace `diff` command which is depended on OS to golang plugin.
	// could't find something better
	differ := testutils.NewExecDiffer()
	// Loop on all test/*.yml files
	for _, snapshot := range snapshotSuites {

		t.Run(snapshot.Name, func(t *testing.T) {

			// create a temp folder for the test
			tempFolder, err := os.MkdirTemp(t.TempDir(), strings.ReplaceAll(snapshot.Name, " ", ""))
			// descrive the base snapshot data of the test
			snapshotFolder := filepath.Join(tempFolder, "snapshot")
			assert.Nil(t, err, "could not create temp folder")
			defer os.RemoveAll(tempFolder)

			if len(snapshot.InitSnapshot) > 0 {
				err = snapshot.CreateSnapshotData(snapshot.InitSnapshot, snapshotFolder)
				assert.Nil(t, err)
			}

			if snapshot.ConfigFileName != "" {
				err = snapshot.CrateConfig(snapshotFolder)
				assert.Nil(t, err)
			}

			flagsCommand := strings.TrimPrefix(snapshot.Command, removeBinaryPlaceholder)
			stdout, stderr, err := testutils.ExecCmd(binaryPath, strings.Split(flagsCommand, " "), snapshotFolder)
			if stdout == "" {
				assert.Nil(t, err, stderr)
			}

			// In case the stdout/stderr include the dynamic folder path, we want to replace with static-content for better snapshot text compare
			stdout, stderr = replaceFolderName(stdout, stderr, snapshotFolder)

			if len(snapshot.ReplaceStdOutContent) > 0 {
				for _, r := range snapshot.ReplaceStdOutContent {
					var re = regexp.MustCompile(r.Search)
					stdout = re.ReplaceAllString(stdout, r.Replace)
				}
			}
			if len(snapshot.ReplaceStdErrContent) > 0 {
				for _, r := range snapshot.ReplaceStdErrContent {
					var re = regexp.MustCompile(r.Search)
					stderr = re.ReplaceAllString(stderr, r.Replace)
				}
			}

			if snapshot.ExpectedStdOut != "" {
				assert.Equal(t, snapshot.ExpectedStdOut, stdout)
			}

			if snapshot.ExpectedStdErr != "" {
				assert.Equal(t, snapshot.ExpectedStdErr, stderr)
			}

			if len(snapshot.ExpectedSnapshot) > 0 {
				destSnapshotFolder := filepath.Join(tempFolder, "dest")
				err = snapshot.CreateSnapshotData(snapshot.ExpectedSnapshot, destSnapshotFolder)
				assert.Nil(t, err)

				diffResult, err := testutils.FolderDiff(differ, destSnapshotFolder, snapshotFolder, []string{snapshot.ConfigFileName})
				if err != nil {
					t.Fatalf("snapshot folder is not equal. results: %v", diffResult)
				}
				assert.Nil(t, err)
			}
		})
	}
	// loop on register suites (from *.go files)
	for name, suite := range register.GetSuites() {
		t.Run(name, func(t *testing.T) {

			// creates temp dir for test path.
			tempFolder, err := os.MkdirTemp(t.TempDir(), strings.ReplaceAll(name, " ", ""))
			assert.Nil(t, err, "could not create temp folder")
			defer os.RemoveAll(tempFolder)

			// initialized test case
			testInstance := suite(tempFolder)

			err = testInstance.SetupTest()
			assert.Nil(t, err)

			// get Teller flags command
			flags := testInstance.GetFlags()

			stdout, stderr, err := testutils.ExecCmd(binaryPath, flags, tempFolder)
			assert.Nil(t, err)

			stdout, stderr = replaceFolderName(stdout, stderr, tempFolder)
			err = testInstance.Check(stdout, stderr)
			assert.Nil(t, err)
		})
	}
}

func replaceFolderName(stdout, stderr, workingDirectory string) (string, string) {
	stdout = strings.ReplaceAll(stdout, workingDirectory, replaceToStaticPath)
	stderr = strings.ReplaceAll(stderr, workingDirectory, replaceToStaticPath)
	shortFolderPath := workingDirectory[0:13]
	stdout = strings.ReplaceAll(stdout, shortFolderPath, replaceToShortStaticPath)
	stderr = strings.ReplaceAll(stderr, shortFolderPath, replaceToShortStaticPath)

	return stdout, stderr
}

func getBinaryPath() (string, error) {
	binaryPath, isExists := os.LookupEnv("BINARY_PATH")
	if !isExists {
		return "", errors.New("missing `BINARY_PATH`")
	}

	info, err := os.Stat(binaryPath)
	errors.Is(err, os.ErrNotExist)

	if err != nil || info.IsDir() {
		return "", fmt.Errorf("%s not found", binaryPath)
	}
	return binaryPath, nil
}
