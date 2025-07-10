package cmd

import (
	"fmt"
	"os/exec"
)

func ValidateSystemRequirements() error {
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is required before running this command, please install it")
	}

	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is required before running this command, please install it")
	}

	return nil
}
