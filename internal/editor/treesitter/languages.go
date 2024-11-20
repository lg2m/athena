package treesitter

import (
	"embed"
	"fmt"
	"io/fs"

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

// LanguageProvider defines and interface for langauge support.
type LanguageProvider interface {
	Language() *sitter.Language
	Name() string
}

// Registry manages supported langauges and their configurations.
type Registry struct {
	languages map[string]LanguageProvider
	queries   map[string]map[QueryType]*sitter.Query
	styles    StyleMap
}

// NewRegistry creates a new language registry with default settings.
func NewRegistry(defaultStyles StyleMap) *Registry {
	return &Registry{
		languages: make(map[string]LanguageProvider),
		queries:   make(map[string]map[QueryType]*sitter.Query),
		styles:    defaultStyles,
	}
}

// RegisterLanguage adds a new language to the registry.
func (r *Registry) RegisterLanguage(provider LanguageProvider) error {
	name := provider.Name()
	r.languages[name] = provider

	// Initialize queries map for this language
	r.queries[name] = make(map[QueryType]*sitter.Query)

	// Load all query types
	queryTypes := []QueryType{QueryHighlights, QueryInjections, QueryLocals, QueryOutline}
	for _, qt := range queryTypes {
		if query, err := r.loadQuery(name, qt); err == nil {
			r.queries[name][qt] = query
		}
	}

	return nil
}

// loadQuery loads a query file for a specific language and query type
func (r *Registry) loadQuery(lang string, queryType QueryType) (*sitter.Query, error) {
	path := fmt.Sprintf("runtime/queries/%s/%s.scm", lang, queryType)
	content, err := fs.ReadFile(queriesFS, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load query file %s: %w", path, err)
	}

	language := r.languages[lang].Language()
	query, queryErr := sitter.NewQuery(language, string(content))
	if queryErr != nil {
		return nil, fmt.Errorf("%s", queryErr.Error())
	}

	return query, nil
}
