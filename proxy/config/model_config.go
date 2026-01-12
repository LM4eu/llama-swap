package config

import (
	"errors"
	"runtime"
	"slices"
	"strings"
)

type ModelConfig struct {
	Metadata         map[string]any `yaml:"metadata"`
	SendLoadingState *bool          `yaml:"sendLoadingState"`
	Filters          ModelFilters   `yaml:"filters"`
	Cmd              string         `yaml:"cmd"`
	CmdStop          string         `yaml:"cmdStop"`
	CheckEndpoint    string         `yaml:"checkEndpoint"`
	Proxy            string         `yaml:"proxy"`
	Description      string         `yaml:"description"`
	UseModelName     string         `yaml:"useModelName"`
	Name             string         `yaml:"name"`
	Aliases          []string       `yaml:"aliases"`
	Macros           MacroList      `yaml:"macros"`
	Env              []string       `yaml:"env"`
	ConcurrencyLimit int            `yaml:"concurrencyLimit"`
	UnloadAfter      int            `yaml:"ttl"`
	Unlisted         bool           `yaml:"unlisted"`
}

func (m *ModelConfig) UnmarshalYAML(unmarshal func(any) error) error {
	type rawModelConfig ModelConfig
	defaults := rawModelConfig{
		Cmd:              "",
		CmdStop:          "",
		Proxy:            "http://localhost:${PORT}",
		Aliases:          []string{},
		Env:              []string{},
		CheckEndpoint:    "/health",
		UnloadAfter:      0,
		Unlisted:         false,
		UseModelName:     "",
		ConcurrencyLimit: 0,
		Name:             "",
		Description:      "",
	}

	// the default cmdStop to taskkill /f /t /pid ${PID}
	if runtime.GOOS == "windows" {
		defaults.CmdStop = "taskkill /f /t /pid ${PID}"
	}

	err := unmarshal(&defaults)
	if err != nil {
		return err
	}

	*m = ModelConfig(defaults)
	return nil
}

func (m *ModelConfig) SanitizedCommand() ([]string, error) {
	return SanitizeCommand(m.Cmd)
}

// ModelFilters see issue #174.
type ModelFilters struct {
	StripParams string `yaml:"stripParams"`
}

func (m *ModelFilters) UnmarshalYAML(unmarshal func(any) error) error {
	type rawModelFilters ModelFilters
	defaults := rawModelFilters{
		StripParams: "",
	}

	err := unmarshal(&defaults)
	if err != nil {
		return err
	}

	// Try to unmarshal with the old field name for backwards compatibility
	if defaults.StripParams == "" {
		var legacy struct {
			StripParams string `yaml:"strip_params"`
		}
		legacyErr := unmarshal(&legacy)
		if legacyErr != nil {
			return errors.New("failed to unmarshal legacy filters.strip_params: " + legacyErr.Error())
		}
		defaults.StripParams = legacy.StripParams
	}

	*m = ModelFilters(defaults)
	return nil
}

func (f ModelFilters) SanitizedStripParams() ([]string, error) {
	if f.StripParams == "" {
		return nil, nil
	}

	params := strings.Split(f.StripParams, ",")
	cleaned := make([]string, 0, len(params))
	seen := make(map[string]bool)

	for _, param := range params {
		trimmed := strings.TrimSpace(param)
		if trimmed == "model" || trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		cleaned = append(cleaned, trimmed)
	}

	// sort cleaned
	slices.Sort(cleaned)
	return cleaned, nil
}
