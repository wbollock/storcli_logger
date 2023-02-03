# storcli_logger

`storcli_logger` is designed to parse the StorCLI event log into a more preferable format and regularly write these logs to disk, optionally clearing the event log on a configurable basis.

This is only needed because the storcli event log is not compatible with JSON output by default as of at least version `Ver 007.1017.0000.0000`.
