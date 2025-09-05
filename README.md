# Harsh Taskmaster

harsh is habit tracking for geeks. A minimalist, command line tool for tracking,
understanding, and forging habits.

Succinctly: it's quick, simple, and gets out of your way. And gives you excellent
visibility on your habits.

There are only 3 commands: `log`, `ask`, and `todo`.

Designed for simplicity, visibility, and longevity, harsh uses simple text files
for tracking that are human-grokable and editable in your favourite text editor.
It's simpler, less messy, and more portable than commercial or mobile
applications and less fussy to manage than emacs habit tracking (imho). While
quantified individual tracking is exhaustive (and exhausting), important habits
get lost in the data deluge, so this provides deliberate and explicit habits to
track and progress.

It's written in Go and adds features and fixes on top of habitctl (its
inspiration) such as skips, streak break warnings, stats, quantities, and targets
over intervals (eg. 3 times in a week).

If you're a geek, I think you'll like it. Despite trying an exhaustive number of
habit trackers, this was what worked for me. YMMV. If you're interested in why I
wrote it, there's a [launch post about my motivations on my
blog](https://daryl.wakatara.com/harsh-a-minimalist-cli-habit-tracker).

My hope is that it helps you get done what you're trying to get done in
your life and live a better one.

As a major update, in `0.11.0` we ported to [cobra](https://cobra.dev),
refactored heavily for efficiency (~60% performance gains) and better structure,
and wrote exhaustive tests.

## How It Works

Harsh lets you see your habits in a consistency graph (aka Seinfeld Chain) from
left to right over the time you've been doing the habit. The interface shows the
last 100 days (but dynamically adapts to a smaller period if your terminal
window won't support 100 days.).

For each day, you enter a yes/no/skip on whether you did the habit in question,
and the horizontal line for each habit is like an x-axis time graph showing you
how your habits have fared. There's a sparkline graph at the top of the
interface to show how you are doing percentage wise each day over the 100 days,
and some interface hints with a M(onday), W(ednesday), F(riday) label below
that to help you figure out which days goes with what (very helpful if you find
particular days are where your habits fall down consistently since you can then
fix whatever that is in that day that messes things up.).

Consistency graphs are also known as the "Seinfield Method" due to an apocryphal
story about a habit method about getting better at writing jokes attributed to
Jerry Seinfeld. Basically, the idea was he had a big year long calendar with a
day for every year. Every day he wrote jokes, he put an X down. He could see how
consistent he was being. The idea is never to bfreak the chain.

Harsh works the same way.

The idea is that turning something into a consistent habit makes you good —
eventually. The original place I saw it (Lifehacker) has killed the original
post (largely I think because Seinfeld repudiated it) - but [you can see
here](https://lifehacker.com/jerry-seinfelds-productivity-secret-281626)

## Commands

There are 3 commands: `ask`, `log`, and `todo`.
(and one subcommand, `log stats`).

`harsh ask` asks you about all your unanaswered habits.  
`harsh log` shows your full consistency graph across all habits.  
`harsh todo` shows your unanswered habits across all days.

You can also use `harsh ask <habit substring>` and `harsh log <habit
substring>` to narrow down to answering just a single habit or to see the
consistency graph for a single habit ( _eg. For the habit "Ship Every Day"
typing `harsh ask ship` narrows your outcome responses to the unanswered days
for that habit. Likewise, `harsh log ship` would show you only the consistency
graph log for that habit._)

`harsh ask` also allows you to use an ISO date as a substring (ie.
`2025-02-20`) to answer just _that_ day's unanswered habit outcomes and the
shortcut `harsh ask yday` or `harsh ask yd` to answer just yesterday's prompts.

There is one subcommand:

`harsh log stats` provides summary statistics for your entire log file of habits
done, skipped, and chains broken.

## Installation

harsh is available on OSX (also, via homebrew), Linux (as homebrew and snap),
FreeBSD, OpenBSD, and Windows. A specific goal was increasing uptake and
adoption of a portable, command line, text-based approach to habit tracking (and
that didn't require a subscription, service, or hoops to jump through to use
your own data).

harsh also supports ARM architectures for OSX (M1, M2, and M3 chipped Macs and
Linux and BSDs) as of 0.8.8. Binaries for FreeBSD and OpenBSD are also available
as of 0.8.23.

### Install via package manager

**Homebrew tap** (Mac and Linux):

```sh
brew install wakatara/tap/harsh
```

(this will also nicely alert you of updates when you `brew update`)

**Snap** (Linux only):

```sh
sudo snap install harsh
```

(you'll also get alerted when there are updates)

### Easy one-line install

If you're not using a package manager, by far the easiest way to install harsh
and get started is to use this one-line bash command if you're on OSX or Linux:

```bash
curl -sSLf https://raw.githubusercontent.com/wakatara/harsh/master/install.sh | sh
```

The script downloads the latest release directly from github, verifies it
cryptographically, and then moves the executable to `usr/local/bin` ready for
immediate use on the command line.

### With working Go environment

You may prefer a direct Go install if you have a working Go environment and Go 1.14+.

```bash
go install github.com/wakatara/harsh

```

from the command line. Unlike a package manager like brew or snap, you won't
get informed of new version releases.

### Manually

Alternatively, you can download the pre-compiled binaries from
[the releases page](https://github.com/wakatara/harsh/releases)
and copy to the desired location (`/usr/local/bin` recommended),
making sure it's in your `$PATH`.

### Compiling from source

If you want to build from source cause you like, harsh is open source and
can be git cloned from:

```sh
git clone https://github.com/wakatara/harsh
```

You can build and then move the file to the desired location in your `$PATH`.

## Getting Started

The first time you run `harsh ask`, it will set up your required files:

```sh
    $ harsh ask
    Welcome to harsh!

    Created /Users/daryl/.config/harsh/habits  This file lists your habits.
    Created /Users/daryl/.config/harsh/log     This file is your habit log.

    What? No habits yet?
    Open the habits file and edit the habit list using a text editor.
    Then run `harsh ask` to start tracking
    'harsh todo` will show you undone habits for today.
    `harsh log` will show you a consistency graph of your efforts
    (trust me, it looks way cooler looking over time)
    Happy tracking! I genuinely hope this helps you get better.
```

On OSX and Linux based systems, the `habits` and `log` files will be under
`~/.config/harsh/`. On Windows, you can find the files under `%APPDATA%\harsh`.

Alternatively, you can set a different directory using the `HARSHPATH` environment variable.

| :warning: **WARNING**                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Ubuntu's snap overrides (sensible) defaults and forcibly places harsh's config _and_ log files in `~/snap/harsh/current/`. If you `snap remove` the harsh app, snap's uninstaller will **nuke** your config _and_ log files with the sandbox and you may lose your config and log data if it's not backed up (please _always_ exercise a good backup regime). For this reason, we _highly_ recommend snap users set the `HARSHPATH` env variable to `~/.config/harsh/` and move config and log files there right after installing to protect them. Or _never_ uninstall harsh. :grin: |

Open the `habits` file in your text editor of choice (nano, vim, VS Code, Sublime, or emacs).

You'll see an example file like this:

```sh
  harsh
    # This is your habits file.
    # It tells harsh what to track and how frequently in days.
    # 1 means daily, 7 means weekly, 14 every two weeks.
    # You can also track targets within a set number of days
    # For example, Gym 3 times a week would translate to 3/7
    # 0 is for tracking a habit. 0 frequency habits will not warn or score.
    # Examples:

    ! Dailies
    Gymmed: 1
    Bed by midnight: 1

    ! Weeklies
    Gymmed: 3/7
    Called Mom: 7

    ! Monthly+
    Did Finances: 30
    New Skill: 90

    ! Tracking
    Too much coffee: 0
    Used harsh: 0
```

A big change starting in version `0.10.0`, harsh now allows you to pick a target
number of times you want to perform a habit over a set interval of time (in
days). Examples would be `3/7` for 3 times a week, `2/30` for twice a month etc
etc.

So, for example, wanting to make it to the gym 3 times a week (ie. 7 days) would
translate into a line in the habit file like:

```sh
    Gymmed: 3/7
```

You can also simply pick a number of days you want to repeat a habit (also, for
backwards compatibility with older habit files).

_Note that this uses a rolling window of days than actual week delimiters (mostly
because having hard start of week, month, and quarter breaks broke the point of
consistency graphs when I implemented it. This is much nicer. Trust me.)._

If it's not obvious from the example file, habits can have any character that is
not a `:` as that delimits the period. We also use `:` as the separator in log
files as well for easy parsing.

Headings are denoted by a "!" at the start of a line and allow you to categorize
habits visually (useful if you have a lot of them).

Comments can go in the habits file by starting a line with "#" and are not
parsed by the program.

The real trick of tracking is figuring out what habits you want to track
building or breaking. Too many, you'll fail. Too few, and the app loses its
edge. Too short-term, you feel good but fail on longer-term objectives.

If you're getting started, try 5-8 and mix short term and long term and see how
you go. Tracking your habits is _strangely_ also a habit you need to build.
There're no right answers, but if this is new, [focus on foundational keystone
habits](https://daryl.wakatara.com/resolution-keystone-habits-and-foundational-hacks/)
that will feed future ones. If you're coming into this cold, I'd also recommend
a good read of James Clear's Atomic Habits.

Here are some ideas of what to track:

- Went to bed on time
- Got X hours of sleep
- Inbox zero
- Practiced Italian
- Morning Pages
- Blogged
- Studied astrophysics
- Socialized
- Stuck to budget
- Coded
- 2 coffees only
- Thanked someone
- Went for a walk
- Told SO they're amazing

## Usage and Commands

Simply run `harsh ask` regularly, specify whether you did the habit from the
prompt on a particular day (or needed to skip the habit for some reason - eg.
could not clean apartment because you were away for week), and get pretty
graphs!

As of `v0.10.6` you can also add a fragment of a habit after `harsh ask` which
will then only ask you about that the outcomes for that single habit.
Taking the example of the `Pullups` habit below, if you had `Pullups: 5/7`
in your habit file, typing `harsh ask pul` (or any variations of the habit
name's string) will ask you only about the outcome for the unanswered todos
for that habit (an oft-requested feature).

`harsh ask` allows you to pick between `[y/n/s/⏎]` which is yes/no/skip/don't
answer right now. CTRL-c breaks you out of the ask cycle at any point and
returns you to your prompt.

As of version `0.9.0`, to support longer term pattern tracking (and fallible
memories), you can follow any of the `y | n | s` options with an optional typed
`@` symbol to then denote a quantity you want to track for the daily habit
(example: number of words written, km's run, pullups performed etc), and/or an
optional typed `#` symbol for a comment. Primarily for analysis at a later date.

As of `v0.10.12` we added date options to `harsh ask`. You can use an ISO Date (ie. `2025-02-14`) to answer _just_ that day's habits if you prefer narrowing your responses to specific days. The convenience functions `yday` and `yd` also work for yesterday. So, as examples, you can use `harsh ask 2025-02-14` and see the open habits for that day and `harsh ask yday` or `harsh ask yd` to answer just yesterday's habits.

The `log stats` subcommand will also now total up any amounts you've entered for
a habit and show you the total along with your streaks, skips, breaks, and days
tracked.

An example of how to use this for a habit like PullUps with `harsh ask pull` might be:

This will, besides adding y to the log file, also record "15" and the comment
"Crushed workout today" to your log file. The feature is backwards compatible
even with older log files so even if you don't use it, you're alright. 0.9.0
will also parse older logs. You must use a valid number in the `@` position and
the only other caveat is if you use both an `@` and `#` in your prompt response,
the `@` must come before the `#`. The only disallowed character in a prompt is
`:` as it is used as a field separator in the log files (to make for easy
importing and parsing by spreadsheet and other tools like pandas.)

The annotating features are primarily for analysis at a later date to help
uncover patterns between events, reasons you may have written a comment, and
good or bad knock-on effects.

Personally, I use comments sparingly and to denote why you had to skip, broke a
consistency chain with an `n`, or for when you're trying to figure out
something notable on a day so when you look back you can see why habit _X_ may
or may not have succeeded.

```sh
    $ harsh ask pull
    2024-01-05:
    Dailies
                PullUps ━  ━ ━  ━━  ━ ━   ━ ━ ━━━━━━━━━━━  ━ ━ ━[y/n/s/⏎] y @ 15 # Crushed workout today!
```

This would also be what `harsh ask 2024-01-05` would look like (ok, slightly cheating on examples, but...)

````sh
    $ harsh ask 2024-01-05
    2024-01-05:
    Dailies
                PullUps ━  ━ ━  ━━  ━ ━   ━ ━ ━━━━━━━━━━━  ━ ━ ━[y/n/s/⏎] y @ 15 # Crushed workout today!



Also, `harsh log` supports the same fragment querying `harsh ask`
does. You can add a fragment of a habit after `harsh log` which will then only
show you the consistency graph outcomes for that single habit. This supercedes
the `harsh log check <habit substring>` subcommand (which is not deprecated).

Taking the example of the `Pullups` habit mentioned above, if you typed `harsh
log pul` (or any variations of the habit name's string) you'll see just the
consistency graph for that single habit along along with the sparkline for all
habits and usual log command metrics.


`harsh log` by itself shows your consistency graph for the last 100 days.


```sh
    $ harsh ask
    2020-01-05:
    Dailies
                    Meditated ━       ━ ━  ━━         ━    ━   ━ ━   ━━━━━━━━━━━   ━ ━   ━[y/n/s/⏎] y

    Weeklies
        Cleaned the apartment ━──────                 ━──────           ━──────    •······[y/n/s/⏎] n

    Tracking
               Had a headache             ━  ━     ━━                  ━━   ━   ━━        [y/n/s/⏎] n
                   Used harsh ━ ━━━ ━  ━━━   ━ ━ ━       ━ ━ ━  ━ ━ ━━ ━ ━ ━━━━   ━       [y/n/s/⏎] y
                                             ... some habits omitted ...
````

(Some weeks later)

```sh
    $ harsh log
                              ▄▃▃▄▄▃▄▆▆▆▅▆▆▇▆▄▃▄▆▃▆▃▆▂▅▄▃▄▅▆▅▃▃▃▆▂▄▅▄▅▅▅▆▄▄▆▇▆▅▅▄▃▅▆▄▆▃▃▂▅▆
                    Meditated ━       ━ ━  ━━         ━    ━   ━ ━   ━━━━━━━━━━━   ━ ━   ━━
        Cleaned the apartment ━──────                 ━──────           ━──────    •······
               Had a headache             ━  ━     ━━                  ━━   ━   ━━
                   Used harsh ━ ━━━ ━  ━━━   ━ ━ ━       ━ ━ ━  ━ ━ ━━ ━ ━ ━━━━   ━    ━
                                             ... some habits omitted ...

    Yesterday's score: 88.3%
```

The sparkline at the top give a graphical representation of each day's score.
The M W F stands for M(onday), W(ednesday), and F(riday) to provide visual
hinting as to which days are which and possibly help diagnoze whether
particular days are villains or heros in building your habits (How is this
ueful? As an example, because of then way I had meetings structured on Tuesday,
I often fell off the truck on certain habits on Tuesdays. Seeing this clearly
in the graphs let me restructure my Tuesday work days to get me back on track.)

The score at the bottom specifies how many of your habits you met that previous
day of total possible and removes any you may have skipped from the calculation.

### Subcommands

Run `harsh log stats` gives an analysis of your entire log file and a quantified
idea of days you've been on streak, how many days you broke your consistency
chain, and how many days you may have skipped -- as well as the total number of
days you've been tracking a particular habit for (note: I swap out my file every
year, but at least one person wanted this feature to track over 800+ days of log
files they'd converted from another app.). It's also nicely coloured to help
visually separate the information.

In particular, when you've been tracking for longer than 100 days (the visual
length of the consistency graph), you can get summary stats for your entire log
(tbh, I did not think this was useful until I implemented it. I was wrong.).
This can be surprisingly useful to see longer trends quantified.

```sh
               Slept 7h+  Streaks 173 days      Breaks 147 days Skips  1 days   Tracked 320 days
           Morning Pages  Streaks 310 days      Breaks 9 days   Skips  2 days   Tracked 320 days
             Share daily  Streaks 320 days      Breaks 0 days   Skips  1 days   Tracked 320 days
                Workouts  Streaks 246 days      Breaks 27 days  Skips  48 days  Tracked 320 days
              Read Daily  Streaks 302 days      Breaks 18 days  Skips  0 days   Tracked 320 days
             Write daily  Streaks 320 days      Breaks 0 days   Skips  1 days   Tracked 320 days
                     TIL  Streaks 285 days      Breaks 26 days  Skips  9 days   Tracked 320 days
                 Running  Streaks 117 days      Breaks 56 days  Skips  148 days Tracked 320 days
        Blog Fortnightly  Streaks 314 days      Breaks 6 days   Skips  0 days   Tracked 320 days
                                            ...
```

As you can see here, I need to work on sleep more than anything, but digging
down on these stats I shocked myself at skipped and breaks in workouts (covid
vaccine related in some cases), and how I need to rethink how some of these were
set up or I'm doing them (running, blogging etc.). The point is not how terrible
I am, but that looking into the numbers revealed patterns (sleep, affects
workouts, running, and TIL - today I learned - rather terribly). YMMV.

Run `harsh log <habit search term>` gives a slightly more in depth analysis of
individual habits in conjunction with your topline aparkline. The idea here is
that you can examine individual habits graphically against your topline to see
if there are patterns of correlation between the variation in an individual
habit and your overall daily score over time.

Say you expect there's a close correlation between you not getting good sleep
and your habits generally falling from a state of grace. It's easy to check
with a `log <habit substring>` and any string you type from your habits. In my
case, I have a habit around getting to bed on time and getting a solid 7 I call
`Bed by 12+Slept 7h+`. I can just use the handy `bed` as a short match term for
the habit and get the result.

```sh
    $ harsh log bed

                             █▇▇▇▇▇▇▇▇▇▇▇▇▇▇█▇▇▇▇▇▇█████▇██▇▇████████▇███████▇█████████████
        Bed by 12+Slept 7h+      ━ ━    ━ ━━━━    ━━━    ━━    ━ ━━━  ━• ━━━━━━━━ ━━━━ ━━━━

```

As we can see, there is a pretty close correlation here to not getting enough
sleep (or going to bed too late) and me hitting all my daily habits.

### Done

A done habit gives you a nice bright `━` on the consistency graph line. It's done.

Additionally, the app checks in future days if you are still within the "every
x days" period of performing the habit by drawing a dimmer `─` after the done
marker to let you know you've satisfied the requirement for that habit.

### Skips

Sometimes, it's impossible to exercise a habit cause life happens. If cleaning
the house is a habit you want to exercise, but you happen to be away on
a business trip, that is an impossibility. And sometimes, you decide to skip and
push the habit to the next period (or a simple day or so). Skips being selected
(s in the prompt) allows this to happen. A skip is denoted by a bright `•`.

Much like satisfied habits where you've performed them once in the period,
"skipified" habits let you know you're still withing the calculated grace period
of the skip with a lighter dot `·`.

### Warnings

harsh also has a warnings feature to help flag to you when you're in danger of
breaking your consistency graph. Harsh will give you a warning by showing a "!"
symbol in your upcoming habits.

For habits of every less than 7 days period, you get a warning sigil on the day
the chain will break if you do not perform the habit. For a week or longer,
you'll start to see a warning sigil of `1 + days/7` rounded down (eg. so,
2 weeks' warning would get you the sigil 3 days ahead of breaking the chain
etc.).

## Halps

Enter `harsh help` if you're lost:

```text
    λ ~/harsh help
    NAME:
    Harsh - habit tracking for geeks

    USAGE:
    harsh [global options] command [command options] [arguments...]

    VERSION:
    0.10.0

    DESCRIPTION:
    A simple, minimalist CLI for tracking and understanding habits.

    COMMANDS:
    ask, a   Asks and records your undone habits
    log, l   Shows graph of habits
    todo, t  Shows undone habits for today.
    help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
    --help, -h     show help (default: false)
    --no-color, -n  no colors in output (default: false)
    --version, -v  print the version (default: false)
```

## Quality of Life Usage Improvement

As you increasingly use `harsh` for tracking, you'll inevitably end up making
small errors in your log file or will want to edit your config file. While
people have suggested adding a `harsh config` command to the app, this feels
like overkill for the intended audience (geeks) but you can get the same effect
through using aliases.

As a simple quality of life improvement, add the following to your `bash`,
`zsh`, `fish` (what I use), or shell of choice:

```fish
alias h="harsh"
alias hc="nvim ~/.config/harsh/habits"
alias hl="nvim ~/.config/harsh/log"
```

For me, this equates to me ending up typing `h log` for the graph, `h ask` etc
etc. When I need to edit the log file because I was a bit too itchy with my
trigger finger (or decide I should add a note), `hl` is my friend.

## No Colour option

New from 0.8.22: if you are logging the output of `harsh log stat` and other
features which contains colour codes, you can instead use `--no-color` flag via
`harsh --no-color [command]` (or `harsh -n [command]`) to suppress colourized
output. This is also very helpful in n/vim with the `:.! harsh -n log stats`
command stanza to record harsh's output to n/vim's buffer to augment your
weekly, monthly, or daily tracking (learned that vim snippet from the feature
requester!).

Much like the above feature of accessing config and log files, you can alias
`harsh -n` in your shell if you prefer to suppress all colour output coming from
all harsh commands.

## License: MIT License

_harsh_ is free software. You can redistribute it and/or modify it under the
terms of the [MIT License](LICENSE).

## Contributing

Primo, check out the [Contributing guidelines](CONTRIBUTING.md).

1. Fork it (<https://github.com/wakatara/harsh/fork>)
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

## Contributors

- [Daryl Manning](https://daryl.wakatara.com) - creator, maintainer, and evil mastermind
- [JamesD](https://github.com/jnd-au) - fix for small terminal width panic.
- [Aisling](https://github.com/ais) - improved configuration discovery.
- [vchslv13](https://github.com/vchslv13) - improved shell installation script.
- [manu-cyber](https://github.com/manu-cyber) - documentation fixes.
- [nimaaskarian](https://github.com/nimaaskarian) - cli improvements and refactors.

## Thanks

- [Bjorn A](https://github.com/gaqzi) - for initial code review and improvements pre-release
- [James RC](https://github.com/yarbelk) - for initial code review and improvements pre-release
- [blinry](https://github.com/blinry) - for writing habitctl which harsh is an homage to and riff on.
