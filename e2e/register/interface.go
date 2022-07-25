package register

type NewSuite func(tempFolderPath string) TestCaseDescriber

type TestCaseDescriber interface {
	SetupTest() error
	GetFlags() []string
	Check(stdOut, stderr string) error
}
