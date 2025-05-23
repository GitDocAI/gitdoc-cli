package webconfig

import (
	"encoding/json"
	"fmt"

	config "github.com/YHVCorp/DockitAi/cli/internals/domain/models"
)

type WebConfigAdapter struct{}

func NewWebConfigAdapter() *WebConfigAdapter {
	return &WebConfigAdapter{}
}

func (w *WebConfigAdapter) StringifyConfig(cfg config.WebConfig) (string, error) {
	type fullConfigMap map[string]any
	out := fullConfigMap{
		"defaultVersion": cfg.DefaultVersion,
		"versions":       getKeys(cfg.Versions),
	}

	for k, v := range cfg.Versions {
		out[k] = v
	}

	bytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (w *WebConfigAdapter) ParseConfig(data string) (config.WebConfig, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return config.WebConfig{}, fmt.Errorf("json parse error: %w", err)
	}

	var cfg config.WebConfig
	if err := json.Unmarshal(raw["defaultVersion"], &cfg.DefaultVersion); err != nil {
		return config.WebConfig{}, fmt.Errorf("invalid defaultVersion: %w", err)
	}

	var versions []string
	if err := json.Unmarshal(raw["versions"], &versions); err != nil {
		return config.WebConfig{}, fmt.Errorf("invalid versions array: %w", err)
	}

	cfg.Versions = make(map[string]config.Version)
	for _, ver := range versions {
		if val, ok := raw[ver]; ok {
			var v config.Version
			if err := json.Unmarshal(val, &v); err != nil {
				return config.WebConfig{}, fmt.Errorf("invalid version %q: %w", ver, err)
			}
			cfg.Versions[ver] = v
		}
	}

	return cfg, nil
}

// Función utilitaria para extraer las claves del map
func getKeys(m map[string]config.Version) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
