use anyhow::Result;
use crossterm::{
    style::{Attribute, Color, Print, SetAttribute, SetBackgroundColor, SetForegroundColor},
    QueueableCommand,
};
use itertools::Itertools;

use athena_core::{
    state::{coords_at_pos, EditorEvent},
    EditorState, Mode, StatusBarConfig, StatusBarItem,
};

use crate::terminal::Terminal;

use super::View;

pub struct StatusBar {
    config: StatusBarConfig,
    dirty: bool,
}

impl StatusBar {
    pub fn new(config: &StatusBarConfig) -> Self {
        Self {
            config: config.clone(),
            dirty: true,
        }
    }

    fn build_section(&self, items: &[StatusBarItem], state: &EditorState) -> String {
        items
            .iter()
            .map(|item| match item {
                StatusBarItem::Mode => format!(
                    "{}",
                    match state.mode {
                        Mode::Normal => self.config.mode.normal.to_string(),
                        Mode::Insert => self.config.mode.insert.to_string(),
                    }
                ),
                StatusBarItem::CursorPosition => {
                    let pos = state.cursor.index;
                    let offset = coords_at_pos(&state.buffer.slice(..), pos);
                    format!("{}:{}", offset.0, offset.1)
                }
                StatusBarItem::Language => "rust".to_string(),
                StatusBarItem::LineCount => state.buffer.len_lines().to_string(),
                StatusBarItem::FileName => "test.rs".to_string(),
                StatusBarItem::FileEncoding => "UTF-8".to_string(),
                StatusBarItem::FileType => "".to_string(),
            })
            .collect::<Vec<_>>()
            .join(" | ")
    }

    fn format_sections(&self, sections: Vec<String>, width: usize) -> String {
        let total_len: usize = sections.iter().map(String::len).sum();

        let (left, center, right) = if total_len > width {
            let max_widths = [width / 3, width / 3, width.saturating_sub(2 * (width / 3))];
            sections
                .into_iter()
                .zip(max_widths.iter())
                .map(|(s, &max)| s.chars().take(max).collect::<String>())
                .collect_tuple()
                .unwrap_or_default()
        } else {
            sections.into_iter().collect_tuple().unwrap_or_default()
        };

        let remaining_width = width.saturating_sub(left.len() + center.len() + right.len());
        let left_padding = remaining_width / 2;
        let right_padding = remaining_width - left_padding;

        format!(
            "{:<width_left$}{:^width_center$}{:>width_right$}",
            left,
            center,
            right,
            width_left = left.len() + left_padding,
            width_center = center.len(),
            width_right = right.len() + right_padding,
        )
    }
}

impl View for StatusBar {
    fn render(&mut self, terminal: &mut Terminal, state: &EditorState) -> Result<()> {
        if !self.is_dirty() {
            return Ok(());
        }

        self.mark_clean();

        let (width, height) = terminal.size()?;
        let sections = [&self.config.left, &self.config.center, &self.config.right]
            .iter()
            .map(|&section| self.build_section(section, state))
            .collect::<Vec<_>>();

        let content = self.format_sections(sections, width);

        terminal.goto(0, height)?;
        terminal
            .stdout
            .queue(SetAttribute(Attribute::NormalIntensity))?
            .queue(SetBackgroundColor(Color::Rgb {
                r: 59,
                g: 59,
                b: 84,
            }))?
            .queue(SetForegroundColor(Color::Rgb {
                r: 35,
                g: 240,
                b: 144,
            }))?
            .queue(SetAttribute(Attribute::Bold))?
            .queue(Print(content))?
            .queue(SetAttribute(Attribute::Reset))?
            .queue(SetBackgroundColor(Color::Rgb {
                r: 41,
                g: 41,
                b: 61,
            }))?
            .queue(SetForegroundColor(Color::Rgb {
                r: 35,
                g: 240,
                b: 144,
            }))?;

        Ok(())
    }

    fn handle_event(&mut self, event: &EditorEvent, _state: &EditorState) -> Result<()> {
        match event {
            EditorEvent::CursorMoved(_, _)
            | EditorEvent::ModeChanged(_)
            | EditorEvent::BufferChanged => self.dirty = true,
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
