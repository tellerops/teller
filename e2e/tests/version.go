package test

import (
	"errors"
	"regexp"

	"github.com/spectralops/teller/e2e/register"
)

func init() { //nolint
	register.AddSuite("version", NewSuiteVersionCommand)
}

type SuiteVersionCommand struct {
	tempFolderPath string
}

func NewSuiteVersionCommand(tempFolderPath string) register.TestCaseDescriber {
	return &SuiteVersionCommand{
		tempFolderPath: tempFolderPath,
	}
}

func (v *SuiteVersionCommand) SetupTest() error {
	return nil
}

func (v *SuiteVersionCommand) GetFlags() []string {
	return []string{"version"}
}

func (v *SuiteVersionCommand) Check(stdOut, stderr string) error {
	var re = regexp.MustCompile(`(?m)Teller ([0-9]+)(\.[0-9]+)?(\.[0-9]+)
Revision [a-z0-9]{40}, date: [0-9]{4}-[0-9]{2}-[0-9]{2}`)

	if re.MatchString(stdOut) {
		return nil
	}

	return errors.New("invalid teller version")
}
