package treesitter

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Highlight represents a syntax highlighting span.
type Highlight struct {
	Start Position
	End   Position
	Style tcell.Style
}

// Position represents a position in the source code.
type Position struct {
	Row    uint32
	Column uint32
}

// Highlighter represents a syntax highlighter for a specific language.
type Highlighter struct {
	parser   *sitter.Parser
	language LanguageProvider
	registry *Registry

	query  *sitter.Query
	styles StyleMap
}

// NewHighlighter creates a new syntax highlighter for the specified language.
func NewHighlighter(registry *Registry, languageName string) (*Highlighter, error) {
	lang, exists := registry.languages[languageName]
	if !exists {
		return nil, fmt.Errorf("hl: unsupported langauge: %s", languageName)
	}

	parser := sitter.NewParser()
	if err := parser.SetLanguage(lang.Language()); err != nil {
		return nil, fmt.Errorf("hl: failed to set language: %w", err)
	}

	return &Highlighter{
		parser:   parser,
		language: lang,
		registry: registry,
	}, nil
}

// GetHighlights returns syntax highlighting information for the given code.
func (h *Highlighter) GetHighlights(code []byte) ([]Highlight, error) {
	tree := h.parser.Parse(code, nil)
	defer tree.Close()

	query := h.registry.queries[h.language.Name()][QueryHighlights]
	if query == nil {
		return nil, fmt.Errorf("hl: no highlights query available for %s", h.language.Name())
	}

	qc := sitter.NewQueryCursor()
	defer qc.Close()

	matches := qc.Matches(query, tree.RootNode(), code)

	var highlights []Highlight
	for match := matches.Next(); match != nil; match = matches.Next() {
		for _, capture := range match.Captures {
			captureName := query.CaptureNames()[capture.Index]
			if style, ok := h.registry.styles[captureName]; ok {
				node := capture.Node
				startPos := node.StartPosition()
				endPos := node.EndPosition()
				highlight := Highlight{
					Start: Position{
						Row:    uint32(startPos.Row),
						Column: uint32(startPos.Column),
					},
					End: Position{
						Row:    uint32(endPos.Row),
						Column: uint32(endPos.Column),
					},
					Style: style,
				}
				highlights = append(highlights, highlight)
			}
		}
	}

	return highlights, nil
}
