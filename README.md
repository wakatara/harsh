# Harsh Taskmaster

## Usage

Harsh is a minimalist habit tracking command line tool for tracking and examining your habits. 

It's inspired by and a GoLang riff on blinary's habitctl. It's designed to have simple, contained (eg. one year) habit tracking files that are human grokable and easily edited in a text editor. It's much simpler, more understandable, and more portable than commercial application and much easier to manage than emacs native org-mode habit tracking. It incorporates a few features I wrote in habitctl that were not merged such as skips and warns and a few other goodies.

Succintly: it's quick and gets out of your way. 

I personally have it in a `tmux` tab that track against as I tick items off.

## Installation

harsh is coded in [GoLang](https://golang.org). It is avaiable on Linux, OSX, and Windows.

On OSX

```
brew install harsh
```

On linux

```
apt install harsh
```

If you want to compile it yourself, you can: 

```
go get 
```

I symlink `/usr/local/bin` to the executable so commands shorten to `h log` or similar. Makes it vastly easier to track for some reason.

## Usage

When you run ` harsh` for the first time, it will set up the required files:

    $ h
    Welcome to  harsh!
    
    Created /home/seb/. harsh/habits. This file will list your currently tracked habits.
    Created /home/seb/. harsh/log. This file will contain your habit log.
    
    You don't have any habits set up!
    Run `harsh edith` to modify the habit list using your default $EDITOR.
    Then, run ` harsh`! Happy tracking!

Run `h edith` and change the content of the habits file, for example like this:

    # The numbers specifies how often you want to do a habit:
    # 1 means daily, 7 means weekly, 0 means you're just tracking the habit. Some examples:

    Meditated :1 
    Cleaned the apartment: 7
    Had headache: 0
    Used harsh: 1

Here are some more ideas of what to track:

- got up at a certain time
- used a spaced repetition software like Anki
- took a multivitamin
- cleared my email inbox
- answered to all texts
- visited and read all Slack workspaces
- practiced a language
- self reflection/used a diary
- autogenic training
- published something on my blog/homepage
- worked on a project
- did the dishes
- tidied the apartment
- closed all browser tabs
- tracked caloric intake
- happy
- flow
- relaxed
- coffee intake
- left my comfort zone
- thanked someone

Then, simply run `h` regularly, specify whether or not you did the habit (or needed to skip the habit for some reason - eg. could not clean apartment because you were away for week), and get pretty graphs! 

The consistency graph shows your last 100 days (though bi plan to make a flag at some point.)

```
    $ h
    2018-09-15:
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

    Yesterday's score: 73.3%
```

The score specifies how many of the due habits you did that day and removes any you may have skipped from the calculation. The sparkline at the top give a graphical representation of the score. The thick lines in the graph say that you did the habit, the thin lines say that that it was okay that you didn't to it. A thick dot implies you had to skip or were unable to exercise a habit for whatever good reason, and a thin dot indicates the period for which a skip would normally be in effect (in the example above, we are suggesting you could not clean your apartment because you were on a business trip when you'd normally clean it and can thus excuse yourself.).

Also, on the day, if a habit chain is in danger of breaking because it's the last day you can do it before the consistency graph would have a gap in it, harsh will give you a warning by turning the " " symbol you'd see on the day into an "!".

Enter `h help` if you're lost:

    $ h help
    USAGE:
        harsh [SUBCOMMAND]
    
    FLAGS:
        -h, --help       Prints help information
        -V, --version    Prints version information
    
    SUBCOMMANDS:
        ask      Ask for status of all habits for a day
        edit     Edit habit log file
        edith    Edit list of current habits
        help     Prints this message or the help of the given subcommand(s)
        log      Print habit log
        todo     Print unresolved tasks for today

## License: MIT License

*harsh* is free software. You can redistribute it and/or modify it under the terms of the [MIT License](LICENSE).


TODO: Write usage instructions here

## Development

TODO: Write development instructions here

## Contributing

1. Fork it (<https://github.com/wakatara/harsh/fork>)
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

## Contributors

- [Daryl Manning](https://github.com/wakatara) - creator and maintainer