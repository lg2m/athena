use anyhow::Result;
use crossterm::{
    style::{Attribute, Color, Print, SetAttribute, SetBackgroundColor, SetForegroundColor},
    QueueableCommand,
};

use athena_core::{
    state::{coords_at_pos, EditorEvent},
    EditorState, GutterElement, GuttersConfig, LineNumbersConfig, EDITOR_BG, EDITOR_FG,
    LINE_NUMBER_BG, LINE_NUMBER_FG,
};

const LN_BG: Color = Color::Rgb {
    r: LINE_NUMBER_BG.0,
    g: LINE_NUMBER_BG.1,
    b: LINE_NUMBER_BG.2,
};

const LN_FG: Color = Color::Rgb {
    r: LINE_NUMBER_FG.0,
    g: LINE_NUMBER_FG.1,
    b: LINE_NUMBER_FG.2,
};

const E_FG: Color = Color::Rgb {
    r: EDITOR_FG.0,
    g: EDITOR_FG.1,
    b: EDITOR_FG.2,
};
const E_BG: Color = Color::Rgb {
    r: EDITOR_BG.0,
    g: EDITOR_BG.1,
    b: EDITOR_BG.2,
};

use crate::terminal::Terminal;

use super::View;

pub struct Document {
    config: GuttersConfig,
    dirty: bool,
}

impl Document {
    pub fn new(config: &GuttersConfig) -> Self {
        Self {
            config: config.clone(),
            dirty: true,
        }
    }

    fn dent(&self, state: &EditorState) -> usize {
        let mut width = 0;
        for element in &self.config.layout {
            match element {
                GutterElement::Spacer => {
                    width += 1;
                }
                GutterElement::LineNumbers => {
                    width += state.buffer.len_lines().to_string().len() + 1
                }
            }
        }
        width
    }

    fn width(&self, txt: &str, width: usize) -> usize {
        let tabs = txt.matches('\t').count();
        (txt.len() + tabs * width).saturating_sub(tabs)
    }

    fn get_line_number_display(
        &self,
        state: &EditorState,
        y: u16,
        config: &LineNumbersConfig,
    ) -> String {
        if y as usize >= state.buffer.len_lines() {
            "~".to_string()
        } else if config.relative {
            let pos = state.cursor.index;
            let offset = coords_at_pos(&state.buffer.slice(..), pos);
            if y as usize == offset.0 {
                (y + 1).to_string() // Show current line number
            } else {
                (y as i32 - offset.0 as i32).abs().to_string()
            }
        } else {
            (y + 1).to_string()
        }
    }

    fn render_gutter(
        &mut self,
        terminal: &mut Terminal,
        state: &EditorState,
        y: u16,
    ) -> Result<()> {
        terminal
            .stdout
            .queue(SetAttribute(Attribute::NormalIntensity))?
            .queue(SetBackgroundColor(LN_BG))?
            .queue(SetForegroundColor(LN_FG))?;

        for element in &self.config.layout {
            match element {
                GutterElement::Spacer => {
                    terminal.stdout.queue(Print(" "))?;
                }
                GutterElement::LineNumbers => {
                    if let Some(line_numbers_config) = &self.config.line_numbers {
                        let line_num = self.get_line_number_display(state, y, line_numbers_config);
                        let min_width = line_numbers_config.min_width.min(4) as usize;

                        terminal.stdout.queue(Print(format!(
                            "{:width$}",
                            line_num,
                            width = min_width
                        )))?;
                    }
                }
            }
        }

        // Reset colors to prepare for rendering the line content
        terminal
            .stdout
            .queue(SetForegroundColor(E_FG))?
            .queue(SetBackgroundColor(E_BG))?;

        Ok(())
    }

    fn render_line_content(
        &self,
        terminal: &mut Terminal,
        state: &EditorState,
        y: u16,
        width: u16,
    ) -> Result<()> {
        let spacer_count = self
            .config
            .layout
            .iter()
            .filter(|e| matches!(e, GutterElement::Spacer))
            .count();
        let pad_amount;

        if let Some(line) = state.buffer.get_line(y.into()) {
            for c in line.chars() {
                terminal.stdout.queue(Print(c))?;
            }

            let tab_width = 4; // TODO: Define in config
            let line_width = self.width(line.to_string().as_str(), tab_width);
            pad_amount = width
                .saturating_sub(spacer_count as u16)
                .saturating_sub(line_width as u16);
        } else {
            pad_amount = width.saturating_sub(spacer_count as u16);
        }

        // Fill the remaining space with background color
        terminal
            .stdout
            .queue(SetForegroundColor(E_FG))?
            .queue(SetBackgroundColor(E_BG))?
            .queue(Print(" ".repeat(pad_amount as usize)))?;

        Ok(())
    }
}

impl View for Document {
    fn render(&mut self, terminal: &mut Terminal, state: &EditorState) -> Result<()> {
        let (width, height) = terminal.size()?;
        let w = width.saturating_sub(self.dent(state));

        for y in 0..u16::try_from(height).unwrap_or(0) {
            terminal.goto(0, y as usize)?;

            self.render_gutter(terminal, state, y)?;

            self.render_line_content(terminal, state, y, w as u16)?;

            terminal.stdout.queue(SetAttribute(Attribute::Reset))?;
        }

        Ok(())
    }

    fn handle_event(&mut self, event: &EditorEvent, _state: &EditorState) -> Result<()> {
        match event {
            EditorEvent::CursorMoved(_, _)
            | EditorEvent::BufferChanged
            | EditorEvent::ModeChanged(_) => {
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
