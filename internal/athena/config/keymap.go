package config

// KeyAction represents either a direct action string or a nested map of actions
type KeyAction interface{}

type KeyMap map[string]KeyAction

// KeymapConfig represents key mappings
type KeymapConfig struct {
	Normal KeyMap `toml:"normal"`
	Insert KeyMap `toml:"insert"`
}

func defaultKeymap() KeymapConfig {
	return KeymapConfig{
		Normal: map[string]KeyAction{
			"i": "enter_insert_mode",
			"j": "move_down",
			"k": "move_up",
			"h": "move_left",
			"l": "move_right",
			"w": "move_next_word",
			"b": "move_prev_word",
			"g": map[string]string{
				"g": "go_to_top",
				"e": "go_to_bottom",
				"h": "go_to_line_start",
				"l": "go_to_line_end",
			},
			"<left>":  "move_left",
			"<right>": "move_right",
			"<up>":    "move_up",
			"<down>":  "move_down",
		},
		Insert: map[string]KeyAction{
			"<esc>": "enter_normal_mode",
			"<cr>":  "new_line",
			"<bs>":  "delete_backwards",
			"<del>": "delete_forward",
		},
	}
}
