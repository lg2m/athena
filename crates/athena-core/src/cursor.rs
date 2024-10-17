use ropey::{Rope, RopeSlice};

use crate::GraphemeOperations;

// TODO: fixme.. major issues with movements lookinto the grapheme impl and cursor

#[derive(Clone, Debug)]
pub struct Cursor {
    pub index: usize,
}

impl Cursor {
    pub fn new() -> Self {
        Self { index: 0usize }
    }

    /// Move to the previous grapheme cluster boundary.
    pub fn move_prev_grapheme(&mut self, buffer: &RopeSlice) {
        self.index = buffer.prev_grapheme_boundary(self.index);
    }

    /// Move to the previous word boundary.
    pub fn move_prev_word(&mut self, buffer: &RopeSlice) {
        self.index = buffer.prev_word_boundary(self.index);
    }

    /// Move to the previous line boundary.
    pub fn move_prev_line(&mut self, buffer: &RopeSlice) {
        let line_idx = buffer.char_to_line(self.index);
        if line_idx > 0 {
            self.index = buffer.line_to_char(line_idx - 1);
        } else {
            self.index = 0;
        }
    }

    /// Move to the next grapheme cluster boundary.
    pub fn move_next_grapheme(&mut self, buffer: &RopeSlice) {
        self.index = buffer.next_grapheme_boundary(self.index);
    }

    /// Move to the next word boundary.
    pub fn move_next_word(&mut self, buffer: &RopeSlice) {
        self.index = buffer.next_word_boundary(self.index);
    }

    /// Move to the next line boundary.
    pub fn move_next_line(&mut self, rope: &Rope) {
        let line_idx = rope.char_to_line(self.index);
        if line_idx + 1 < rope.len_lines() {
            self.index = rope.line_to_char(line_idx + 1);
        } else {
            self.index = rope.len_chars();
        }
    }

    /// Move to the end of the current line
    pub fn move_to_end_of_line(&mut self, rope: &Rope) {
        let line_idx = rope.char_to_line(self.index);
        let line = rope.line(line_idx);
        let line_len = line.len_chars();

        // Exclude the newline character at the end of the line
        let line_end_index = rope.line_to_char(line_idx) + line_len.saturating_sub(1);
        self.index = line_end_index;
    }
}

#[derive(Clone, Copy, Debug, Eq, PartialEq, Hash)]
pub enum SelectionScope {
    Grapheme,
    Word,
    Line,
    // Paragraph,
}

#[derive(Clone, Debug)]
pub struct Selection {
    pub start: usize,
    pub end: usize,
}

impl Selection {
    pub fn new() -> Self {
        Self {
            start: 0usize,
            end: 0usize,
        }
    }

    /// Check if a selection is active.
    #[inline]
    pub fn is_active(&self) -> bool {
        self.start != self.end
    }

    /// Clear selection.
    #[inline]
    pub fn clear(&mut self) {
        self.start = 0;
        self.end = 0;
    }

    /// Set the selection range.
    #[inline]
    pub fn set(&mut self, start: usize, end: usize) {
        self.start = start;
        self.end = end;
    }

    pub fn select_to_prev_word(&mut self, cursor: &Cursor, buffer: &RopeSlice) {
        let start = buffer.prev_word_boundary(cursor.index);
        self.set(start, cursor.index);
    }

    pub fn select_to_next_word(&mut self, cursor: &Cursor, buffer: &RopeSlice) {
        let end = buffer.next_word_boundary(cursor.index);
        self.set(cursor.index, end);
    }

    pub fn select_to_end_of_line(&mut self, cursor: &Cursor, rope: &Rope) {
        let line_idx = rope.char_to_line(cursor.index);
        let line_end = rope.line_to_char(line_idx + 1).min(rope.len_chars());
        self.set(cursor.index, line_end);
    }

    pub fn ensure_grapheme_boundaries(&mut self, buffer: &RopeSlice) {
        if !buffer.is_grapheme_boundary(self.start) {
            self.start = buffer.prev_grapheme_boundary(self.start);
        }
        if !buffer.is_grapheme_boundary(self.end) {
            self.end = buffer.next_grapheme_boundary(self.end);
        }
    }

    pub fn select_scope(&mut self, cursor: &Cursor, scope: &SelectionScope, rope: &Rope) {
        match scope {
            SelectionScope::Grapheme => {
                let end = rope.slice(..).next_grapheme_boundary(cursor.index);
                self.set(cursor.index, end);
            }
            SelectionScope::Word => {
                let end = rope.slice(..).next_word_boundary(cursor.index);
                self.set(cursor.index, end);
            }
            SelectionScope::Line => self.select_to_end_of_line(cursor, rope),
        }
        self.ensure_grapheme_boundaries(&rope.slice(..));
    }
}
