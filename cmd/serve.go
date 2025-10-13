package cmd

import (
	"fmt"

	"github.com/YHVCorp/gitdocai-cli/config"
	"github.com/YHVCorp/gitdocai-cli/utils"
)

func Serve(serveInput,webhook_url,authentication string) error {
	err := Deploy(serveInput, "")
	if err != nil {
		return fmt.Errorf("failed to deploy: %v", err)
	}

	if webhook_url==""{
		webhook_url=fmt.Sprintf(config.WebhookRoute,"8082")
		webhook := NewWebhookServer(serveInput,"8082")
		go func(){
			webhook.Start()
			fmt.Printf("Webhook listening on: %s",webhook_url)
		}()
	}

	//this is for now is just a test mode for local dev
	envVars := map[string]string{
		"WEBHOOK_URL":webhook_url,
		"AUTHENTICATION":authentication,
	}
	cnf := config.GetConfig()

	errStr, isError := utils.ExecuteCmd("npm", cnf.OutputPath, "install", "--force")
	if isError {
		return fmt.Errorf("failed to execute npm install: %v", errStr)
	}


	go func(){
		InitFolderListen(*cnf)
	}()


	err = utils.RunBlockingWithTerminalIO("npm", cnf.OutputPath,envVars, "run", "dev")
	if err != nil {
		 fmt.Errorf("failed to run npm dev: %v", err)
	}
	return nil
}
