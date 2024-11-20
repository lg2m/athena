package treesitter

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// StyleMap maps node types to tcell styles
type StyleMap map[string]tcell.Style

// Embed query files
//
//go:embed runtime/queries/*/*.scm
var queriesFS embed.FS

// QueryType represents different types of tree-sitter queries.
type QueryType string

const (
	QueryHighlights QueryType = "highlights"
	QueryInjections QueryType = "injections"
	QueryLocals     QueryType = "locals"
	QueryOutline    QueryType = "outline"
)

// LanguageProvider defines an interface for language support.
type LanguageProvider interface {
	Language() *sitter.Language
	Name() string
	Extensions() []string
}

// Registry manages supported languages and their configurations.
type Registry struct {
	languages map[string]LanguageProvider
	queries   map[string]map[QueryType]*sitter.Query
	styles    StyleMap
}

// NewRegistry creates a new language registry with default settings.
func NewRegistry() *Registry {
	return &Registry{
		languages: make(map[string]LanguageProvider),
		queries:   make(map[string]map[QueryType]*sitter.Query),
		styles:    DefaultStyles,
	}
}

// RegisterLanguage adds a new language provider to the registry.
func (r *Registry) RegisterLanguage(provider LanguageProvider) error {
	r.languages[provider.Name()] = provider
	for _, ext := range provider.Extensions() {
		r.languages[ext] = provider
	}

	// Load queries for this language
	queryTypes := []QueryType{QueryHighlights, QueryInjections, QueryLocals, QueryOutline}
	queryMap := make(map[QueryType]*sitter.Query)

	for _, queryType := range queryTypes {
		queryPath := fmt.Sprintf("runtime/queries/%s/%s.scm", provider.Name(), queryType)
		queryContent, err := queriesFS.ReadFile(queryPath)
		if err != nil {
			if queryType == QueryHighlights {
				return fmt.Errorf("missing essential query file: %s", queryPath)
			}
			continue
		}

		query, queryErr := sitter.NewQuery(provider.Language(), string(queryContent))
		if queryErr != nil {
			return fmt.Errorf("failed to load query %s: %s", queryPath, queryErr.Error())
		}
		queryMap[queryType] = query
	}

	r.queries[provider.Name()] = queryMap
	return nil
}

// DetectLanguage detects the language from the filename.
func (r *Registry) DetectLanguage(filename string) (string, error) {
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	for name, provider := range r.languages {
		for _, langExt := range provider.Extensions() {
			if langExt == ext {
				return name, nil
			}
		}
	}
	return "", fmt.Errorf("unsupported file extension: %s", ext)
}
