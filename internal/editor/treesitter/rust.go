package treesitter

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
)

type RustLanguage struct{}

func (r *RustLanguage) Language() *sitter.Language {
	return sitter.NewLanguage(rust.Language())
}

func (r *RustLanguage) Name() string {
	return "rust"
}
