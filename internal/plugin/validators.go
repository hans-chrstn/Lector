package plugin

import (
	"fmt"
	"github.com/yuin/gopher-lua"
)

type PluginValidator interface {
	Validate(p *LuaPlugin) error
}

type SourceValidator struct{}

func (v *SourceValidator) Validate(p *LuaPlugin) error {
	hasSourceFuncs := true
	for _, fn := range []string{"search", "get_document", "get_chapter"} {
		p.Mu.Lock()
		f := p.L.GetGlobal(fn)
		p.Mu.Unlock()
		if f.Type() != lua.LTFunction {
			hasSourceFuncs = false
			break
		}
	}

	if !hasSourceFuncs {
		return fmt.Errorf("source plugin must implement 'search', 'get_document', and 'get_chapter'")
	}
	return nil
}

type UtilityValidator struct{}

func (v *UtilityValidator) Validate(p *LuaPlugin) error {

	p.ManifestMu.RLock()
	hasUI := len(p.Tabs) > 0 || len(p.Actions) > 0 || len(p.SettingsGroups) > 0 || len(p.Sections) > 0 || len(p.UIOverrides) > 0
	hasUtilityCap := false
	for _, cap := range p.Capabilities {
		if cap == "background" || cap == "storage" || cap == "theming" {
			hasUtilityCap = true
			break
		}
	}
	p.ManifestMu.RUnlock()

	if !hasUI && !hasUtilityCap {
		return fmt.Errorf("utility plugin must define UI elements or enable 'background'/'storage' capabilities")
	}
	return nil
}

func GetValidator(pluginType string) PluginValidator {
	switch pluginType {
	case "utility":
		return &UtilityValidator{}
	case "source":
		fallthrough
	default:
		return &SourceValidator{}
	}
}
