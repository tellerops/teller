package tfa

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spectralops/teller/pkg"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/tfa/auth/sudo"
	"github.com/spectralops/teller/pkg/tfa/auth/touchid"
)

const (
	// folderPermission is the permission mode that the folder created.
	folderPermission = 055 // r--rw-r--
	// filePermission is the permission that 2fa file is created.
	filePermission = 060 // r--rw----
)

type TwoFactorAuthenticationDescriber interface {
	Enable() error
	Disable() error
	Prompt(command string) error
}

type TwoFactorAuthentication struct {
	Porcelain     *pkg.Porcelain
	logger        logging.Logger
	types         map[string]string
	locationsPath []string
}

func NewTwoFactorAuthentication(logger logging.Logger, locationsPath []string) TwoFactorAuthenticationDescriber {
	types := BuiltinTfaTypes{}

	return &TwoFactorAuthentication{
		Porcelain:     &pkg.Porcelain{Out: os.Stdout},
		logger:        logger,
		types:         types.TypeHumanToMachine(),
		locationsPath: locationsPath,
	}
}

func (tfa *TwoFactorAuthentication) Enable() error {

	if len(tfa.types) == 0 {
		return fmt.Errorf("two factor authentication not supported in os: %s", runtime.GOOS)
	}

	isFound, TfaType := tfa.getTwoFAConfigureFile()
	if isFound {
		return fmt.Errorf("`%s` 2fa is already configure, to change the 2fa type you need delete it before configure a new one", TfaType)

	}

	answers, err := tfa.Porcelain.StartTwoFAWizard(tfa.types, tfa.locationsPath)
	if err != nil {
		return err
	}

	val, ok := tfa.types[answers.Type]
	if !ok {
		tfa.logger.WithError(err).WithFields(map[string]interface{}{
			"type":    answers.Type,
			"options": tfa.types,
		}).Debug("unexpected answer type")
		return errors.New("invalid 2fa type")
	}

	return tfa.crateTwoFAFile(answers.Path, val)
}

func (tfa *TwoFactorAuthentication) Disable() error {
	fmt.Println("disable")
	return nil
}

func (tfa *TwoFactorAuthentication) Prompt(command string) error {
	isFound, TfaType := tfa.getTwoFAConfigureFile()

	if !isFound {
		return nil
	}

	switch TfaType {
	case ModeTouchID:
		return touchid.Auth(command)
	case ModeSudo:
		return sudo.Auth(command)
	default:
		tfa.logger.WithField("file_name", TfaType).Debug("unrecognized type")
	}
	return nil
}

// crateTwoFAFile creates a two factor authentication file from the given parameters
func (tfa *TwoFactorAuthentication) crateTwoFAFile(path, filename string) error {

	err := os.MkdirAll(path, folderPermission)
	if err != nil {
		tfa.logger.WithField("path", path).WithError(err).Fatal("could't crete folder")
		return err
	}

	filePath := filepath.Join(path, filename)
	err = os.WriteFile(filePath, []byte(""), filePermission)
	if err != nil {
		tfa.logger.WithField("path", filePath).WithError(err).Fatal("could't cerate file")
	}
	return nil
}

// getTwoFAConfigureFile search if two-factor authentication is enabled.
// This function will look in several locations when searching if a file exists in one of the paths.
func (tfa *TwoFactorAuthentication) getTwoFAConfigureFile() (bool, string) { //nolint

	for _, path := range tfa.locationsPath {
		logger := tfa.logger.WithField("path", path)
		logger.Debug("search 2fa file")
		files, err := os.ReadDir(path)
		if err != nil {
			logger.Debug("folder not found")
			continue
		}

		if len(files) == 0 {
			logger.Debug("2fa file not found in  path")
			continue
		}
		// todo:: thinks on the behavior
		if len(files) > 1 {
			logger.Debug("found more the one 2fa file, taking the first")
		}
		return true, files[0].Name()
	}

	return false, ""

}
