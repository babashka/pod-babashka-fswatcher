# pod-babashka-fswatcher

A [babashka pod](https://github.com/babashka/babashka.pods) for watching files.

Implemented using the Go [fsnotify](https://github.com/fsnotiy/fsnotify) library.

## Status

Experimental.

## Usage

Load the pod:

``` clojure
(require '[babashka.pods :as pods])
(pods/load-pod 'org.babashka/fswatcher "0.0.1")

(require '[pod.babashka.fswatcher :as fw])
```

Watchers can be created with `fw/watch`:

```clojure
(def watcher (fw/watch "src" (fn [event] (prn event))))
```

You can create multiple watchers that run concurrently, even on the same
directory.

The `watch` function returns a value which can be passed to `unwatch` which
stops and cleans up the watcher:

```clojure
(fw/unwatch watcher)
```

See [test/script.clj](test/script.clj) for an example test script.

### Watch recursively

By default watchers do not watch recursively. Pass `{:recursive true}` in the
options map to enable it.

```clojure
(fw/watch "src" (fn [event] (prn event)) {:recursive true})
```

## Build

### Requirements

- [Go](https://golang.org/dl/) 1.15+ should be installed.
- Clone this repo.
- Run `go build -o pod-babashka-fswatcher main.go` to compile the binary.

## License

## License

Copyright © 2020 Rahuλ Dé and Michiel Borkent

License: TODO
