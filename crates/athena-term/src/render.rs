use anyhow::Result;
use crossterm::{
    event::{self, Event, KeyCode, KeyModifiers},
    terminal::{
        disable_raw_mode, enable_raw_mode, Clear, ClearType, EnterAlternateScreen,
        LeaveAlternateScreen,
    },
    QueueableCommand,
};
use std::{
    collections::HashMap,
    io::{stdout, Stdout, Write},
    sync::Arc,
};
use tokio::sync::{
    mpsc::{self, Receiver, Sender},
    RwLock,
};

use athena_core::{
    commands::EditorCommand,
    state::{coords_at_pos, EditorEvent, EditorState},
    Direction, Granularity, Mode,
};

use crate::view::{
    editor::Document as TextEditor,
    status_bar::{default_status_bar_config, StatusBar},
    View,
};

pub struct Editor {
    state: Arc<RwLock<EditorState>>,
    views: HashMap<String, Box<dyn View>>,
    event_sender: Sender<EditorEvent>,
    event_receiver: Receiver<EditorEvent>,
    command_sender: Sender<EditorCommand>,
    command_receiver: Receiver<EditorCommand>,
}

impl Editor {
    pub fn new() -> Self {
        let (event_sender, event_receiver) = mpsc::channel(100);
        let (command_sender, command_receiver) = mpsc::channel(100);
        Self {
            state: Arc::new(RwLock::new(EditorState::new())),
            views: HashMap::new(),
            event_sender,
            event_receiver,
            command_sender,
            command_receiver,
        }
    }

    pub fn with_default(mut self) -> Self {
        self.add_view(
            "status_bar",
            Box::new(StatusBar::new(default_status_bar_config()).with_default()),
        );
        self.add_view("text_editor", Box::new(TextEditor::new()));
        self
    }

    pub fn add_view(&mut self, name: &str, view: Box<dyn View>) {
        self.views.insert(name.to_string(), view);
    }

    pub async fn run(&mut self) -> Result<()> {
        enable_raw_mode()?;

        let mut stdout = stdout();
        stdout
            .queue(EnterAlternateScreen)?
            .queue(Clear(ClearType::All))?;

        self.render(&mut stdout).await?; // initial render

        loop {
            tokio::select! {
                Some(command) = self.command_receiver.recv() => {
                    if command == EditorCommand::Quit {
                        return Ok(());
                    }
                    self.handle_command(command).await?;
                    self.render(&mut stdout).await?;
                }
                Some(event) = self.event_receiver.recv() => {
                    self.handle_event(event).await?;
                    self.render(&mut stdout).await?;
                }
            }
        }
    }

    pub async fn render(&mut self, stdout: &mut Stdout) -> Result<()> {
        let state = self.state.read().await;

        for view in self.views.values_mut() {
            if view.is_dirty() {
                view.render(stdout, &state)?;
                view.mark_clean();
            }
        }

        stdout.flush()?;

        Ok(())
    }

    async fn handle_command(&mut self, command: EditorCommand) -> Result<()> {
        let mut state = self.state.write().await;
        match command {
            EditorCommand::InsertChar(ch) => {
                state.insert_char(ch);
                self.event_sender.send(EditorEvent::BufferChanged).await?;
            }
            EditorCommand::InsertNewLine => {
                state.insert_newline();
                self.event_sender.send(EditorEvent::BufferChanged).await?;
            }
            EditorCommand::DeleteChar => {
                state.backspace();
                self.event_sender.send(EditorEvent::BufferChanged).await?;
            }
            EditorCommand::MoveCursor(dir, gran) => {
                state.move_pos(dir, gran);
                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.0, coords.1))
                    .await?;
            }
            EditorCommand::UpdateMode(mode) => {
                state.update_mode(mode);
                self.event_sender
                    .send(EditorEvent::ModeChanged(mode))
                    .await?;
            }

