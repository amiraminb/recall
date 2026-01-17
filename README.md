# Recall

Spaced repetition for your wiki notes. Never forget what you learn.

Recall scans your markdown wiki for topics marked with `review: true` in YAML frontmatter, tracks them using the FSRS algorithm, and reminds you when it's time to review.

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

## Quick Start

1. Initialize with your wiki path:
```bash
recall init ~/wiki
```

2. Mark topics for review using YAML frontmatter:
```markdown
---
tags:
  - devops
  - k8s
review: true
---
# Kubernetes Architecture

Content about kubernetes...
```

3. Scan your wiki:
```bash
recall scan
```

4. Check what's due:
```bash
recall due
```

5. Review a topic:
```bash
recall review "Kubernetes Architecture"
```

## Commands

```bash
recall -h
```

| Command                | Description                                     |
|------------------------|-------------------------------------------------|
| recall init <path>     | Initialize with wiki path                       |
| recall scan            | Scan wiki for review topics                     |
| recall due             | Show status and topics due                      |
| recall due --week      | Show topics due this week                       |
| recall due --tag <tag> | Filter by tag                                   |
| recall read <title>    | Mark first read and schedule first review       |
| recall open <title>    | Open the first markdown link in a topic         |
| recall review <title>  | Review a topic and rate recall                  |
| recall tags            | List all tags with counts                       |
| recall history <title> | Show review history for a topic                 |
| recall remove <title>  | Remove a topic from tracking                    |

## Topic Format

Mark any markdown file for review using YAML frontmatter:

```markdown
---
id: "Topic Title"
tags:
  - tag1
  - tag2
review: true
---

Your content here...
```

- `review: true` - Marks the file as a reviewable topic
- `tags` - Categorize topics (e.g., leetcode, architecture)

The topic title is taken from the `id` field in frontmatter, or the filename if not set.

## FSRS Algorithm

Recall uses [FSRS](https://github.com/open-spaced-repetition/fsrs4anki) (Free Spaced Repetition Scheduler), the same algorithm used in Anki. When reviewing, rate your recall:

| Rating    | Meaning              | Effect           |
|-----------|----------------------|------------------|
| 1 - Again | Forgot completely    | Reset interval   |
| 2 - Hard  | Difficult to recall  | Shorter interval |
| 3 - Good  | Recalled with effort | Normal interval  |
| 4 - Easy  | Effortless recall    | Longer interval  |

## Workflow

### When you learn something new

1. **Take notes in your wiki:**
```markdown
---
tags:
  - devops
  - docker
review: true
---
# Docker Networking

- Bridge network: default, containers on same host
- Host network: shares host's network stack
- Overlay: multi-host communication
```

2. **Run scan to track it:**
```bash
recall scan
```

3. **Mark as read after studying:**
```bash
recall read "Docker Networking"
```
Rate your understanding (1-4) to schedule your first review.

### Daily routine

1. **Check what's due:**
```bash
recall due
```

2. **For each due topic:**
   - Open the file, read your notes
   - Run review:
   ```bash
   recall review "Docker Networking"
   ```
   - Rate how well you remembered (1-4)

3. **FSRS schedules next review:**
   - Good recall → longer interval (e.g., 3 days → 7 days → 14 days)
   - Poor recall → shorter interval (reset or reduced)

### Summary

| Action | When |
|--------|------|
| Add frontmatter with `review: true` | When learning something new |
| `recall scan` | After adding new topics |
| `recall due` | Daily - see what needs attention |
| `recall read "Topic"` | First time reading a topic |
| `recall open "Topic"` | Open the first link in a topic |
| `recall review "Topic"` | For topics due for review |

## Data Storage

- Config: `~/.config/recall/config.json`
- Review data: `<wiki>/.srs/reviews.json`

## License

MIT
