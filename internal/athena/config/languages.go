package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type LanguagesConfig struct {
	Languages map[string]LanguageConfig `toml:"langauges"`
}

type LanguageConfig struct {
	AltNames           []string          `toml:"alt_names"`
	MimeTypes          []string          `toml:"mime_types"`
	FileTypes          []string          `toml:"file_types"`
	Files              []string          `toml:"files"`
	LineCommentTokens  []string          `toml:"line_comment_tokens"`
	BlockCommentTokens []CommentToken    `toml:"block_comment_tokens"`
	AutoPairs          []AutoPair        `toml:"auto_pairs"`
	Grammar            GrammarDefinition `toml:"grammar"`
}

type CommentToken struct {
	Start string `toml:"start"`
	End   string `toml:"end"`
}

type AutoPair struct {
	Open  string `toml:"open"`
	Close string `toml:"close"`
}

type GrammarDefinition struct {
	Name       string         `toml:"name"`
	SymbolName string         `toml:"symbol_name"`
	Install    InstallOptions `toml:"install"`
}

type InstallOptions struct {
	Git     string `toml:"git"`
	Rev     string `toml:"rev"`
	Ref     string `toml:"ref"`
	RefType string `toml:"ref_type"`
}

// LoadLanguagesConfig loads the configuration from default path or arg.
func LoadLanguagesConfig(filePath *string) (*LanguagesConfig, []string) {
	var errors []string

	// Load from file and merge
	fileCfg, fileErrors := loadLanguagesConfigFile(filePath)
	errors = append(errors, fileErrors...)

	// validateErrors := validateAndFixConfig(defaultCfg)
	// errors = append(errors, validateErrors...)

	return fileCfg, errors
}

func loadLanguagesConfigFile(filePath *string) (*LanguagesConfig, []string) {
	var errors []string
	if filePath == nil || *filePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error finding home directory: %v", err))
			return nil, errors
		}
		cfgPath := filepath.Join(homeDir, ".config", "athena", "languages.toml")
		filePath = &cfgPath
	}

	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		return nil, errors // No file, no problem
	}

	cfg := &LanguagesConfig{}
	if _, err := toml.DecodeFile(*filePath, cfg); err != nil {
		errors = append(errors, fmt.Sprintf("Error decoding file: %v", err))
	}

	return cfg, errors
}
