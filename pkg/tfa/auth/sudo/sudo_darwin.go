package sudo

import (
	"fmt"
	"os"
)

func Auth(command string) error {
	if os.Geteuid() == 0 {
		return nil
	}
	return fmt.Errorf("2fa: needed sudo access to execute `%s`", command)
}
