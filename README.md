# storcli_logger

`storcli_logger` is designed to parse the StorCLI event log into a more preferable format and regularly write these logs to disk, optionally clearing the event log on a configurable basis.

This is only needed because the storcli event log is not compatible with JSON output by default as of at least version `Ver 007.1017.0000.0000`.

## Motivation

RAID card logs are [only accessible](https://unix.stackexchange.com/questions/347940/where-lsi-megaraid-log-files-stored-on-linux) via "controller-specific programs" like `storcli`, but it can be useful for them to live on the disk of the physical host.

`storcli` also does not have any built-in features to regularly capture event logs and write them to a disk, useful for log aggregation or metrics gathering. This application is designed to provide that.

## Format

`storcli_logger` uses the `logfmt` format:

```console
# /var/log/storcli.log

time="2023-02-03T09:40:18-05:00" level=info Class=0 Code=0x00000010 Event Data="Event Data:;===========;None;" Event Description="Factory defaults restored" Locale=0x20 SeqNum=0x0000c65c Time="Thu Jan 12 17:16:15 2023"
time="2023-02-03T09:40:18-05:00" level=info Class=0 Code=0x00000223 Event Data= Event Description="Encl PD 41 Inquiry info: Info- LSI  SAS3x40           0 GB" Locale=0x02 SeqNum=0x0000c66b Time="Thu Jan 12 17:16:42 2023"
```
