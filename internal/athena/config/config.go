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
	cfg := defaultConfig()
	var errors []string

	var fileCfg *Config
	var fileErrors []string
	var err error

	if filePath != nil && *filePath != "" {
		fileCfg, fileErrors, err = loadConfigFromFile(*filePath)
	} else {
		homeDir, errs := os.UserHomeDir()
		if errs != nil {
			errors = append(errors, fmt.Sprintf("Error getting home directory: %v", errs))
			return cfg, errors
		}

		cfgPath := filepath.Join(homeDir, ".config", "athena", "config.toml")
		fileCfg, fileErrors, err = loadConfigFromFile(cfgPath)
	}

	if err != nil {
		errors = append(errors, fmt.Sprintf("Error loading config file: %v", err))
	} else {
		errors = append(errors, fileErrors...)
		if fileCfg != nil {
			applyConfig(cfg, fileCfg)
		}
	}

	return cfg, errors
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
		Keymap: KeymapConfig{
			Normal: map[string]interface{}{
				"i": "enter-insert-mode",
				"j": "move-down",
				"k": "move-up",
				"h": "move-left",
				"l": "move-right",
			},
			Insert: map[string]interface{}{
				"<Esc>": "enter-normal-mode",
			},
		},
	}
}

// loadConfigFromFile loads the configuration from the given file path
func loadConfigFromFile(filePath string) (*Config, []string, error) {
	var errors []string

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors, nil
	}

	cfg := &Config{}
	if _, err := toml.DecodeFile(filePath, cfg); err != nil {
		return nil, errors, fmt.Errorf("error decoding config file: %v", err)
	}

	errors = validateConfig(cfg)

	return cfg, errors, nil
}

// validateConfig validates and ensures the values are in a usable state.
func validateConfig(cfg *Config) []string {
	var errors []string

	// Validate Editor Config
	editor := &cfg.Editor

	// Validate LineNumber
	if editor.LineNumber != "" && !editor.LineNumber.IsValid() {
		errors = append(errors, fmt.Sprintf("Invalid line-number option: %s", editor.LineNumber))
		editor.LineNumber = LineNumberRelative // Reset to default
	}

	// Validate CursorShape
	if editor.CursorShape.Insert != "" && !editor.CursorShape.Insert.IsValid() {
		errors = append(errors, fmt.Sprintf("Invalid cursor-shape insert option: %s", editor.CursorShape.Insert))
		editor.CursorShape.Insert = CursorBar // Reset to default
	}
	if editor.CursorShape.Normal != "" && !editor.CursorShape.Normal.IsValid() {
		errors = append(errors, fmt.Sprintf("Invalid cursor-shape normal option: %s", editor.CursorShape.Normal))
		editor.CursorShape.Normal = CursorBlock // Reset to default
	}

	// Validate Gutters
	var validGutters []GutterOption
	for _, gutter := range editor.Gutters {
		if gutter.IsValid() {
			validGutters = append(validGutters, gutter)
		} else {
			errors = append(errors, fmt.Sprintf("Invalid gutter option: %s", gutter))
		}
	}
	editor.Gutters = validGutters
	if len(editor.Gutters) == 0 {
		editor.Gutters = []GutterOption{GutterSpacer, GutterLineNumbers, GutterSpacer} // Reset to default if empty
	}

	// Validate StatusBar Options
	validateStatusBarConfig(&editor.StatusBar, &errors)

	return errors
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

func applyConfig(dst *Config, src *Config) {
	if src == nil {
		return
	}

	// Apply Editor config
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

	// Apply Keymap config
	if len(src.Keymap.Normal) > 0 {
		dst.Keymap.Normal = src.Keymap.Normal
	}
	if len(src.Keymap.Insert) > 0 {
		dst.Keymap.Insert = src.Keymap.Insert
	}
}
