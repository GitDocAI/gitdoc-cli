package webconfig

import config "github.com/YHVCorp/DockitAi/cli/internals/domain/models"

type WebConfigServicePort interface {
	StringifyConfig(config.WebConfig) (string, error)
	ParseConfig(string) (config.WebConfig, error)
}
