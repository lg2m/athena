package config

// LineNumberOption represents how way to display line numbers.
type LineNumberOption string

const (
	LineNumberAbsolute LineNumberOption = "absolute"
	LineNumberRelative LineNumberOption = "relative"
)

func (o LineNumberOption) IsValid() bool {
	switch o {
	case LineNumberAbsolute, LineNumberRelative:
		return true
	default:
		return false
	}
}

// CursorShape defines cursor style options.
type CursorShape string

const (
	CursorBar   CursorShape = "bar"
	CursorBlock CursorShape = "block"
	CursorLine  CursorShape = "line"
	CursorUnder CursorShape = "underline"
)

func (cs CursorShape) IsValid() bool {
	switch cs {
	case CursorBar, CursorBlock, CursorLine, CursorUnder:
		return true
	default:
		return false
	}
}

// CursorShapeConfig holds cursor shape settings.
type CursorShapeConfig struct {
	Insert CursorShape `toml:"insert"`
	Normal CursorShape `toml:"normal"`
}

// GutterLayoutOption defines layout parts for gutters.
type GutterOption string

const (
	GutterDiff        GutterOption = "diff"
	GutterLineNumbers GutterOption = "line-numbers"
	GutterSpacer      GutterOption = "spacer"
)

func (o GutterOption) IsValid() bool {
	switch o {
	case GutterDiff, GutterLineNumbers, GutterSpacer:
		return true
	default:
		return false
	}
}

// StatusBarOption defines valid types for status bar sections.
type StatusBarOption string

const (
	SectionMode             StatusBarOption = "mode"
	SectionFileName         StatusBarOption = "file-name"
	SectionFileAbsPath      StatusBarOption = "file-absolute-path"
	SectionFileModified     StatusBarOption = "file-modified"
	SectionFileEncoding     StatusBarOption = "file-encoding"
	SectionFileType         StatusBarOption = "file-type"
	SectionVersionControl   StatusBarOption = "version-control"
	SectionCursorPos        StatusBarOption = "cursor-position"
	SectionLineCount        StatusBarOption = "line-count"
	SectionCursorPercentage StatusBarOption = "cursor-percentage"
	SectionSpacer           StatusBarOption = "spacer"
)

func (o StatusBarOption) IsValid() bool {
	switch o {
	case SectionMode, SectionFileName, SectionFileAbsPath, SectionFileModified,
		SectionFileEncoding, SectionFileType, SectionVersionControl,
		SectionCursorPos, SectionLineCount, SectionCursorPercentage, SectionSpacer:
		return true
	default:
		return false
	}
}

// StatusBarModeConfig represents the mode names.
type StatusBarModeConfig struct {
	Normal string `toml:"normal"`
	Insert string `toml:"insert"`
}

// StatusBarConfig represents status bar configurations.
type StatusBarConfig struct {
	Left   []StatusBarOption   `toml:"left"`
	Center []StatusBarOption   `toml:"center"`
	Right  []StatusBarOption   `toml:"right"`
	Mode   StatusBarModeConfig `toml:"mode"`
}

// EditorConfig represents editor-specific configurations
type EditorConfig struct {
	ScrollPadding int               `toml:"scroll-padding"` // padding around edge of screen
	LineNumber    LineNumberOption  `toml:"line-number"`    // absolute or relative
	CursorShape   CursorShapeConfig `toml:"cursor-shape"`
	BufferLine    bool              `toml:"buffer-line"` // whether to render buffer line
	Gutters       []GutterOption    `toml:"gutters"`
	StatusBar     StatusBarConfig   `toml:"status-bar"`
}
