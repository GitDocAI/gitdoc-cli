package webconfig

import (
	"errors"
	"fmt"
	"os"

	wc_repository_port "github.com/YHVCorp/DockitAi/cli/internals/domain/ports/repository/webconfig"
	wc_service_port "github.com/YHVCorp/DockitAi/cli/internals/domain/ports/service/webconfig"
)

type WebConfigController struct {
	webconfig_service wc_service_port.WebConfigServicePort
	webconfig_repo    wc_repository_port.WebConfigRepositoryPort
}

func NewWebConfigController(
	wc_repo wc_repository_port.WebConfigRepositoryPort,
	wc_service wc_service_port.WebConfigServicePort,
) *WebConfigController {
	return &WebConfigController{
		webconfig_repo:    wc_repo,
		webconfig_service: wc_service,
	}
}

func (wcc *WebConfigController) InitializeWebConfig(client_config_path string) error {
	raw_client_config, err := os.ReadFile(client_config_path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading client config: %s", err.Error()))
	}
	client_config, err := wcc.webconfig_service.ParseConfig(string(raw_client_config))
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing client config: %s", err.Error()))
	}
	error := wcc.webconfig_repo.SaveConfig(client_config, string(raw_client_config))
	if error != nil {
		return errors.New(fmt.Sprintf("Error parsing client config: %s", error.Error()))
	}

	return nil
}
