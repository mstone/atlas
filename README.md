## Overview

atlas is a BSD-licensed experimental wiki intended to explore low-latency
search, editing, and mapping of the kinds of [boundary
objects](http://en.wikipedia.org/wiki/Boundary_object) found in [complex
sociotechnical systems](http://mitpress.mit.edu/books/engineering-safer-world).

## Caveats

Warning: as initially published, atlas has [known issues](https://github.com/mstone/atlas/issues), several of
which have security implications that may make atlas inappropriate for use in
your environment.

## Dependencies

atlas:

  * build-depends on [Golang](http://golang.org), [sqlite3](http://sqlite.org),
    and several MIT- and Apache 2.0-licensed Golang libraries including 
    [glog](https://github.com/golang/glog),
    [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3), and
    [blackfriday](https://github.com/russross/blackfriday),

  * run-depends on [etherpad-lite](http://etherpad.org), and 

  * bundles [atom.go](https://code.google.com/p/go/source/browse/blog/atom/atom.go?repo=tools),
    [jQuery](http://jquery.org), [svg-edit](https://code.google.com/p/svg-edit/), 
    and several MIT-licensed jQuery plugins (Chosen, BBQ, and AutoSize).

## Use

This repository is intended to be mounted in your GOPATH, presently at
`$GOPATH/src/akamai/atlas`.

For ideas on how to run an atlas instance, please see our example
[setup.sh](./setup.sh) script.
