# Harsh Taskmaster

A minimalist, command-line tool for tracking, understanding, and forging habits.
Simple text files, informative graphs, zero friction. Habit tracking for geeks.

Track habits with just 3 commands: `ask`, `log`, `todo`.

Your data stays in human-readable text files you can edit, back up, and own
forever.

Harsh lets you see your habits in a consistency graph (aka Seinfeld chart) based
on your logging of the habit over time.

My hope is that it helps you get done what you're trying to get done in
your life and live a better one.

More info in the [launch
post](https://daryl.wakatara.com/harsh-a-minimalist-cli-habit-tracker) and the
[better habit
tracking](https://daryl.wakatara.com/better-habit-tracking-in-harsh) post.

## Quick Start

```sh
brew install harsh    # macOS/Linux
# or: yay -S harsh-bin (Arch), snap install harsh (Linux)
```

Run `harsh ask` to create your config files, then edit `~/.config/harsh/habits`:

```
# Habits file format: Name: frequency
# Frequency: 1 = daily, 7 (or 1w) = weekly, 3/7 = 3x per week

! Morning
Meditated: 1
Journaled: 1

! Weekly
Gym: 3/7
Called Mom: 1w
Cleaned house: 7

! Tracking only (no warnings)
Coffee: 0
```

## Commands

| Command           | Description                              |
| ----------------- | ---------------------------------------- |
| `harsh ask`       | Prompt for today's unrecorded habits     |
| `harsh log`       | Show consistency graph (last 100 days)   |
| `harsh todo`      | List today's pending habits with urgency |
| `harsh log stats` | Summary statistics for all habits        |

### Filtering

All commands accept optional filters:

```sh
harsh ask gym          # Ask only habits matching "gym"
harsh log sleep        # Show only sleep-related graphs
harsh ask yday         # Ask about yesterday only (also: yd, yesterday)
harsh ask 2025-01-15   # Ask about specific date
harsh ask week         # Ask about last 7 days (also: w, last-week)
```

## Recording Habits

When prompted `[y/n/s/⏎]`:

- **y** = Yes, did it
- **n** = No, missed it (breaks streak)
- **s** = Skip (life happened, doesn't break streak)
- **⏎** = Leave unanswered for now

### Adding Notes and Amounts

Track quantities and comments with `@` and `#` in `harsh ask`:

```
Pullups [y/n/s/⏎] y @ 25 # New personal record
```

The `@` amount and `#` comment are optional. Use `@` before `#` if using both.

## Reading the Graph

```
                          ▄▃▃▄▆▆▅▆▇▆▄▃▄▆▃▆▃▆▂▅▄▃▄▅▆▅▃▃▆▂▄▅▄
                Meditated ━━  ━━ ━━━━━━━━━━━━━━━━ ━━━━━━━━━━
    Cleaned the apartment ━──────━──────━──────•······━─────
           Had a headache    ━  ━     ━━      ━━   ━   ━━

Yesterday's score: 66.6%
```

### Symbol Legend

| Symbol | Meaning                                       |
| ------ | --------------------------------------------- |
| `━`    | Done                                          |
| `─`    | Satisfied (within interval, no action needed) |
| `•`    | Skipped                                       |
| `·`    | Skipified (within skip grace period)          |
| `!`    | Warning (streak at risk)                      |
| ` `    | Not recorded / not due                        |
| `▏`  | Habit tracking ended                          |

The sparkline at the top shows daily completion percentage. The score excludes
skipped habits.

## Todo with Urgency

```sh
$ harsh todo
2025-01-15 Wed:
              Write    (2 days)   # 2 days before streak breaks
               Read     (1 day)
            Spanish     (Today)   # Do today or streak breaks
                Fit     (——|  )   # Streak already broken
      Bowling Night     (——>  )   # Streak is skipping
```

## Habits File Format

Location: `~/.config/harsh/habits` (or `%APPDATA%\harsh` on Windows, override
with `HARSHPATH`)

```
# Comments start with #
! Headings start with !

Habit Name: frequency
```

**Frequency formats:**

- `1` - Daily
- `7` or `1w` - Weekly
- `3/7` - 3 times per 7 days (rolling window)
- `2/30` - Twice per month
- `0` - Track only (no warnings, doesn't affect score)

**Optional end date:**

Retire a habit by adding an end date (format: `YYYY-MM-DD`):

```
Old Habit: 1: 2024-06-15
```

After the end date:

- Habit is excluded from scoring and sparklines
- Graph shows blank after end date
- Habit no longer appears in `ask` or `todo`
- Habit name, graph, and end marker muted
- Use `harsh -H log` to hide ended habits from log output

NB: Do not use `:` in habit names (it is used as delimiter in the habit files).

## Log File Format

Location: `~/.config/harsh/log`

```
2025-01-15 : Habit Name : y : optional comment : optional amount
```

Entries are appended automatically. Edit manually if needed.

## Installation

### Package Managers (recommended)

```sh
brew install harsh              # macOS, Linux
yay -S harsh-bin                # Arch AUR
sudo snap install harsh         # Linux snap
```

### One-Line Install (macOS/Linux)

```sh
curl -sSLf https://raw.githubusercontent.com/wakatara/harsh/master/install.sh | sh
```

### Distro Packages

Download from [releases](https://github.com/wakatara/harsh/releases):

- **DEB**: `sudo apt install ./harsh_*.deb`
- **RPM**: `sudo dnf install ./harsh-*.rpm`
- **APK**: `sudo apk add --allow-untrusted harsh-*.apk`

### Go Install

```sh
go install github.com/wakatara/harsh@latest
```

### Shell Completion

```sh
harsh completion bash > /etc/bash_completion.d/harsh
harsh completion zsh > "${fpath[1]}/_harsh"
harsh completion fish > ~/.config/fish/completions/harsh.fish
```

## Options

**Global options:**

```
-C, --color string   Color output: "always", "never", "auto" (default "auto")
-H, --hide-ended     Hide habits that have an end date
-h, --help           Show help
-v, --version        Show version
```

Suppress colors for logging or piping: `harsh --color never log stats`

Hide ended habits from all output: `harsh -H log` or `harsh -H log stats`

## Tips

**Shell aliases** for faster access:

```sh
alias h="harsh"
alias ha="harsh ask"
alias hl="harsh log"
alias ht="harsh todo"
alias hc="$EDITOR ~/.config/harsh/habits"
```

**Snap users**: Set `HARSHPATH=~/.config/harsh/` and move your files there.
Snap's uninstaller deletes `~/snap/harsh/` including your data.

**Start small**: 5-8 habits mixing daily and weekly. Start slow. Too many habits
leads to being stretched thin and not achieving your goals (and sometimes
burnout).

## License

[MIT License](LICENSE)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). PRs welcome!

## Credits

Created by [Daryl Manning](https://daryl.wakatara.com).  
Inspired by [habitctl](https://github.com/blinry/habitctl) by blinry.

Contributors:

- [JamesD](https://github.com/jnd-au)
- [Aisling](https://github.com/ais)
- [vchslv13](https://github.com/vchslv13)
- [manu-cyber](https://github.com/manu-cyber)
- [nimaaskarian](https://github.com/nimaaskarian)
