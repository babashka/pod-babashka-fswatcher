# pod-babashka-fswatcher

A [babashka pod](https://github.com/babashka/babashka.pods) for watching files.

Implemented using the Go [fsnotify](https://github.com/fsnotify/fsnotify) library.

## Status

Experimental.

## Usage

Load the pod:

``` clojure
(require '[babashka.pods :as pods])
(pods/load-pod 'org.babashka/fswatcher "0.0.5")

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

### Usage in bb.edn

In babashka 0.8.0 it is possible to specify pods in `bb.edn`:

``` clojure
{:pods {org.babashka/fswatcher {:version "0.0.7"}}
 :tasks {watch {:requires ([pod.babashka.fswatcher :as fw])
                :task (do (fw/watch "project.clj"
                                    (fn [event]
                                      (when (#{:write :write|chmod} (:type event))
                                        (println "hello!"))))
                          (deref (promise)))}}}
```

### Watch recursively

By default watchers do not watch recursively. Pass `{:recursive true}` in the
options map to enable it.

```clojure
(fw/watch "src" (fn [event] (prn event)) {:recursive true})
```

## Build

### Requirements

- [Go](https://golang.org/dl/) 1.18+ should be installed.
- Clone this repo.
- Run `go build -o pod-babashka-fswatcher main.go` to compile the binary.

## License

Copyright © 2020 Rahul De and Michiel Borkent

License: [BSD 3-Clause](https://opensource.org/licenses/BSD-3-Clause)
