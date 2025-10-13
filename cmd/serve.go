package cmd

import (
	"fmt"

	"github.com/YHVCorp/gitdocai-cli/config"
	"github.com/YHVCorp/gitdocai-cli/utils"
)

func Serve(serveInput string) error {
	err := Deploy(serveInput, "")
	if err != nil {
		return fmt.Errorf("failed to deploy: %v", err)
	}


	webhook := NewWebhookServer(serveInput,"8082")

	webhook_url:=fmt.Sprintf(config.WebhookRoute,"8082")
	//this is for now is just a test mode for local dev
	envVars := map[string]string{
		"WEBHOOK_URL":webhook_url,
		"AUTHENTICATION":"Bearer jojojoijojoijio",
	}
	cnf := config.GetConfig()

	errStr, isError := utils.ExecuteCmd("npm", cnf.OutputPath, "install", "--force")
	if isError {
		return fmt.Errorf("failed to execute npm install: %v", errStr)
	}


	go func(){
		InitFolderListen(*cnf)
	}()

	go func(){
		webhook.Start()
		fmt.Printf("Webhook listening on: %s",webhook_url)
	}()

	err = utils.RunBlockingWithTerminalIO("npm", cnf.OutputPath,envVars, "run", "dev")
	if err != nil {
		 fmt.Errorf("failed to run npm dev: %v", err)
	}
	return nil
}
