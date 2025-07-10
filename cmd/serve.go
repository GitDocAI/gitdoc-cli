package cmd

import (
	"fmt"

	"github.com/YHVCorp/DockitAICLI/config"
	"github.com/YHVCorp/DockitAICLI/utils"
)

func Serve(serveInput string) error {
	err := Deploy(serveInput, "")
	if err != nil {
		return fmt.Errorf("failed to deploy: %v", err)
	}

	cnf := config.GetConfig()

	errStr, isError := utils.ExecuteCmd("npm", cnf.OutputPath, "install", "--force")
	if isError {
		return fmt.Errorf("failed to execute npm install: %v", errStr)
	}

	err = utils.RunBlockingWithTerminalIO("npm", cnf.OutputPath, "run", "dev")
	if err != nil {
		return fmt.Errorf("failed to run npm dev: %v", err)
	}

	return nil
}
