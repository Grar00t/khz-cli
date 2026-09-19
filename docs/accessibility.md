# Accessibility

KHZ is a CLI. Semantic states are always the words `OK`, `WARN`, `FAIL`, `SKIP`, and `INFO`. Color is optional decoration.

## Color off

ANSI is disabled when any of these hold:

- `--no-color`
- `NO_COLOR` is set (https://no-color.org)
- `TERM=dumb`
- stdout is not a TTY
- policy `allow_color` is false (unless `--color`)

`--color` forces ANSI on a non-TTY. `--no-color` wins over `--color`.

No blinking, no animations, no mandatory emoji.

## Narrow terminals

`COLUMNS` is honored when it is an integer greater than 20. Lines are truncated with `...`.

## Language

`--lang ar` translates human labels on the board. It does not translate:

- hashes
- paths
- branch names
- JSON keys
- executable names
- protocol identifiers
- the state tokens OK/WARN/FAIL/SKIP/INFO

Terminal RTL/bidi correctness is not claimed.

## Injection

Untrusted strings are stripped of CSI/OSC/C0 controls before display so a hostile branch name cannot restyle the board.
