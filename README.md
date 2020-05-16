# Harsh Taskmaster

## Usage

harsh is habit tracking for geeks. A minimalist, command line tool for tracking and examining your habits.

Succintly: it's quick and gets out of your way. And gives you amazing visibililty on your habits.

There're 3 commands: ask, log, and todo.

Designed for simplicity, visibility, and longevity, harsh uses simple text files for tracking that are human-grokable and editable in your favourite text editor. It's simpler, less messy, and more portable than commercial or mobile applications and less fussy to manage than emacs habit tracking (imho). While quantified individual tracking is exhaustive, important habits get lost in the data deluge so this provides deliberated, explicit habits to track. 

It's written in GoLang and adds features such as bug fixes, skips, warns, and commenting. If you're a geek, I think you'll like it. Despite trying an exhausting number of habit trackers, this was what worked for me. If you're interested in why I wrote it, there's [launch blog post on my motivations](https://daryl.wakatara.com/harsh-a-minimalist-cli-habit-tracker).

My biggest hope that it helps you accomplish what you're trying to do in your life.


## Installation

harsh is available on Linux, OSX, and Windows (note: currently untested on Windows), as a specific goal was increasing take up and adoption of a portable, command line, text-based approach.

On OSX

```
brew install wakatara/harsh/formula
```

which will also alert you to future updates when you `brew update`.


On Linux

```
apt install harsh
```

If you want to compile it yourself, you can grab the source code from this repo and

```
go install harsh
```

Personally, I alias the `harsh` executable to `h` and shorten commands to `h log` etc. Strangely, it increases my use of the app and makes tracking easier. YMMV.

## Usage

When you run `harsh` for the first time, it will set up the required files:

```
    $ h
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

On OSX and Linux based systems, the `habits` and `log` files will be under `~/.config/harsh/`. On Windows, you can find the files under 


Open the `habits` file in your text editor of choice (nano, vim, VS Code, Sublime, or emacs). 

You'll see an example file like this:
```
    # The numbers specifies how frequently you want to do a habit:
    # 1 means daily, 7 means weekly, 0 means you're just tracking the habit.
    # 0 frequency tracking habits will not flag warnings or count towards scoring. 
    # Examples:

    Gym: 3
    Cleaned: 7
    Used harsh: 1
    Called Mom: 7
```

The trick is figuring out what habits you want to track building or breaking. Too many, you'll fail. Too few, and the app loses its edge. Too short-term, you feel good but fail on longer-term objectives.

If you're getting started, try 5-8 and mix short term and long term and see how you go. Tracking your habits is *strangely* also a habit you need to build. There're no right answers, but if this is new, [focus on foundational keystone habits](https://daryl.wakatara.com/resolution-keystone-habits-and-foundational-hacks/) that will feed future ones.

Here are some more ideas of what to track:

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

Then, simply run `h` regularly, specify whether or not you did the habit from the prompt (or needed to skip the habit for some reason - eg. could not clean apartment because you were away for week), and get pretty graphs! 

The consistency graph shows your last 100 days.

```
    $ h
    2020-01-05:
                    Meditated ━       ━ ━  ━━         ━    ━   ━ ━   ━━━━━━━━━━━   ━ ━   ━[y/n/s/⏎] y
        Cleaned the apartment ━──────                 ━──────           ━──────    •······[y/n/s/⏎] n
               Had a headache             ━  ━     ━━                  ━━   ━   ━━        [y/n/s/⏎] n
                Used harsh    ━ ━━━ ━  ━━━   ━ ━ ━       ━ ━ ━  ━ ━ ━━ ━ ━ ━━━━   ━       [y/n/s/⏎] y
    2018-09-16+02:00:
                    Meditated        ━ ━  ━━         ━    ━   ━ ━   ━━━━━━━━━━━   ━ ━   ━ [y/n/s/⏎] y
        Cleaned the apartment ──────                 ━──────           ━──────    •······ [y/n/s/⏎] n
               Had a headache            ━  ━     ━━                  ━━   ━   ━━         [y/n/s/⏎] n
                Used harsh   ━ ━━━ ━  ━━━   ━ ━ ━       ━ ━ ━  ━ ━ ━━ ━ ━ ━━━━   ━    ━[y/n/s/⏎] y
```

(Some weeks later)

```
    $ h log
                              ▄▃▃▄▄▃▄▆▆▆▅▆▆▇▆▄▃▄▆▃▆▃▆▂▅▄▃▄▅▆▅▃▃▃▆▂▄▅▄▅▅▅▆▄▄▆▇▆▅▅▄▃▅▆▄▆▃▃▂▅▆
                    Meditated ━       ━ ━  ━━         ━    ━   ━ ━   ━━━━━━━━━━━   ━ ━   ━━
        Cleaned the apartment ━──────                 ━──────           ━──────    •······        
               Had a headache             ━  ━     ━━                  ━━   ━   ━━         
                Used harsh    ━ ━━━ ━  ━━━   ━ ━ ━       ━ ━ ━  ━ ━ ━━ ━ ━ ━━━━   ━    ━

                                             ... some habits omitted ...

    Yesterday's score: 88.3%
```

The sparkline at the top give a graphical representation of each day's score.

The score at the bottom specifies how many of your habits you fulfilled that day of total possible and removes any you may have skipped from the calculation.

### Done

A done habit gives you a nice bright `━` on the consistency graph line. It's done.

Additionally, the app checks in future days if you are still within the "everyx days" period of performing the habit by drawing a dimmer `─` after the done marker to let you know you've satisfied the requirement for that habit.

### Skips

Sometimes, it's impossible to exercise a habit cause life happens. If cleaning the house is a habit you want to exercise, but you happen to be away on a business trip, that is an impossibility. And sometimes, you decide to skip and push the habit to the next period (or a simple day or so). Skips being selected (s in the prompt) allows this to happen. A skip is denoted by a bright `•`.

Much like satisfied habits where you've performed them once in the period, "skipified" habits let you know you're still withing the grace period of the skip with a lighter dot `·`. 

### Warnings

harsh also has a warnings feature to help flag to you when you're in danger of breaking your consistency graph. Harsh will give you a warning by showing a "!" symbol in your upcoming habits.

For habits of every less than 7 days period, you get a warning sigil on the day the chain will break if you do not perform the habit. For a week or longer, you'll start to see a warning sigil of `1 + days/7` rounded down (eg. so, 2 weeks' warning would get you the sigil 3 days ahead of breaking the chain etc.).


Enter `h help` if you're lost:

    $ h help
        NAME:
        harsh - A minimalist habit tracking CLI for geeks

        USAGE:
        harsh [global options] command [command options] [arguments...]

        COMMANDS:
        ask, a   Asks about and logs your undone habits
        log, l   Shows graphs of habits log
        todo, t  Shows undone habits for today.
        help, h  Shows a list of commands or help for one command

        GLOBAL OPTIONS:
        --version value  Version of the Harsh app
        --help, -h       show help

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

* [Daryl Manning](https://github.com/wakatara) - creator, maintainer, and evil supervillain mastermind

## Thanks

* [Bjorn A](https://github.com/gaqzi) - for initial Go code review and improvements pre-release
* [James RC](https://github.com/yarbelk) - for initial Go code review and improvements pre-release
* [Blinary](https://github.com/blinry) - for writing habitctl which this is an homage to and riff off of.
