use anyhow::Result;
use crossterm::{
    cursor,
    style::{Color, Print, SetForegroundColor},
    terminal::{self, Clear, ClearType},
    QueueableCommand,
};
use std::io::{Stdout, Write};

use athena_core::{
    state::{coords_at_pos, AppEvent},
    Mode, State,
};

use super::View;

pub struct Editor {
    dirty: bool,
    rendered_lines: usize,
    previous_cursor_pos: (usize, usize),
    current_mode: Mode,
}

impl Editor {
    pub fn new() -> Self {
        Self {
            dirty: true,
            rendered_lines: 0,
            previous_cursor_pos: (0, 0),
            current_mode: Mode::Normal,
        }
    }

    fn render_lines(&mut self, stdout: &mut Stdout, state: &State, size: (u16, u16)) -> Result<()> {
        let visible_lines = state.buffer.lines().take(size.1 as usize);
        let new_line_count = visible_lines.clone().count();

        let mut previous_lengths = Vec::with_capacity(new_line_count);

        if self.is_dirty() {
            stdout.queue(cursor::Hide)?;
        }

        for (i, line) in visible_lines.enumerate() {
            let line_str = line.as_str().unwrap_or_default();
            let current_length = line_str.len();
            let clear_after = if i < previous_lengths.len() {
                previous_lengths[i]
            } else {
                0
            };

            stdout
                .queue(cursor::MoveTo(0, i as u16))?
                .queue(Clear(ClearType::CurrentLine))?
                .queue(SetForegroundColor(Color::Red))?
                .queue(Print(format!("{:4}", i + 1)))?
                .queue(SetForegroundColor(Color::Reset))?
                .queue(cursor::MoveTo(5, i as u16))?
                .queue(Print(line_str))?;

            // Clear the part of the line that exceeds the current line length
            if clear_after > current_length {
                stdout.queue(Clear(ClearType::UntilNewLine))?;
            }

            // Store the current line length for comparison
            if i < previous_lengths.len() {
                previous_lengths[i] = current_length;
            } else {
                previous_lengths.push(current_length);
            }
        }

        // Remove old lines if the buffer shrunk
        if new_line_count < self.rendered_lines {
            for i in new_line_count..self.rendered_lines {
                stdout
                    .queue(cursor::MoveTo(0, i as u16))?
                    .queue(Clear(ClearType::CurrentLine))?;
            }
        }

        self.rendered_lines = new_line_count;
        Ok(())
    }

    fn render_cursor(&mut self, stdout: &mut Stdout, state: &State) -> Result<()> {
        // Only update the cursor shape and position if necessary
        if state.mode != self.current_mode {
            let cursor_shape = match state.mode {
                Mode::Insert => "\x1B[6 q", // Block cursor for insert mode
                Mode::Normal => "\x1B[2 q", // Line cursor for normal mode
            };
            stdout.write(cursor_shape.as_bytes())?;
            self.current_mode = state.mode;
        }

        // Update cursor position only if it has changed
        let pos = state.cursor.index;
        let coords = coords_at_pos(&state.buffer.slice(..), pos);
        self.previous_cursor_pos = coords;
        if coords != self.previous_cursor_pos || self.is_dirty() {
            stdout
                .queue(cursor::MoveTo((coords.1 + 5) as u16, coords.0 as u16))?
                .queue(cursor::Show)?;
        }

        Ok(())
    }
}

impl View for Editor {
    fn render(&mut self, stdout: &mut Stdout, state: &State) -> Result<()> {
        let size = terminal::size()?;

        self.render_lines(stdout, state, size)?;
        self.render_cursor(stdout, state)?;

        stdout.flush()?;

        Ok(())
    }

    fn handle_event(&mut self, event: &AppEvent, _state: &State) -> Result<()> {
        match event {
            AppEvent::CursorMoved(_, _) | AppEvent::BufferChanged | AppEvent::ModeChanged(_) => {
                self.dirty = true;
            }
            _ => {}
        }
        Ok(())
    }

    fn is_dirty(&self) -> bool {
        self.dirty
    }

    fn mark_clean(&mut self) {
        self.dirty = false;
    }
}
