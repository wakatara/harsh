# Harsh Taskmaster

harsh is habit tracking for geeks. A minimalist, command line tool for tracking and understanding your habits.

Succinctly: it's quick and gets out of your way. And gives you excellent visibility on your habits.

There're 3 commands: `ask`, `log`, and `todo`.
(and one subcommand `log stats`)

Designed for simplicity, visibility, and longevity, harsh uses simple text files for tracking that are human-grokable and editable in your favourite text editor. It's simpler, less messy, and more portable than commercial or mobile applications and less fussy to manage than emacs habit tracking (imho). While quantified individual tracking is exhaustive, important habits get lost in the data deluge, so this provides deliberate, explicit, habits to track.

It's written in GoLang and adds features on top of habitctl (its inspiration) such as bug fixes, skips, and streak break warnings. If you're a geek, I think you'll like it. Despite trying an exhaustive number of habit trackers, this was what worked for me. YMMV. If you're interested in why I wrote it, there's a [launch post about my motivations on my blog](https://daryl.wakatara.com/harsh-a-minimalist-cli-habit-tracker).

My biggest hope that it helps you get done what you're trying to get done in your life.

## Installation

harsh is available on OSX (as a homebrew too), Linux (also as a Snap and homebrew), and Windows. A specific goal was increasing uptake and adoption of a portable, command line, text-based approach. Windows is somewhat untested though we had a nice pull request that helped with config issues. We have also updated harsh to support ARM architecture for OSX (M1 chipped Macs) and Linux as of 0.8.8.
### Install via package manager

**Homebrew tap** (Mac and Linux):

```
$ brew install wakatara/tap/harsh
```
(this will also nicely alert you of updates when you `brew update`)

**Snap** (Linux only):

