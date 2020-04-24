# Harsh Taskmaster

## Usage

harsh is habit tracking for geeks. A minimalist, command line tool for tracking and examining your habits.

Designed for simplicity, visibility, and longevity, it uses self-contained tracking files that are human-grokable and editable in your text editor of choice. It's much simpler, more understandable, and portable than commercial or mobile applications, and less complex than emacs habit tracking. While quantitative individual movement tracking is exhaustive, the important habits get lost in the deluge so this provides deliberated, explicit habits to track. It's a riff on blinary's habitctl (rop-off). It's written in GoLang and incorporates features that were never merged into habitctl such as skips, warns, bug fixes, and some other goodies. If you're a geek, I think you'll like it. Despite trying an exhaustive number of habit trackers, this was what worked for me. If you're interested int he whys, there a [launch blog post here on my motivations writing it](https://daryl.wakatara.com/harsh-a-minimalist-cli-habit-tracker).

Succintly: it's quick and gets out of your way. And gives you amazing visibililty on your habits.

## Installation

harsh is coded in [GoLang](https://golang.org). It's available on Linux, OSX, and Windows via Go's [goreleaser](https://github.com/goreleaser/goreleaser), as a specific goal was increasing take up and adoption of the CLI-based approach.

On OSX

```
brew install harsh
```

On linux

```
apt install harsh
```

If you want to compile it yourself, you can grab the source code from this repo and

```
go install harsh
```

I alias (or you can symlink) the `/usr/local/bin/harsh` executable to `h` and shorten commands to `h log`or similar. For me, strangely increases my use. YMMV.

## Usage

When you run ` harsh` for the first time, it will set up the required files:

```
    $ h
    Welcome to harsh!
    
    Created /Users/daryl/.config/harsh/habits  This file lists your tracked habits.
    Created /Users/daryl/.config/harsh/log     This file contains your habit log.
    
    You don't have any habits set up!
    Open the habits file and modify the habit list using a text editor.
    Then run `harsh ask` to start tracking
    'harsh todo` will show you what you have ledt today.
    `harsh log` will show you a consistency graph over your efforts
    (trust me, it gets way cooler looking over time)
    Happy tracking! I genuinely hope this helps you get better.
```

Open habits in your text editor of choice, such as nano, vim, VS Code, Sublime, or emacs. 

You'll see an example file like this:
```
    # The numbers specifies how often you want to do a habit:
    # 1 means daily, 7 means weekly, 0 means you're just tracking the habit. 
    # Some examples:

    Meditated :1
    Cleaned : 7
    Used harsh: 1
    Ate in restaurant: 0
```

The trick is figuring out what habits you want to track building or breaking. Too many, you'll fail. Too few, and the app loses its edge. Too short-term, you feel good but fail long-term. 

If you're getting started, try 5-10 to start missing short term and long term and see how you go. Tracking your habits is strangely a habit you *also* need to build. There're no right answers, but if this is new, [focus on foundational keystone habits](https://daryl.wakatara.com/resolution-keystone-habits-and-foundational-hacks/) that will feed future ones.

Here are some more ideas of what to track:

- Went to bed on time
- Got X hours of sleep
- Inbox zeroed
- practiced a language
- journalled
- blogged
- studied
- socialized
- tracked finances
- worked on a project
- cleaned house
- 2 coffees only
- thanked someone

Then, simply run `h` regularly, specify whether or not you did the habit from the prompt (or needed to skip the habit for some reason - eg. could not clean apartment because you were away for week), and get pretty graphs! 

The consistency graph shows your last 100 days (though bi plan to make a flag at some point.)

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

The sparkline at the top give a graphical representation of each day's score. The thick lines in the graph say when you did a habit, the thin lines say that you were within your threshold for doing it every x days. A thick dot implies you had to skip or were unable to exercise a habit for whatever reason, and a thin dot indicates the period for which a skip would normally be in effect (eg you could not clean your apartment because you were on a business trip).

The score specifies how many of the due habits you did that day of total possible and removes any you may have skipped from the calculation. 

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
    USAGE:
        harsh [SUBCOMMAND]
    
    FLAGS:
        -h, --help       Prints help information
        -V, --version    Prints version information
    
    SUBCOMMANDS:
        ask      Ask for status of all habits for a day
        help     Prints this message or the help of the given subcommand(s)
        log      Print habit log
        todo     Print unresolved tasks for today

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

- [Daryl Manning](https://github.com/wakatara) - creator, maintainer, and evil supervillain