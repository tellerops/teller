package testutils

import (
	"bytes"
	"os/exec"
)

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