```
$ sudo snap install harsh
```
(you'll also get alerted when there are updates)

### Easy one-line install

If you're not using a package manager, by far the easiest way to install harsh and get started is to use this one-line bash command if you're on OSX or Linux:

```bash
curl -sSLf https://raw.githubusercontent.com/wakatara/harsh/master/install.sh | sh
```

The script downloads the latest release directly from github, verifies it cryptographically, and then moves the executable to `usr/local/bin` ready for immediate use on the command line.

### With working Go environment

You may prefer a direct Go install if you have a working Go environment and Go 1.14+.

``` bash
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

If you want to build from source cause you like that sort of thing, follow these steps:

**Clone:**

```
$ git clone https://github.com/wakatara/harsh
$ cd harsh
```

**Get the dependencies:**

```
$ go get ./...
```

**Build:**

```
$ go build -o harsh .
```

**Verify it works:**

```
$ ./harsh --version
```

You can then move the file to the desired location in your `$PATH`.

## Getting Started

When you run `harsh ask` for the first time, it will set up the required files:

```
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

On OSX and Linux based systems, the `habits` and `log` files will be under `~/.config/harsh/`. On Windows, you can find the files under `%APPDATA%\harsh`.

Alternatively, you can set a different directory using the `HARSHPATH` environment variable.

| :warning: **WARNING**      |
|:---------------------------|
| Ubuntu's snap overrides (sensible) defaults and forcibly places harsh's config *and* log files in `~/snap/harsh/current/`. If you `snap remove` the harsh app, snap's uninstaller will nuke your config *and* log files with the sandbox and you may lose your config and log data if it's not backed up (please *always* exercise a good backup regime). For this reason, we *highly* recommend snap users set the `HARSHPATH` env variable to `~/.config/harsh/` and move config and log files there right after installing to protect them. Or *never* uninstall harsh. :grin: |

Open the `habits` file in your text editor of choice (nano, vim, VS Code, Sublime, or emacs). 

You'll see an example file like this:

``` harsh
    # This is your habits file.
    # It tells harsh what to track and how frequently in days.
    # 1 means daily, 7 means weekly, 14 every two weeks.
    # 0 is for tracking a habit. 0 frequency habits will not warn or score.
    # Examples:

    ! Dailies
    Gymmed: 1
    Bed by midnight: 1
    
    ! Weeklies
    Cleaned House: 7
    Called Mom: 7
    
    ! Monthly+
    Did Finances: 30
    New Skill: 90

    ! Tracking
    Too much coffee: 0
    Used harsh: 0
```

For each habit, pick an integer number of days that you want to be repeating that habit in. If it's not obvious, habits can have any character that is not a `:` as that delimits the period. We also use `:` as the separator in log files as well for easy parsing.

Headings are denoted by a "!" at the start of a line and allow you to categorize habits visually (useful if you have a lot of them).

Comments can go in the habits file by starting a line with "#" and are not parsed by the program.

The real trick of tracking is figuring out what habits you want to track building or breaking. Too many, you'll fail. Too few, and the app loses its edge. Too short-term, you feel good but fail on longer-term objectives.

If you're getting started, try 5-8 and mix short term and long term and see how you go. Tracking your habits is *strangely* also a habit you need to build. There're no right answers, but if this is new, [focus on foundational keystone habits](https://daryl.wakatara.com/resolution-keystone-habits-and-foundational-hacks/) that will feed future ones. If you're coming into this cold, I'd also recommend a good read of James Clear's Atomic Habits. 

Here are some ideas of what to track:

- Went to bed on time
- Got X hours of sleep
- Inbox zeroed
- Practiced language Y
- Journalled
- Blogged
- Studied genomics
- Socialized
- Stuck to budget
- Did open source
- Cleaned house
- 2 coffees only
- Thanked someone
- Went for a walk
- Told SO they're amazing

## Usage and Commands

Simply run `harsh ask` regularly, specify whether you did the habit from the prompt on a particular day (or needed to skip the habit for some reason - eg. could not clean apartment because you were away for week), and get pretty graphs!

`harsh ask` allows you to pick between `[y/n/s/⏎]` which is yes/no/skip/don't answer right now. CTRL-c breaks you out of the ask cycle at any point and returns you to your prompt.

As of `0.8.12`, to support longer term pattern tracking (and unreliable memories), you can optionally follow any of the `y | n | s` options with a typed `#` symbol and a comment that will get recorded in your log file. This is primarily for analysis at a later date to help uncover patterns between events, reasons you may have written a comment, and either good or bad knock-on effects. The comment gets fed into the log file so can be easily loaded along into spreadsheets or pandas for analysis on the `:` deliimiter. 

Personally, I'd use the comments sparing and to denote why you had to skip, broke a consistency chain with an `n`, or for when you're trying to figure out something notable on a day so when you look back you can see why habit *X* may or may not have succeeded. 

`harsh log` The consistency graph shows your last 100 days.

```
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
```

(Some weeks later)

```
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

The score at the bottom specifies how many of your habits you met that previous day of total possible and removes any you may have skipped from the calculation.

### Subcommands

Run `harsh log stats` gives an analysis of your entire log file and a quantified idea of days you've been on streak, how many days you broke your consistency chain, and how many days you may have skipped -- as well as the total number of days you've been tracking a particular habit for (note: I swap out my file every year, but at least one person wanted this feature to track over 800+ days of log files they'd converted from another app.). It's also nicely coloured to help visually separate the information.

In particular, when you've been tracking for longer than 100 days (the visual length of the consistency graph), you can get summary stats for your entire log (tbh, I did not think this was useful until I implemented it. I was wrong.). This can be surprisingly useful to see longer trends quantified.

```
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
As you can see here, I need to work on sleep more than anything, but digging down on these stats I shocked myself at skipped and breaks in workouts (covid vaccine related in some cases), and how I need to rethink how some of these were set up or I'm doing them (running, blogging etc.). The point is not how terrible I am, but that looking into the numbers revealed patterns (sleep, affects workouts, running, and TIL - today I learned - rather terribly). YMMV.


### Done

A done habit gives you a nice bright `━` on the consistency graph line. It's done.

Additionally, the app checks in future days if you are still within the "every x days" period of performing the habit by drawing a dimmer `─` after the done marker to let you know you've satisfied the requirement for that habit.

### Skips

Sometimes, it's impossible to exercise a habit cause life happens. If cleaning the house is a habit you want to exercise, but you happen to be away on a business trip, that is an impossibility. And sometimes, you decide to skip and push the habit to the next period (or a simple day or so). Skips being selected (s in the prompt) allows this to happen. A skip is denoted by a bright `•`.

Much like satisfied habits where you've performed them once in the period, "skipified" habits let you know you're still withing the grace period of the skip with a lighter dot `·`. 

### Warnings

harsh also has a warnings feature to help flag to you when you're in danger of breaking your consistency graph. Harsh will give you a warning by showing a "!" symbol in your upcoming habits.

For habits of every less than 7 days period, you get a warning sigil on the day the chain will break if you do not perform the habit. For a week or longer, you'll start to see a warning sigil of `1 + days/7` rounded down (eg. so, 2 weeks' warning would get you the sigil 3 days ahead of breaking the chain etc.).

## Halps

Enter `harsh help` if you're lost:


``` text
    λ ~/harsh help 
    NAME:
    Harsh - habit tracking for geeks

    USAGE:
    harsh [global options] command [command options] [arguments...]

    VERSION:
    0.8.10

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

As you increasingly use `harsh` for tracking, you'll inevitably end up making small errors in your log file or will want to edit your config file. While people have suggested adding a `harsh config` command to the app, this feels like overkill for the intended audience (geeks) but you can gert the same effect through using aliases. 

As a simple quality of life improvement, add the following to your `bash`, `zsh`, `fish` (what I use), or shell of choice:

``` bash
alias h="harsh"
alias hc="nvim ~/.config/harsh/habits"
alias hl="nvim ~/.config/harsh/log"
```

For me, this equates to me ending up typing `h log` for the graph, `h ask` etc etc. When I need to edit the log file because I was a bit too itchy with my trigger finger (or decide I should add a note), `hl` is my friend.

## No Colour option

New from 0.8.22: if you are logging the output of `harsh log stat` and other features which contains colour codes, you can instead use `--no-color` flag via `harsh --no-color [command]` (or `harsh -n [command]`) to suppress colourized output. This is also very helpful in n/vim with the `:.! harsh -n log stats` command stanza to record harsh's output to n/vim's buffer (learned that from the feature requester!).

Much like the above feature of accessing config and log files, you can alias `harsh -n` in your shell if you prefer to suppress all colour output coming from all harsh commands.

## License: MIT License

*harsh* is free software. You can redistribute it and/or modify it under the terms of the [MIT License](LICENSE).

## Contributing

Primo, check out the [Contributing guidelines](CONTRIBUTING.md).

1. Fork it (<https://github.com/wakatara/harsh/fork>)
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

## Contributors

* [Daryl Manning](https://daryl.wakatara.com) - creator, maintainer, and evil mastermind
* [Aisling](https://github.com/ais) - improved configuration discovery.
* [vchslv13](https://github.com/vchslv13) - improved shell installation script.
* [manu-cyber](https://github.com/manu-cyber) - documentation fixes.

## Thanks

* [Bjorn A](https://github.com/gaqzi) - for initial code review and improvements pre-release
* [James RC](https://github.com/yarbelk) - for initial code review and improvements pre-release
* [blinry](https://github.com/blinry) - for writing habitctl which harsh is an homage to and riff on.

