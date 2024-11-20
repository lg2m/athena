package languages

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
)

// RustProvider implements the LanguageProvider interface for Rust.
type RustProvider struct{}

// Language returns the Tree-sitter Rust language implementation.
func (r *RustProvider) Language() *sitter.Language {
	return sitter.NewLanguage(tree_sitter_rust.Language())
}

// Name returns the name of the Rust language.
func (r *RustProvider) Name() string {
	return "rust"
}

// Extensions returns the file extensions associated with Rust.
func (r RustProvider) Extensions() []string {
	return []string{"rs"}
}
