package languages

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

// GoProvider implements the LanguageProvider interface for Go.
type GoProvider struct{}

// Language returns the Tree-sitter Go language implementation.
func (g GoProvider) Language() *sitter.Language {
	return sitter.NewLanguage(tree_sitter_go.Language())
}

// Name returns the name of the Go language.
func (g GoProvider) Name() string {
	return "go"
}

// Extensions returns the file extensions associated with Go.
func (g GoProvider) Extensions() []string {
	return []string{"go"}
}
