package watcher

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/babashka/pod-babashka-fswatcher/babashka"
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
	WatcherId int `json:"watcher/id"`
}

type FsWatcher struct {
	Watcher     *fsnotify.Watcher
	Opts        *Opts
	WatcherInfo *WatcherInfo
	Path        string
}

var (
	watcher_idx = 0
	watchers    = make(map[int]*FsWatcher)
)

func listDirRec(dir string) ([]string, error) {
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
	go func() {
		for {
			select {
			case event, ok := <-input:
				if !ok {
					return
				}
				output <- &event
			case <-time.After(50 * time.Millisecond):
				time.Sleep(delay)
			}
		}
	}()

	return output
}

func startWatcher(message *babashka.Message, watcherId int) error {
	fsWatcher := watchers[watcherId]
	opts := fsWatcher.Opts
	path := fsWatcher.Path
	watcher := fsWatcher.Watcher

	if opts.Recursive {
		files, err := listDirRec(path)
		if err != nil {
			return err
		}

		for _, file := range files {
			err = watcher.Add(file)
		}
	} else {
		if err := watcher.Add(path); err != nil {
			return err
		}
	}

	debounced := debounce(time.Millisecond*time.Duration(opts.DelayMs), watcher.Events)

	go func() error {
		for {
			select {
			case event := <-debounced:
				err := babashka.WriteInvokeResponse(
					message,
					Response{strings.ToLower(event.Op.String()), event.Name, nil, nil},
				)
				if err != nil {
					msg := err.Error()
					json, _ := json.Marshal(Response{"error", path, nil, &msg})
					babashka.WriteInvokeResponse(message, json)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return err
				}
				msg := err.Error()
				json, _ := json.Marshal(Response{"error", path, nil, &msg})
				babashka.WriteInvokeResponse(message, json)
			}
		}
	}()

	return nil
}

func createWatcher(message *babashka.Message, path string, opts Opts) (*WatcherInfo, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watcher_idx++
	info := WatcherInfo{watcher_idx}
	fsWatcher := FsWatcher{watcher, &opts, &info, path}
	watchers[watcher_idx] = &fsWatcher

	return fsWatcher.WatcherInfo, nil
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
							Name: "-create-watcher",
						},
						{
							Name: "watch",
							Code: `
(defn watch
  ([path cb]
   (println :watch)
   (watch path cb {}))
  ([path cb opts]
   (let [ret (pod.babashka.filewatcher/-create-watcher path opts)]
     (babashka.pods/invoke
       "pod.babashka.filewatcher"
       'pod.babashka.filewatcher/-start-watcher
       [(:watcher/id ret)]
       {:handlers {:success (fn [event]
                              (cb (update event :type keyword)))
                   :error   (fn [{:keys [:ex-message :ex-data]}]
                              (binding [*out* *err*]
                                (println "ERROR:" ex-message)))}})
   ret)))`,
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
		case "pod.babashka.filewatcher/-create-watcher":
			args := []json.RawMessage{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}

			opts := Opts{DelayMs: 2000, Recursive: false}
			if err := json.Unmarshal([]byte(args[1]), &opts); err != nil {
				return nil, err
			}

			var path string
			json.Unmarshal(args[0], &path)

			return createWatcher(message, path, opts)

		case "pod.babashka.filewatcher/-start-watcher":
			args := []int{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}

			return startWatcher(message, args[0]), nil

		case "pod.babashka.filewatcher/unwatch":
			args := []WatcherInfo{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}

			watcher := args[0]
			idx := watcher.WatcherId

			_, ok := watchers[idx]
			if ok {
				watchers[idx].Watcher.Close()
				delete(watchers, idx)
			}

			return watcher, nil

		default:
			return nil, fmt.Errorf("Unknown var %s", message.Var)
		}
	default:
		return nil, fmt.Errorf("Unknown op %s", message.Op)
	}
}
