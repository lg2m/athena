package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the entire app config.
type Config struct {
	Editor EditorConfig `toml:"editor"`
	Keymap KeymapConfig `toml:"keys"`
}

// LoadConfig loads the configuration from default path or arg.
func LoadConfig(filePath *string) (*Config, []string) {
	defaultCfg := defaultConfig()
	var errors []string

	// Load from file and merge
	fileCfg, fileErrors := loadConfigFile(filePath)
	errors = append(errors, fileErrors...)
	mergeConfig(defaultCfg, fileCfg)

	validateErrors := validateAndFixConfig(defaultCfg)
	errors = append(errors, validateErrors...)

	return defaultCfg, errors
}

// defaultConfig provides a default configuration
func defaultConfig() *Config {
	return &Config{
		Editor: EditorConfig{
			ScrollPadding: 5,
			LineNumber:    LineNumberRelative,
			CursorShape: CursorShapeConfig{
				Insert: CursorBar,
				Normal: CursorBlock,
			},
			BufferLine: true,
			Gutters:    []GutterOption{GutterSpacer, GutterLineNumbers, GutterSpacer},
			StatusBar: StatusBarConfig{
				Left:   []StatusBarOption{SectionMode},
				Center: []StatusBarOption{SectionFileName, SectionVersionControl},
				Right:  []StatusBarOption{SectionCursorPercentage, SectionCursorPos, SectionLineCount, SectionFileType},
				Mode: StatusBarModeConfig{
					Normal: "NOR",
					Insert: "INS",
				},
			},
		},
		Keymap: defaultKeymap(),
	}
}

func loadConfigFile(filePath *string) (*Config, []string) {
	var errors []string
	if filePath == nil || *filePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error finding home directory: %v", err))
			return nil, errors
		}
		cfgPath := filepath.Join(homeDir, ".config", "athena", "config.toml")
		filePath = &cfgPath
	}

	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		return nil, errors // No file, no problem
	}

	cfg := &Config{}
	if _, err := toml.DecodeFile(*filePath, cfg); err != nil {
		errors = append(errors, fmt.Sprintf("Error decoding file: %v", err))
	}

	return cfg, errors
}

func mergeConfig(dst *Config, src *Config) {
	if src == nil {
		return
	}
	if src.Editor.ScrollPadding != 0 {
		dst.Editor.ScrollPadding = src.Editor.ScrollPadding
	}
	if src.Editor.LineNumber != "" {
		dst.Editor.LineNumber = src.Editor.LineNumber
	}
	if src.Editor.CursorShape.Insert != "" {
		dst.Editor.CursorShape.Insert = src.Editor.CursorShape.Insert
	}
	if src.Editor.CursorShape.Normal != "" {
		dst.Editor.CursorShape.Normal = src.Editor.CursorShape.Normal
	}
	dst.Editor.BufferLine = src.Editor.BufferLine
	if len(src.Editor.Gutters) > 0 {
		dst.Editor.Gutters = src.Editor.Gutters
	}
	if len(src.Editor.StatusBar.Left) > 0 {
		dst.Editor.StatusBar.Left = src.Editor.StatusBar.Left
	}
	if len(src.Editor.StatusBar.Center) > 0 {
		dst.Editor.StatusBar.Center = src.Editor.StatusBar.Center
	}
	if len(src.Editor.StatusBar.Right) > 0 {
		dst.Editor.StatusBar.Right = src.Editor.StatusBar.Right
	}
	if src.Editor.StatusBar.Mode.Normal != "" {
		dst.Editor.StatusBar.Mode.Normal = src.Editor.StatusBar.Mode.Normal
	}
	if src.Editor.StatusBar.Mode.Insert != "" {
		dst.Editor.StatusBar.Mode.Insert = src.Editor.StatusBar.Mode.Insert
	}
	for key, action := range src.Keymap.Normal {
		dst.Keymap.Normal[key] = action
	}
	for key, action := range src.Keymap.Insert {
		dst.Keymap.Insert[key] = action
	}
}

// validateAndFixConfig validates and ensures the values are in a usable state.
func validateAndFixConfig(cfg *Config) []string {
	var errors []string

	// Validate Editor Config
	editor := &cfg.Editor

	// Validate LineNumber
	if !editor.LineNumber.IsValid() {
		errors = append(errors, fmt.Sprintf("Invalid line-number option: %s", editor.LineNumber))
		editor.LineNumber = LineNumberRelative // Reset to default
	}

	// Validate CursorShape
	if !editor.CursorShape.Insert.IsValid() {
		errors = append(errors, fmt.Sprintf("Invalid cursor-shape insert option: %s", editor.CursorShape.Insert))
		editor.CursorShape.Insert = CursorBar
	}
	if !editor.CursorShape.Normal.IsValid() {
		errors = append(errors, fmt.Sprintf("Invalid cursor-shape normal option: %s", editor.CursorShape.Normal))
		editor.CursorShape.Normal = CursorBlock
	}

	// Validate Gutters
	editor.Gutters = filterValidGutters(editor.Gutters, &errors)

	// Validate StatusBar
	validateStatusBarConfig(&editor.StatusBar, &errors)

	for i := 0; i < len(errors); i++ {
		fmt.Printf("%s\n", errors[i])
	}

	return errors
}

func filterValidGutters(gutters []GutterOption, errors *[]string) []GutterOption {
	var valid []GutterOption
	for _, gutter := range gutters {
		if gutter.IsValid() {
			valid = append(valid, gutter)
		} else {
			*errors = append(*errors, fmt.Sprintf("Invalid gutter option: %s", gutter))
		}
	}
	if len(valid) == 0 {
		return []GutterOption{GutterSpacer, GutterLineNumbers, GutterSpacer} // Default
	}
	return valid
}

func validateStatusBarConfig(statusBar *StatusBarConfig, errors *[]string) {
	// Validate Left sections
	var validLeft []StatusBarOption
	for _, option := range statusBar.Left {
		if option.IsValid() {
			validLeft = append(validLeft, option)
		} else {
			*errors = append(*errors, fmt.Sprintf("Invalid status-bar left option: %s", option))
		}
	}
	statusBar.Left = validLeft

	// Validate Center sections
	var validCenter []StatusBarOption
	for _, option := range statusBar.Center {
		if option.IsValid() {
			validCenter = append(validCenter, option)
		} else {
			*errors = append(*errors, fmt.Sprintf("Invalid status-bar center option: %s", option))
		}
	}
	statusBar.Center = validCenter

	// Validate Right sections
	var validRight []StatusBarOption
	for _, option := range statusBar.Right {
		if option.IsValid() {
			validRight = append(validRight, option)
		} else {
			*errors = append(*errors, fmt.Sprintf("Invalid status-bar right option: %s", option))
		}
	}
	statusBar.Right = validRight

	// If sections are empty, set defaults
	if len(statusBar.Left) == 0 {
		statusBar.Left = []StatusBarOption{SectionMode}
	}
	if len(statusBar.Center) == 0 {
		statusBar.Center = []StatusBarOption{SectionFileName, SectionVersionControl}
	}
	if len(statusBar.Right) == 0 {
		statusBar.Right = []StatusBarOption{SectionCursorPercentage, SectionCursorPos, SectionLineCount, SectionFileType}
	}
}
