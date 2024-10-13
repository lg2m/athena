use ropey::{Rope, RopeSlice};

use crate::graphemes::GraphemeOperations;

#[derive(Debug, Eq, PartialEq, Hash)]
pub enum AppEvent {
    CursorMoved(usize, usize),
    ModeChanged(Mode),
    BufferChanged,
    ViewportChanged,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq, Hash)]
pub enum Mode {
    Normal,
    Insert,
}

#[derive(Copy, Clone, Debug, PartialEq, Eq)]
pub enum Direction {
    Forward,
    Backward,
}

#[derive(Copy, Clone, Debug, PartialEq, Eq)]
pub enum Granularity {
    Character,
    Word,
    Line,
}

#[derive(Clone, Debug, Default)]
pub struct CursorPosition {
    pub index: usize,
}

impl CursorPosition {
    pub fn new(index: usize) -> Self {
        Self { index }
    }
}

// TODO: rename to editor or app state
#[derive(Clone, Debug)]
pub struct State {
    pub buffer: Rope,
    pub cursor: CursorPosition,
    pub mode: Mode,
}

impl State {
    pub fn new() -> Self {
        Self {
            buffer: Rope::new(),
            cursor: CursorPosition::new(0),
            mode: Mode::Normal,
        }
    }

    pub fn insert_char(&mut self, c: char) {
        if self.mode == Mode::Insert {
            self.buffer.insert_char(self.cursor.index, c);
            let slice = self.buffer.slice(..);
            self.cursor.index = slice.next_grapheme_boundary(self.cursor.index);
        }
    }

    pub fn insert_newline(&mut self) {
        if self.mode == Mode::Insert {
            self.buffer.insert_char(self.cursor.index, '\n');
            let slice = self.buffer.slice(..);
            self.cursor.index = slice.next_grapheme_boundary(self.cursor.index);
        }
    }

    pub fn backspace(&mut self) {
        if self.mode == Mode::Insert && self.cursor.index > 0 {
            let slice = self.buffer.slice(..);
            let prev_boundary = slice.prev_grapheme_boundary(self.cursor.index);
            self.buffer.remove(prev_boundary..self.cursor.index);
            self.cursor.index = prev_boundary;
        }
    }

    pub fn move_pos(&mut self, direction: Direction, granularity: Granularity) {
        let slice = &self.buffer.slice(..);
        self.cursor.index = match (direction, granularity) {
            (Direction::Backward, Granularity::Character) => {
                slice.prev_grapheme_boundary(self.cursor.index)
            }
            (Direction::Forward, Granularity::Character) => {
                slice.next_grapheme_boundary(self.cursor.index)
            }
            (_, Granularity::Line) => move_vertically(&slice, direction, self.cursor.index),
            _ => self.cursor.index,
        };
    }

    pub fn update_mode(&mut self, mode: Mode) {
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
