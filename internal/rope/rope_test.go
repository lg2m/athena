package rope

import "testing"

func TestNewRope(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"Hello, World!"},
		{"こんにちは世界"}, // Japanese for "Hello, World"
		{"👋🌍"},      // Emojis
		{"A🇺🇳B"},    // Combining characters
		{""},        // Empty string
	}

	for _, tt := range tests {
		rope := NewRope(tt.input)
		if rope.ToString() != tt.input {
			t.Errorf("NewRope failed for input %q: expected %q, got %q", tt.input, tt.input, rope.ToString())
		}
		if rope.TotalGraphemes() != CountGraphemes(tt.input) {
			t.Errorf("NewRope grapheme count mismatch for input %q: expected %d, got %d",
				tt.input, CountGraphemes(tt.input), rope.TotalGraphemes())
		}
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		initial   string
		insertPos int
		toInsert  string
		expected  string
	}{
		{"Hello, World!", 7, "Beautiful ", "Hello, Beautiful World!"},
		{"こんにちは世界", 5, "！", "こんにちは！世界"},
		{"👋🌍", 1, "😊", "👋😊🌍"},
		{"A🇺🇳B", 2, "C", "A🇺🇳CB"},
		{"", 0, "Start", "Start"},
	}

	for _, tt := range tests {
		rope := NewRope(tt.initial)
		err := rope.Insert(tt.insertPos, tt.toInsert)
		if err != nil {
			t.Errorf("Insert failed for initial %q at pos %d with %q: %v",
				tt.initial, tt.insertPos, tt.toInsert, err)
			continue
		}
		if rope.ToString() != tt.expected {
			t.Errorf("Insert result mismatch: expected %q, got %q", tt.expected, rope.ToString())
		}
		if rope.TotalGraphemes() != CountGraphemes(tt.expected) {
			t.Errorf("After insert, grapheme count mismatch: expected %d, got %d",
				CountGraphemes(tt.expected), rope.TotalGraphemes())
		}
	}
}

func TestInsertOutOfBounds(t *testing.T) {
	rope := NewRope("Test")
	err := rope.Insert(-1, "Invalid")
	if err != ErrOutOfBounds {
		t.Errorf("Expected ErrOutOfBounds for negative index, got %v", err)
	}

	err = rope.Insert(5, "Invalid") // len("Test") is 4 graphemes
	if err != ErrOutOfBounds {
		t.Errorf("Expected ErrOutOfBounds for index beyond length, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		initial  string
		start    int
		end      int
		expected string
	}{
		// Adjusted end index to include the space after "Beautiful"
		{"Hello, Beautiful World!", 7, 17, "Hello, World!"},

		// These test cases are correct as they do not involve trailing spaces
		{"こんにちは！世界", 5, 6, "こんにちは世界"},
		{"👋😊🌍", 1, 2, "👋🌍"},
		{"A🇺🇳CB", 1, 3, "AB"},

		// Adjusted end index to include the space after "Start"
		{"Start and End", 0, 6, "and End"},

		// These test cases are correct
		{"Start and End", 10, 13, "Start and "},
		{"All to delete", 0, CountGraphemes("All to delete"), ""},
	}

	for _, tt := range tests {
		rope := NewRope(tt.initial)
		err := rope.Delete(tt.start, tt.end)
		if err != nil {
			t.Errorf("Delete failed for initial %q from %d to %d: %v",
				tt.initial, tt.start, tt.end, err)
			continue
		}
		if rope.ToString() != tt.expected {
			t.Errorf("Delete result mismatch: expected %q, got %q", tt.expected, rope.ToString())
		}
		if rope.TotalGraphemes() != CountGraphemes(tt.expected) {
			t.Errorf("After delete, grapheme count mismatch: expected %d, got %d",
				CountGraphemes(tt.expected), rope.TotalGraphemes())
		}
	}
}

func TestDeleteInvalidRange(t *testing.T) {
	rope := NewRope("Test")
	err := rope.Delete(-1, 2)
	if err != ErrInvalidRange {
		t.Errorf("Expected ErrInvalidRange for negative start, got %v", err)
	}

	err = rope.Delete(1, 5) // len("Test") is 4 graphemes
	if err != ErrInvalidRange {
		t.Errorf("Expected ErrInvalidRange for end beyond length, got %v", err)
	}

	err = rope.Delete(3, 2) // start > end
	if err != ErrInvalidRange {
		t.Errorf("Expected ErrInvalidRange for start > end, got %v", err)
	}
}

func TestGetTextRange(t *testing.T) {
	tests := []struct {
		initial  string
		start    int
		end      int
		expected string
	}{
		{"Hello, Beautiful World!", 7, 16, "Beautiful"},
		{"こんにちは世界", 0, 5, "こんにちは"},
		{"👋😊🌍", 1, 3, "😊🌍"},
		{"A🇺🇳CB", 1, 2, "🇺🇳"},
		{"Start and End", 6, 9, "and"},
	}

	for _, tt := range tests {
		rope := NewRope(tt.initial)
		substr, err := rope.GetTextRange(tt.start, tt.end)
		if err != nil {
			t.Errorf("GetTextRange failed for initial %q from %d to %d: %v",
				tt.initial, tt.start, tt.end, err)
			continue
		}
		if substr != tt.expected {
			t.Errorf("GetTextRange mismatch: expected %q, got %q", tt.expected, substr)
		}
	}
}

func TestGetTextRangeInvalid(t *testing.T) {
	rope := NewRope("Test")
	_, err := rope.GetTextRange(-1, 2)
	if err == nil {
		t.Error("Expected error for negative start index, got nil")
	}

	_, err = rope.GetTextRange(1, 5) // len("Test") is 4 graphemes
	if err == nil {
		t.Error("Expected error for end index beyond length, got nil")
	}

	_, err = rope.GetTextRange(3, 2) // start > end
	if err == nil {
		t.Error("Expected error for start > end, got nil")
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"Hello, World!"},
		{"こんにちは世界"},
		{"👋😊🌍"},
		{"A🇺🇳B"},
		{""},
	}

	for _, tt := range tests {
		rope := NewRope(tt.input)
		if rope.ToString() != tt.input {
			t.Errorf("ToString mismatch: expected %q, got %q", tt.input, rope.ToString())
		}
	}
}

func TestTotalGraphemes(t *testing.T) {
	tests := []struct {
		input         string
		expectedCount int
	}{
		{"Hello", 5},
		{"こんにちは", 5},
		{"👋😊🌍", 3},
		{"A🇺🇳B", 3}, // '🇺🇳' is a single grapheme cluster
		{"", 0},     // Empty string
	}

	for _, tt := range tests {
		rope := NewRope(tt.input)
		count := rope.TotalGraphemes()
		if count != tt.expectedCount {
			t.Errorf("TotalGraphemes mismatch for input %q: expected %d, got %d",
				tt.input, tt.expectedCount, count)
		}
	}
}
