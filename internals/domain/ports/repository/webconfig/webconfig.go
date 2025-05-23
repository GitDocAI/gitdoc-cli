package webconfig

import config "github.com/YHVCorp/DockitAi/cli/internals/domain/models"

type WebConfigRepositoryPort interface {
	SaveConfig(webconfig config.WebConfig, jsonconfig string) error
}
