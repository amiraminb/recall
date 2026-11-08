# Recall

A CLI for remembering what you learn with [FSRS spaced repetition](https://github.com/open-spaced-repetition/fsrs4anki).

## Install

```bash
go install github.com/amiraminb/recall/cmd/recall@latest
```

## Usage

```bash
recall add "TCP handshake"   # Add a studied topic; first review tomorrow
recall                       # See what's due today
recall review                # Review the queue, most overdue first
```

Recall from memory **before checking your source**, then rate your answer:
**1** Again · **2** Hard · **3** Good · **4** Easy.
Use `s` to skip or `q` to quit; each rating is saved immediately.

Sources and tags are optional:

```bash
recall add "Database indexes" --source "DDIA, chapter 3" --tag databases
recall review --tag databases
```

## Other commands

```bash
recall list                             # All topics and review dates
recall due --week                        # Upcoming reviews
recall review "Topic"                    # Review a specific topic
recall edit "Topic" --source "Book ch. 3" # Update title, source, or tags
recall open "Topic"                      # Open an attached source URL
recall history "Topic"                   # Review history
recall tags                             # Tags and topic counts
recall remove "Topic"                    # Remove a topic
```

Run `recall <command> --help` for options.

## Storage

Data lives in `~/.config/recall/reviews.json`. An absolute `XDG_CONFIG_HOME`
overrides `~/.config`. Back up this file to preserve your topics and history.

## License

MIT