            EditorCommand::Append => {
                state.append();
                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.0, coords.1))
                    .await?;
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;
            }
            EditorCommand::AppendBelow => {
                state.insert_newline_below();
                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.0, coords.1))
                    .await?;
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;
            }
            EditorCommand::AppendAbove => {
                state.insert_newline_above();
                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.0, coords.1))
                    .await?;
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;
            }
            EditorCommand::AppendEnd => {
                state.insert_end_of_line();
                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.0, coords.1))
                    .await?;
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;
            }
            EditorCommand::AppendStart => {
                state.insert_start_of_line();
                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.0, coords.1))
                    .await?;
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;
            }
            _ => {}
        }
        Ok(())
    }

    async fn handle_event(&mut self, event: EditorEvent) -> Result<()> {
        let state = self.state.read().await;
        for view in self.views.values_mut() {
            view.handle_event(&event, &state)?;
        }
        Ok(())
    }
}

pub async fn run_editor() -> Result<()> {
    let mut editor = Editor::new().with_default();
    let command_sender = editor.command_sender.clone();
    let state = editor.state.clone();

    tokio::spawn(async move {
        handle_user_input(command_sender, state).await;
    });

    editor.run().await?;

    // revert cursor
    stdout().write("\x1B[2 q".as_bytes())?;

    stdout()
        .queue(Clear(ClearType::All))?
        .queue(LeaveAlternateScreen)?;

    disable_raw_mode()?;

    Ok(())
}

async fn handle_user_input(sender: Sender<EditorCommand>, state: Arc<RwLock<EditorState>>) {
    loop {
        if let Event::Key(key_event) = event::read().unwrap() {
            let mode = state.read().await.mode;
            let command = match mode {
                Mode::Normal => match (key_event.modifiers, key_event.code) {
                    //// MISC
                    (KeyModifiers::NONE, KeyCode::Char('q')) => Some(EditorCommand::Quit),
                    //// INSERTIONS
                    (KeyModifiers::NONE, KeyCode::Char('i')) => {
                        Some(EditorCommand::UpdateMode(Mode::Insert))
                    }
                    (KeyModifiers::SHIFT, KeyCode::Char('I')) => Some(EditorCommand::AppendStart),
                    (KeyModifiers::NONE, KeyCode::Char('a')) => Some(EditorCommand::Append),
                    (KeyModifiers::SHIFT, KeyCode::Char('A')) => Some(EditorCommand::AppendEnd),
                    (KeyModifiers::NONE, KeyCode::Char('o')) => Some(EditorCommand::AppendBelow),
                    (KeyModifiers::SHIFT, KeyCode::Char('O')) => Some(EditorCommand::AppendAbove),
                    //// MOVEMENTS
                    (KeyModifiers::NONE, KeyCode::Char('h') | KeyCode::Left) => Some(
                        EditorCommand::MoveCursor(Direction::Backward, Granularity::Character),
                    ),
                    (KeyModifiers::NONE, KeyCode::Char('l') | KeyCode::Right) => Some(
                        EditorCommand::MoveCursor(Direction::Forward, Granularity::Character),
                    ),
                    (KeyModifiers::NONE, KeyCode::Char('k') | KeyCode::Up) => Some(
                        EditorCommand::MoveCursor(Direction::Backward, Granularity::Line),
                    ),
                    (KeyModifiers::NONE, KeyCode::Char('j') | KeyCode::Down) => Some(
                        EditorCommand::MoveCursor(Direction::Forward, Granularity::Line),
                    ),
                    _ => None,
                },
                Mode::Insert => match key_event.code {
                    KeyCode::Esc => Some(EditorCommand::UpdateMode(Mode::Normal)),
                    KeyCode::Char(c) => Some(EditorCommand::InsertChar(c)),
                    KeyCode::Left => Some(EditorCommand::MoveCursor(
                        Direction::Forward,
                        Granularity::Character,
                    )),
                    KeyCode::Right => Some(EditorCommand::MoveCursor(
                        Direction::Backward,
                        Granularity::Character,
                    )),
                    KeyCode::Up => Some(EditorCommand::MoveCursor(
                        Direction::Backward,
                        Granularity::Line,
                    )),
                    KeyCode::Down => Some(EditorCommand::MoveCursor(
                        Direction::Forward,
                        Granularity::Line,
                    )),
                    KeyCode::Backspace => Some(EditorCommand::DeleteChar),
                    KeyCode::Enter => Some(EditorCommand::InsertNewLine),
                    _ => None,
                },
            };

            if let Some(cmd) = command {
                sender.send(cmd).await.unwrap();
            }
        }
    }
}
