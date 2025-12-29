# Recall

Spaced repetition for your wiki notes. Never forget what you learn.

Recall scans your markdown wiki for topics marked with `@review`, tracks them using the FSRS algorithm, and reminds you when it's time to review.

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

2. Mark topics for review in your markdown files:
```markdown
## Kubernetes Architecture @review #devops #k8s

Content about kubernetes...
```

```markdown
## Binary Search @review #algorithms #leetcode

Content about binary search...
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
* For more commands look at:
```bash
recall -h
```

| Command                | Description                                     |
|------------------------|-------------------------------------------------|
| recall init <path>     | Initialize with wiki path                       |
| recall scan            | Scan wiki for @review topics                    |
| recall status          | Show overview (total, due today, due this week) |
| recall due             | List topics due for review                      |
| recall due --week      | List topics due this week                       |
| recall due --tag <tag> | Filter by tag                                   |
| recall review <title>  | Review a topic and rate recall                  |
| recall tags            | List all tags with counts                       |
| recall history <title> | Show review history for a topic                 |

## Topic Format

Mark any markdown heading with @review and optional #tags:

```markdown
## Topic Title @review #tag1 #tag2
```

- @review - Marks the heading as a reviewable topic
- #tag - Categorize topics (e.g., #leetcode, #architecture)

FSRS Algorithm

Recall uses https://github.com/open-spaced-repetition/fsrs4anki (Free Spaced Repetition Scheduler), the same algorithm used in Anki. When reviewing, rate your recall:

| Rating    | Meaning              | Effect           |
|-----------|----------------------|------------------|
| 1 - Again | Forgot completely    | Reset interval   |
| 2 - Hard  | Difficult to recall  | Shorter interval |
| 3 - Good  | Recalled with effort | Normal interval  |
| 4 - Easy  | Effortless recall    | Longer interval  |
