package touchid

import (
	"fmt"

	tid "github.com/lox/go-touchid"
)

func Auth(command string) error {
	ok, err := tid.Authenticate(fmt.Sprintf("Execute command: %s", command))
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("2fa: invalid touch ID")
	}
	return nil
}
