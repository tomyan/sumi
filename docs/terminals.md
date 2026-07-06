# Terminal support

Sumi drives the terminal with a small set of widely-implemented escape
sequences and is designed to degrade gracefully: every capability below
is either ignored harmlessly by terminals that lack it, or has a
legacy fallback that sumi keeps active at all times.

## Sequences sumi emits

| Capability | Sequence(s) | Fallback when unsupported |
|---|---|---|
| Alternate screen | `CSI ?1049h` / `CSI ?1049l` | none (fullscreen apps assume it) |
| Cursor visibility | `CSI ?25l` / `CSI ?25h` | cosmetic |
| Bracketed paste | `CSI ?2004h` / `CSI ?2004l` | paste arrives as keystrokes |
| SGR mouse + any-event | `CSI ?1006h` `CSI ?1003h` | no mouse events; keyboard still works |
| Kitty keyboard (flag 1) | `CSI >1u` push / `CSI <u` pop | ignored; legacy key encoding decoded as always |
| Synchronized output | `CSI ?2026h` / `CSI ?2026l` around frames | ignored; frames may tear |
| Clipboard write | `OSC 52 ;c;<base64> BEL` (plus pbcopy / wl-copy / xclip / clip) | platform tool alone |
| Colour-scheme probe | `OSC 11 ;? BEL` | dark scheme assumed |
| Window title | `OSC 2`, `CSI 22;2t` / `CSI 23;2t` save/restore | ignored |
| Cells | CUP + SGR (truecolor, 256, or 16 per detected depth) | depth quantized at emission |

## Sequences sumi decodes

Legacy CSI keys (arrows, Home/End, `~` extendeds, F1–F12 incl. SS3 and
modifier params), kitty `CSI <codepoint>;<mods>u` key reports, SGR
mouse (`CSI <…M/m`), bracketed paste, OSC 11 colour reports, and the
modifier bitmask+1 encoding shared by both key paths.

## Terminal matrix

"✓" = works with sumi's sequences; "–" = capability missing but sumi
degrades as described above.

| Terminal | Kitty keys | OSC 52 | Sync output | SGR mouse | OSC 11 |
|---|---|---|---|---|---|
| kitty | ✓ | ✓ | ✓ | ✓ | ✓ |
| Ghostty | ✓ | ✓ | ✓ | ✓ | ✓ |
| WezTerm | ✓ | ✓ | ✓ | ✓ | ✓ |
| iTerm2 (3.5+) | ✓ | ✓ | ✓ | ✓ | ✓ |
| Alacritty (0.13+) | ✓ | ✓ | ✓ | ✓ | ✓ |
| foot | ✓ | ✓ | ✓ | ✓ | ✓ |
| Terminal.app | – | – | – | ✓ | ✓ |
| tmux (3.3+) | pass-through¹ | needs `set-clipboard`² | ✓ | ✓ | outer terminal |

¹ tmux forwards kitty-protocol sequences when `extended-keys` is
enabled; otherwise sumi's legacy decoding applies.
² `set -s set-clipboard on` lets tmux forward OSC 52 to the outer
terminal.

## CI round-trip verification

`runtime/tui/roundtrip_test.go` runs real apps against injected
streams and replays every emitted byte through the in-repo terminal
model (`runtime/vt100`), asserting the reconstructed screen — content,
styles, and diffed updates — matches the intended frame. The model is
kept in sync with the sequences sumi emits (it consumes private-marker
CSI, OSC, and SGR forms), so any new emission that a terminal parser
would misread fails CI immediately.
