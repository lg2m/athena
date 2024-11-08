use serde::{Deserialize, Serialize};
use std::{collections::HashMap, fs, path::PathBuf};

/// Athena config directory ($HOME/.config/athena)
const CONFIG_DIR: &str = "~/.config/athena";
const CONFIG_FILE: &str = "athena.toml";

#[derive(Serialize, Deserialize, Debug, Default)]
pub struct Config {
    pub editor: EditorConfig,
    pub keymap: KeymapConfig,
}

#[derive(Serialize, Deserialize, Debug, Default)]
pub struct EditorConfig {
    // pub theme: String,
    pub gutters: GuttersConfig,
    pub status_bar: StatusBarConfig,
    pub cursor: CursorConfig,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct GuttersConfig {
    pub layout: Vec<GutterElement>,
    #[serde(rename = "line_numbers")]
    pub line_numbers: Option<LineNumbersConfig>,
}

impl Default for GuttersConfig {
    fn default() -> Self {
        Self {
            layout: vec![GutterElement::Spacer, GutterElement::LineNumbers],
            line_numbers: Some(LineNumbersConfig::default()),
        }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "snake_case")]
pub enum GutterElement {
    Spacer,
    LineNumbers,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct LineNumbersConfig {
    pub relative: bool,
    #[serde(rename = "min_width")]
    pub min_width: u8,
}

impl Default for LineNumbersConfig {
    fn default() -> Self {
        Self {
            relative: true,
            min_width: 3,
        }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct StatusBarConfig {
    pub left: Vec<StatusBarItem>,
    pub center: Vec<StatusBarItem>,
    pub right: Vec<StatusBarItem>,
    pub mode: ModeNames,
}

impl Default for StatusBarConfig {
    fn default() -> Self {
        Self {
            left: vec![StatusBarItem::Mode],
            center: vec![],
            right: vec![
                StatusBarItem::CursorPosition,
                StatusBarItem::LineCount,
                StatusBarItem::Language,
            ],
            mode: ModeNames::default(),
        }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "snake_case")]
pub enum StatusBarItem {
    Mode,
    CursorPosition,
    Language,
    LineCount,
    FileName,
    FileEncoding,
    FileType,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ModeNames {
    pub normal: String,
    pub insert: String,
}

impl Default for ModeNames {
    fn default() -> Self {
        Self {
            normal: "Normal".to_string(),
            insert: "Insert".to_string(),
        }
    }
}

#[derive(Serialize, Clone, Copy, Deserialize, Debug, Eq, PartialEq, Hash)]
#[serde(rename_all = "snake_case")]
pub enum Mode {
    Normal,
    Insert,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct CursorConfig {
    pub normal: CursorShape,
    pub insert: CursorShape,
}

impl Default for CursorConfig {
    fn default() -> Self {
        Self {
            normal: CursorShape::Block,
            insert: CursorShape::Bar,
        }
    }
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "snake_case")]
pub enum CursorShape {
    Block,
    Bar,
    Underline,
}

#[derive(Serialize, Deserialize, Debug, Default)]
pub struct KeymapConfig {
    pub normal: HashMap<String, StringOrNestedMap>,
    pub insert: HashMap<String, StringOrNestedMap>,
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(untagged)]
pub enum StringOrNestedMap {
    String(String),
    NestedMap(HashMap<String, String>),
}

pub fn get_config_or_default() -> Config {
    resolve_config_path()
        .and_then(|f| fs::read_to_string(f).ok())
        .and_then(|c| toml::from_str(&c).ok())
        .unwrap_or_else(Config::default)
}

fn resolve_config_path() -> Option<PathBuf> {
    dirs::home_dir().map(|home| {
        home.join(CONFIG_DIR.trim_start_matches("~/"))
            .join(CONFIG_FILE)
    })
}
