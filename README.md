# StreamDeck UI

# Help Wanted!

If you want to help with the development of streamdeckd and it's related repos, either by submitting code, finding/fixing bugs, or just replying to issues, please join this discord server: https://discord.gg/mgNAKuk5

This repository contains a graphical configuration tool for the StreamDeck
devices that works on Unix, Linux and other OS.
It is a work heavily in progress and we welcome any contributions.

# Dependencies

dbus & zenity

For Debian/Ubuntu and Linux Mint users fallowing depencencies are needed
```
sudo apt install libgl1-mesa-dev xorg-dev
```

# Usage

To use the streamdeck on unix you will need to have a daemon running.
This GUI is built to work with unix-streamdeck/streamdeckd, the install steps
are the following to build from code. You will only need a Go compiler.

```bash
$ go get github.com/unix-streamdeck/streamdeckd
$ `go env path`/bin/streamdeckd
```

Once that is running (you should probably plug in your streamdeck device first)
you can install this package

```bash
$ go get github.com/unix-streamdeck/streamdeckui
$ `go env path`/bin/streamdeckui
```

# Screenshot

![](img/current.png)

