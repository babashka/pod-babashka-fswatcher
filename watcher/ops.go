package watcher

import (
        "encoding/json"
        "fmt"
        "os"
        "time"
        "path/filepath"
        "strings"

        "github.com/babashka/pod-fswatcher/babashka"
        "github.com/fsnotify/fsnotify"
)

type Opts struct {
        DelayMs   uint64 `json:"delay-ms"`
        Recursive bool   `json:"recursive"`
}

type Response struct {
        Type  string  `json:"type"`
        Path  string  `json:"path"`
        Dest  *string `json:"dest,omitempty"`
        Error *string `json:"error,omitempty"`
}

type WatcherInfo struct {
        WatcherId int    `json:"watcher-id"`
        Type      string `json:"type"`
}

var (
        watcher_idx = 0
        watchers    = make(map[int]*fsnotify.Watcher)
)

func allFiles(dir string) ([]string, error) {
        fileInfo, err := os.Stat(dir)
        if err != nil {
                return nil, err
        }

        if !fileInfo.IsDir() {
                return []string{dir}, nil
        }

        files := []string{}
        err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
                if err != nil {
                        return err
                }

                files = append(files, path)
                return nil
        })
        if err != nil {
                return nil, err
        }

        return files, nil
}

func debounce(delay time.Duration, input chan fsnotify.Event) chan *fsnotify.Event {
        output := make(chan *fsnotify.Event)
	//timer :=make(chan time.Time)
	timer := time.After(delay)

        go func() {

                // Start debouncing
                for {
                        select {
                        case <-timer:
                                timer = time.After(delay)
                                for {
					select {
					case event, ok := <-input:
						if !ok {
							return
						}
						output <- &event
					}
				}
                        }
                }
        }()

        return output
}

func watch(message *babashka.Message, path string, opts Opts) (*WatcherInfo, error) {
        watcher, err := fsnotify.NewWatcher()
        if err != nil {
                return nil, err
        }

        if opts.Recursive {
                files, err := allFiles(path)
                if err != nil {
                        return nil, err
                }

                for _, file := range files {
                        err = watcher.Add(file)
                }
        } else {
                err = watcher.Add(path)
        }
        if err != nil {
                return nil, err
        }

        debounced := debounce(time.Millisecond * 5000, watcher.Events)

        go func() {
                for {
                        select {
                        case event := <-debounced:
                                // if !ok {
                                //      return
                                // }
                                babashka.WriteInvokeResponse(
                                        message,
                                        Response{strings.ToLower(event.Op.String()), event.Name, nil, nil},
                                )
                        case err, ok := <-watcher.Errors:
                                if !ok {
                                        return
                                }
                                msg := err.Error()
                                json, _ := json.Marshal(Response{"error", path, nil, &msg})
                                babashka.WriteInvokeResponse(message, json)
                        }
                }
        }()

        watcher_idx++
        watchers[watcher_idx] = watcher

        return &WatcherInfo{watcher_idx, "watcher-info"}, nil
}

func ProcessMessage(message *babashka.Message) (interface{}, error) {
        switch message.Op {
        case "describe":
                return &babashka.DescribeResponse{
                        Format: "json",
                        Namespaces: []babashka.Namespace{
                                {
                                        Name: "pod.babashka.filewatcher",
                                        Vars: []babashka.Var{
                                                {
                                                        Name: "watch",
                                                        Code: `
(defn watch
  ([path cb]
   (watch path cb {}))
  ([path cb opts]
   (babashka.pods/invoke
     "pod.babashka.filewatcher"
     'pod.babashka.filewatcher/watch*
     [path opts]
     {:handlers {:success (fn [event]
                            (cb (update event :type keyword)))
                 :error   (fn [{:keys [:ex-message :ex-data]}]
                            (binding [*out* *err*]
                              (println "ERROR:" ex-message)))}})
   nil))`,
                                                },
                                                {
                                                        Name: "watch*",
                                                },
                                                {
                                                        Name: "unwatch",
                                                },
                                        },
                                },
                        },
                }, nil
        case "invoke":
                switch message.Var {
                case "pod.babashka.filewatcher/watch*":
                        args := []json.RawMessage{}
                        err := json.Unmarshal([]byte(message.Args), &args)
                        if err != nil {
                                return nil, err
                        }

                        opts := Opts{DelayMs: 2000, Recursive: false}
                        err = json.Unmarshal([]byte(args[1]), &opts)
                        if err != nil {
                                return nil, err
                        }

                        var path string
                        json.Unmarshal(args[0], &path)

                        return watch(message, path, opts)
                case "pod.babashka.filewatcher/unwatch":
                        args := []int{}
                        err := json.Unmarshal([]byte(message.Args), &args)
                        if err != nil {
                                return nil, err
                        }

                        idx := args[0]
                        _, ok := watchers[idx]
                        if ok {
                                watchers[idx].Close()
                                delete(watchers, idx)
                        }

                        return nil, nil
                default:
                        return nil, fmt.Errorf("Unknown var %s", message.Var)
                }
        default:
                return nil, fmt.Errorf("Unknown op %s", message.Op)
        }
}
