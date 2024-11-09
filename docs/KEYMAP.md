# Keymap

This is basically a wishlist right now and is currently inspired/borrowed from kakoune, vim, and helix.

## Normal mode

Normal mode is the default mode when you launch the editor. You can return to it from insert mode by pressing the `Escape` key.

### Movement and Selections

| Key/Shortcut     | Description                                                                 |
|------------------|-----------------------------------------------------------------------------|
| `h`              | Select the character on the left of selection end                          |
| `j`              | Select the character below the selection end                               |
| `k`              | Select the character above the selection end                               |
| `l`              | Select the character on the right of selection end                         |
| `w`              | Select the word and following whitespaces on the right of selection end    |
| `b`              | Select preceding whitespaces and the word on the left of selection end     |
| `e`              | Select preceding whitespaces and the word on the right of selection end    |
| `[WBE]`          | Same as `[wbe]`, but selects `WORD` instead of `word`                      |
| `f`              | Select to (including) the next occurrence of the given character           |
| `t`              | Select until (excluding) the next occurrence of the given character        |
| `[FT]`           | Same as `[ft]` but in the other direction                                  |
| `m`              | Select to matching character                                               |
| `M`              | Extend selection to matching character                                     |
| `x`              | Select current line; if already selected, extend to next line              |
| `X`              | Extend selection to line bounds (line-wise selection)                      |
| `<a-x>`          | Trim selection to only line bounds (line-wise selection)                   |
| `%`              | Select the whole buffer                                                    |
| `pageup, <c-b>`  | Scroll one page up                                                         |
| `pagedown, <c-f>`| Scroll one page down                                                       |
| `<c-u>`          | Scroll half a page up                                                      |
| `<c-d>`          | Scroll half a page down                                                    |
