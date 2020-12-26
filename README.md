# pod-fswatcher

A [babashka pod](https://github.com/babashka/babashka.pods) for watching files.
Implemented using the Go [fsnotify](https://github.com/fsnotiy/fsnotify) library.

## Build & Run

Run in [babashka](https://github.com/borkdude/babashka/) or using the
[babashka.pods](https://github.com/babashka/babashka.pods) library on the JVM.

- [Go](https://golang.org/dl/) 1.15+ should be installed
- Clone this repo
- Run `go build -o fswatcher main.go` to compile the binary `fswatcher`

``` clojure
(require '[babashka.pods :as pods])
(pods/load-pod "/path/to/fswatcher")

(require '[pod.babashka.filewatcher :as fw])

(fw/watch "/dir-or-file/to/watch" (fn [event] (prn event)) {:delay-ms 50})
```

As a result of the following terminal sequence:

``` shell
$ touch created.txt
$ mv created.txt created_renamed.txt
$ chmod -w created_renamed.txt
$ chmod +w created_renamed.txt
$ echo "foo" >> created_renamed.txt
```

the following will be printed:

``` clojure
{:path "/private/tmp/created.txt", :type :create}
{:path "/private/tmp/created.txt", :type :notice/remove}
{:dest "/private/tmp/created_renamed.txt", :path "/private/tmp/created.txt", :type :rename}
{:path "/private/tmp/created_renamed.txt", :type :chmod}
{:path "/private/tmp/created_renamed.txt", :type :chmod}
{:path "/private/tmp/created_renamed.txt", :type :notice/write}
{:path "/private/tmp/created_renamed.txt", :type :write}
```

## Watching and UnWatching
To watch on a path P:
```clojure
(fw/watch P (fn [event] (prn event)))
```
This returns a map:
```clojure
{:watcher-id 3
 :type       "watcher-info"}
```
The `watcher-id` uniquely identifies an active watcher in the pod.
More wathchers can be created with calls to `(fw/watch ...)` each returning a unique id.
Multiple watchers can be used on the same dir.

To stop the watcher:
```clojure
(fw/unwatch num-of-the-watcher)
```
This call is idempotent and never fails.
