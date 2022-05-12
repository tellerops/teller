package testutils

import (
	"bytes"
	"os/exec"
)

type DifferExec struct {
}

func ExecCmd(name string, arg []string, workingDirectory string) (stdout, stderr string, err error) {

	var r []string
	for _, str := range arg {
		if str != "" {
			r = append(r, str)
		}
	}
	cmd := exec.Command(name, r...)
	var stdoutBuff bytes.Buffer
	var stderrBuff bytes.Buffer
	cmd.Dir = workingDirectory
	cmd.Stdout = &stdoutBuff
	cmd.Stderr = &stderrBuff

	err = cmd.Run()

	return stdoutBuff.String(), stderrBuff.String(), err
}

func NewExecDiffer() Differ {
	return &DifferExec{}
}

func (de *DifferExec) Diff(dir1, dir2 string, ignores []string) (string, error) {

	flags := []string{"-qr"}
	for _, ignore := range ignores {
		flags = append(flags, "-x", ignore)
	}

	flags = append(flags, dir1, dir2)
	stdout, _, err := ExecCmd("diff", flags, "")

	return stdout, err

}
