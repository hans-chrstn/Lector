package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Manifest struct {
	Type         string   `json:"type"`
	Capabilities []string `json:"capabilities"`
}

func ValidateManifest(path string) error {
	pluginDir := filepath.Dir(path)
	manifestPath := filepath.Join(pluginDir, "manifest.json")

	if _, err := os.Stat(manifestPath); err != nil {
		return nil
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("invalid manifest format: %w", err)
	}

	for _, cap := range m.Capabilities {
		switch cap {
		case "network", "storage", "ui", "theming", "source", "catalog", "background", "interaction", "local_network", "global_documents":
			continue
		default:
			return fmt.Errorf("unknown capability: %s", cap)
		}
	}
	return nil
}

type Validator interface {
	Validate(s *LuaPlugin) error
}

type sourceValidator struct{}

func (v *sourceValidator) Validate(s *LuaPlugin) error {
	return nil
}

type utilityValidator struct{}

func (v *utilityValidator) Validate(s *LuaPlugin) error {
	return nil
}

func GetValidator(pluginType string) Validator {
	if pluginType == "source" {
		return &sourceValidator{}
	}
	return &utilityValidator{}
}
