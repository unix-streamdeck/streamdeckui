# StreamDeck UI

This repository contains a graphical configuration tool for the StreamDeck
devices that works on Unix, Linux and other OS.
It is a work heavily in progress and we welcome any contributions.

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

