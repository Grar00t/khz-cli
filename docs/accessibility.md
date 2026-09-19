# Accessibility

KHZ uses textual state markers (`OK`, `WARN`, `FAIL`, `SKIP`, `INFO`) so color is never the only signal.

Color is disabled when:

- `--no-color` is supplied;
- `NO_COLOR` is set;
- `TERM=dumb`;
- stdout/stderr is not detected as a terminal;
- local policy sets `allow_color` to false.

KHZ does not blink or animate. Human output is simple text with truncation for narrow terminals. Arabic localization affects human-visible text only; hashes, paths, branch names, JSON keys, executable names, and protocol identifiers are not translated. Perfect RTL/Bidi rendering is not claimed.
