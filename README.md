# Harsh Taskmaster

harsh is habit tracking for geeks. A minimalist, command line tool for tracking and understanding your habits.

Succintly: it's quick and gets out of your way. And gives you amazing visibililty on your habits.

There are 3 commands: `ask`, `log`, and `todo`.

Designed for simplicity, visibility, and longevity, harsh uses simple text files for tracking that are human-grokable and editable in your favourite text editor. It's simpler, less messy, and more portable than commercial or mobile applications and less fussy to manage than emacs habit tracking (imho). While quantified individual tracking is exhaustive, important habits get lost in the data deluge so this provides deliberated, explicit habits to track.

It's written in GoLang and adds features on top of habitctl (its inspiration) such as bug fixes, skips, warns, and commenting. If you're a geek, I think you'll like it. Despite trying an exhaustive number of habit trackers, this was what worked for me. YMMV. If you're interested in why I wrote it, there's a [launch post about my motivations on my blog](https://daryl.wakatara.com/harsh-a-minimalist-cli-habit-tracker).

My biggest hope that it helps you get done what you're trying to get done in your life.

## Installation

harsh is available on Linux, OSX (homebrew too), and Windows as a specific goal was increasing take up and adoption of a portable, command line, text-based approach. Windows is completely untested at the moment so any help verifying it works as advertised and correcting any bugs appreciated.

### Install the pre-compiled binary

**homebrew tap** (MacOS only):

```
$ brew install wakatara/tap/harsh
```
(this will also nicely alert you of updates when you `brew update`)

**snapcraft**:

```
$ sudo snap install harsh
```

**manually**:

Download the pre-compiled binaries from the releases page and copy to the desired location.


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

You can then move the file to the desired location. 

## Usage

When you run `harsh` for the first time, it will set up the required files:

```
    $ harsh
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

On OSX and Linux based systems, the `habits` and `log` files will be under `~/.config/harsh/`. On Windows, you can find the files under `~\AppData\`


Open the `habits` file in your text editor of choice (nano, vim, VS Code, Sublime, or emacs). 

You'll see an example file like this:

``` harsh
    # This is your habits file.
    # It tells harsh what to track and how frequently in days.
    # 1 means daily, 7 means weekly, 14 every two weeks.
    # 0 is for tracking a habit. 0 frequency habits will not warn or score.
    # Examples:

    Gymmed: 2
    Bed by midnight: 1
    Cleaned House: 7
    Called Mom: 7
    Tracked Finances: 15
    New Skill: 90
    Too much coffee: 0
    Used harsh: 0
```

For each habit, pick an integer number of days that you want to be repeating that habit in. If it's not obvious, habits can have any character that is not a `:` as that delimits the period. We also use `:` as the separator in log files as well for easy parsing.

The real trick is figuring out what habits you want to track building or breaking. Too many, you'll fail. Too few, and the app loses its edge. Too short-term, you feel good but fail on longer-term objectives.

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

Then, simply run `harsh ask` regularly, specify whether you did the habit from the prompt (or needed to skip the habit for some reason - eg. could not clean apartment because you were away for week), and get pretty graphs! 

The consistency graph shows your last 100 days.

```
    $ harsh ask
    2020-01-05:
                    Meditated ━       ━ ━  ━━         ━    ━   ━ ━   ━━━━━━━━━━━   ━ ━   ━[y/n/s/⏎] y
        Cleaned the apartment ━──────                 ━──────           ━──────    •······[y/n/s/⏎] n
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

The score at the bottom specifies how many of your habits you met that day of total possible and removes any you may have skipped from the calculation.

### Done

A done habit gives you a nice bright `━` on the consistency graph line. It's done.

Additionally, the app checks in future days if you are still within the "everyx days" period of performing the habit by drawing a dimmer `─` after the done marker to let you know you've satisfied the requirement for that habit.

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
    0.8.1

    DESCRIPTION:
    A simple, minimalist CLI for tracking and understanding habits.

    COMMANDS:
    ask, a   Asks and records your undone habits
    log, l   Shows graph of habits
    todo, t  Shows undone habits for today.
    help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
    --help, -h     show help (default: false)
    --version, -v  print the version (default: false)
```

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

## Thanks

* [Bjorn A](https://github.com/gaqzi) - for initial code review and improvements pre-release
* [James RC](https://github.com/yarbelk) - for initial code review and improvements pre-release
* [Blinary](https://github.com/blinry) - for writing habitctl which harsh is an homage to and riff on.
