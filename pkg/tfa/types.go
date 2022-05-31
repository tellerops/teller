package tfa

import "runtime"

const (
	ModeTouchID = "touchID"
	ModeSudo    = "sudo"
)

type BuiltinTfaTypes struct {
}

// TypeHumanToMachine return all the supported 2fa per OS
func (t *BuiltinTfaTypes) TypeHumanToMachine() map[string]string {

	results := map[string]string{}

	switch runtime.GOOS {
	case "darwin":
		results[ModeTouchID] = ModeTouchID
		results[ModeSudo] = ModeSudo
	case "linux":
		results[ModeSudo] = ModeSudo
	}
	return results
}
