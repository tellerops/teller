package sudo

// Currently, we are not supporting Sudo two-factor authentication in windows

func Auth(command string) error {
	return nil
}
