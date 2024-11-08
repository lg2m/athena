use ropey::{Rope, RopeSlice};

use crate::{
    config::Mode,
    cursor::{Cursor, Selection},
    graphemes::GraphemeOperations,
};

/// Editor events that occur after a user performs an action or triggers \
/// a command in the terminal.
#[derive(Debug, Eq, PartialEq, Hash)]
pub enum EditorEvent {
    /// Cursor position updated.
    CursorMoved(usize, usize),
    Char(char),
    /// Editor mode updated.
    ModeChanged(Mode),
    /// Text buffer update.
    BufferChanged,
    /// Terminal / window size change.
    ViewportChanged,
}

#[derive(Copy, Clone, Debug, PartialEq, Eq)]
pub enum Granularity {
    Character,
    Word,
    Line,
}

/// Direction for movement.
#[derive(Clone, Copy, Debug, Eq, PartialEq, Hash)]
pub enum Direction {
    Forward,
    Backward,
}

#[derive(Clone, Debug)]
pub struct EditorState {
    pub buffer: Rope,
    pub cursor: Cursor,
    pub selection: Selection,
    pub mode: Mode,
}

impl EditorState {
    pub fn new() -> Self {
        Self {
            buffer: Rope::from_str("Welcome to Athena, a modern terminal text-editor"),
            cursor: Cursor::new(),
            selection: Selection::new(),
            mode: Mode::Normal,
        }
    }

    /// Append after the cursor position and enter insert mode.
    pub fn append(&mut self) {
        if self.mode == Mode::Normal {
            self.cursor.move_next_grapheme(&self.buffer.slice(..));
            self.mode = Mode::Insert;
        }
    }

    /// Go to the beginning of the current line and enter insert mode.
    pub fn insert_start_of_line(&mut self) {
        if self.mode == Mode::Normal {
            // TODO: implement
            // self.cursor.move_to_beginning_of_line(&self.buffer);
            self.cursor.move_prev_line(&self.buffer.slice(..));
            self.cursor.move_next_line(&self.buffer);
            self.mode = Mode::Insert;
        }
    }

    /// Go to the end of the current line and enter insert mode.
    pub fn append_end_of_line(&mut self) {
        if self.mode == Mode::Normal {
            self.cursor.move_to_end_of_line(&self.buffer);
            self.mode = Mode::Insert;
        }
    }

    /// Insert a newline below the current line, move cursor, and enter insert mode.
    pub fn insert_newline_below(&mut self) {
        if self.mode == Mode::Normal {
            self.cursor.move_to_end_of_line(&self.buffer);
            self.buffer.insert_char(self.cursor.index, '\n');
            let slice = self.buffer.slice(..);
            self.cursor.index = slice.next_grapheme_boundary(self.cursor.index);
            self.mode = Mode::Insert;
        }
    }

    /// Insert a newline above the current line, move cursor, and enter insert mode.
    pub fn insert_newline_above(&mut self) {
        if self.mode == Mode::Normal {
            self.cursor.move_prev_line(&self.buffer.slice(..));
            self.buffer.insert_char(self.cursor.index, '\n');
            let slice = self.buffer.slice(..);
            self.cursor.index = slice.next_grapheme_boundary(self.cursor.index);
            self.mode = Mode::Insert;
        }
    }

    /// Delete the character before the cursor.
    pub fn backspace(&mut self) {
        if self.mode == Mode::Insert && self.cursor.index > 0 {
            let prev_index = self
                .buffer
                .slice(..)
                .prev_grapheme_boundary(self.cursor.index);
            self.buffer.remove(prev_index..self.cursor.index);
            self.cursor.index = prev_index;
        }
    }

    /// Move the cursor to the next line.
    pub fn move_next_line(&mut self) {
        self.cursor.move_next_line(&self.buffer);
        if self.mode == Mode::Normal {
            self.update_selection();
        }
    }

    /// Update the selection end as the cursor moves.
    pub fn update_selection(&mut self) {
        if self.mode == Mode::Normal {
            self.selection.end = self.cursor.index;
        }
    }

    /// Delete the selected text.
    pub fn delete_selection(&mut self) {
        if self.mode == Mode::Normal && self.selection.is_active() {
            let start = self.selection.start.min(self.selection.end);
            let end = self.selection.end.max(self.selection.end);
            self.buffer.remove(start..end);
            self.cursor.index = start;
            self.selection.clear();
        }
    }

    pub fn insert_char(&mut self, c: char) {
        if self.mode == Mode::Insert {
            self.buffer.insert_char(self.cursor.index, c);
            self.cursor.move_next_grapheme(&self.buffer.slice(..));
        }
    }

    pub fn insert_newline(&mut self) {
        if self.mode == Mode::Insert {
            self.buffer.insert_char(self.cursor.index, '\n');
            let slice = self.buffer.slice(..);
            self.cursor.index = slice.next_grapheme_boundary(self.cursor.index);
        }
    }

    pub fn move_cursor(&mut self, direction: Direction, granularity: Granularity) {
        match (direction, granularity) {
            (Direction::Backward, Granularity::Character) => {
                self.cursor.move_prev_grapheme(&self.buffer.slice(..))
            }
            (Direction::Forward, Granularity::Character) => {
                self.cursor.move_next_grapheme(&self.buffer.slice(..))
            }
            (Direction::Backward, Granularity::Word) => {
                self.cursor.move_prev_word(&self.buffer.slice(..))
            }
            (Direction::Forward, Granularity::Word) => {
                self.cursor.move_next_word(&self.buffer.slice(..))
            }
            (Direction::Backward, Granularity::Line) => {
                self.cursor.move_prev_line(&self.buffer.slice(..))
            }
            (Direction::Forward, Granularity::Line) => self.cursor.move_next_line(&self.buffer),
        }
    }

    /// Updates the editor mode.
    pub fn update_mode(&mut self, mode: Mode) {
        if mode == Mode::Normal {
            self.cursor.move_prev_grapheme(&self.buffer.slice(..));
        }
        self.mode = mode;
    }
}

type Coords = (usize, usize); // line, col

/// Convert a character index to (line, column) coordinates.
pub fn coords_at_pos(text: &RopeSlice, pos: usize) -> Coords {
    let line = text.char_to_line(pos);
    let line_start = text.line_to_char(line);
    let col = text.slice(line_start..pos).len_chars();
    (line, col)
}

/// Convert (line, column) coordinates to a character index
fn pos_at_coords(text: &RopeSlice, coords: Coords) -> usize {
    let (line, col) = coords;
    let line_start = text.line_to_char(line);
    text.next_grapheme_boundary(line_start + col)
}

fn move_vertically(text: &RopeSlice, direction: Direction, pos: usize) -> usize {
    let (line, col) = coords_at_pos(text, pos);
    let new_line = match direction {
        Direction::Forward => std::cmp::min(line + 1, text.len_lines() - 1),
        Direction::Backward => line.saturating_sub(1),
    };

    // Convert to 0-indexed, subtract another 1 because len_chars() counts \n
    let new_line_len = text.line(new_line).len_chars().saturating_sub(2);

    let new_col = if new_line_len < col {
        // TODO: preserve horizontal
        new_line_len
    } else {
        col
    };

    pos_at_coords(text, (new_line, new_col))
}
