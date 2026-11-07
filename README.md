# Recall

Remember what you learn. Add something you've studied, review it from memory,
and let spaced repetition handle the timing.

## Installation

```bash
go install github.com/amiraminb/recall/cmd/recall@latest
```

Or build from source:

```bash
git clone https://github.com/amiraminb/recall.git
cd recall
go build ./cmd/recall/
```

## Quick start

After reading, watching, or practicing something:

```bash
recall add "TCP handshake"
```

That's it. Your first review is tomorrow. No setup is required.

Each day:

```bash
recall          # See what's due today
recall review   # Work through the queue, most overdue first
```

For each topic, explain what you remember **before checking your source**.
Then rate your recall. Recall saves each rating and schedules the next review.
Use `s` to skip a topic or `q` to finish early; you can resume with `recall review`.

## Recommended flow

1. **Study anywhere.** Read a chapter, finish a lesson, or practice a skill.
2. **Add a focused topic.** Prefer “TCP three-way handshake” over “Networking.”
3. **Optionally add a source and tags.** Keep a reference handy and organize topics.
4. **Review when due.** Explain, solve, or sketch from memory, then check your work.
5. **Rate your actual recall.** FSRS adapts the next review date to your answer.

Adding a topic means you've studied it. It schedules an initial recall attempt
for tomorrow without assigning a rating. The first actual review initializes
FSRS; later reviews adapt the interval.

### Sources and tags

Only the title is required. Optional details make reviews easier:

```bash
recall add "TCP handshake" \
  --source "https://www.rfc-editor.org/rfc/rfc9293.html" \
  --tag networking,protocols

recall add "Database indexes" --source "Designing Data-Intensive Applications, chapter 3"
```

A source can be a web URL or a reference such as a book chapter. Recall displays
it during reviews. `recall open "TCP handshake"` opens an attached web URL.

Run `recall add` without a title for a simple title prompt. Titles are matched
case-insensitively; adding a duplicate reports the existing topic.

## Commands

| Command | Description |
|---|---|
| `recall` | Today's queue and the next action |
| `recall add "Topic"` | Add a studied topic; first review tomorrow |
| `recall review` | Review today's queue interactively |
| `recall review "Topic"` | Review a specific topic, even before it's due |
| `recall review --tag go` | Review due topics with a tag |
| `recall review "Topic" --rating 3` | Record a rating without prompting |
| `recall due --week` | Show reviews due over the next seven days |
| `recall due --tag go` | Show due topics with a tag |
| `recall list` | List all topics and their next review dates |
| `recall list --tag go` | List topics with a tag |
| `recall edit "Topic" --source "Book chapter 3"` | Update a topic's details |
| `recall open "Topic"` | Open the attached source URL |
| `recall history "Topic"` | Show topic details and review history |
| `recall history` | List all topics, like `recall list` |
| `recall tags` | List tags and topic counts |
| `recall remove "Topic"` | Remove a topic from the schedule |

`recall edit` accepts `--title`, `--source`, and `--tag`.
Editing preserves the topic's schedule and history. `--tag` replaces its tags;
an empty value clears the source or tags.

```bash
recall edit "TCP handshake" --tag networking --source "RFC 9293"
recall edit "TCP handshake" --title "TCP three-way handshake"
```

Use `recall --help` or `recall <command> --help` for details.

## Ratings and scheduling

Recall uses [FSRS](https://github.com/open-spaced-repetition/fsrs4anki)
(Free Spaced Repetition Scheduler).

| Rating | Meaning |
|---|---|
| **1 — Again** | Couldn't recall it; needed to relearn |
| **2 — Hard** | Recalled it, but with difficulty |
| **3 — Good** | Recalled it with some effort |
| **4 — Easy** | Recalled it effortlessly |

Rate what you remembered before checking the source, rather than how familiar
the material feels after rereading it. Skipping leaves the schedule unchanged.

The target retention is `0.9`, with a maximum interval of `365` days.
Late reviews use the elapsed time since your last review when scheduling.
The daily queue includes everything due by the end of your local calendar day.

## Shell completion

Topic completion uses the topics you've added:

```bash
recall completion zsh > "${fpath[1]}/_recall"
```

Restart your shell or run `autoload -Uz compinit && compinit` afterward.

## Data storage and upgrading

Topics, sources, tags, and review history are stored in:

```text
~/.local/share/recall/reviews.json
```

If `XDG_DATA_HOME` is set to an absolute path, Recall uses
`$XDG_DATA_HOME/recall/reviews.json` instead. Back up this file to preserve your
topics and schedule.

When upgrading from the wiki-based version, Recall reads the old configuration
at `~/.config/recall/config.json` and copies the existing `.srs/reviews.json` into
standalone storage on first use. Topic IDs, tags, schedules, and review history
are preserved. Existing standalone data takes precedence, and the original data
is retained. Note contents and links are not imported; attach sources with
`recall edit "Topic" --source "..."`.

The old `init`, `scan`, `read`, and `note` commands have been replaced by the
topic-first flow. Use `recall add` to track new learning and `recall review` to
review any topic, including migrated topics awaiting their first review.

## License

MIT
