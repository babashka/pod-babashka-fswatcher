package watcher

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/babashka/pod-fswatcher/babashka"
	"github.com/fsnotify/fsnotify"
)

type Opts struct {
	DelayMs uint64 `json:"delay-ms"`
}

type Response struct {
	Type  string  `json:"type"`
	Path  string  `json:"path"`
	Dest  *string `json:"dest,omitempty"`
	Error *string `json:"error,omitempty"`
}

func dispatchEvent(event fsnotify.Event, path string, message *babashka.Message) {
	response := Response{"", path, nil, nil}

	switch event.Op.String() {
	case "CHMOD":
		response.Type = "chmod"
	case "CREATE":
		response.Type = "create"
	case "REMOVE":
		response.Type = "remove"
	case "RENAME":
		response.Type = "rename"
		response.Dest = &event.Name
	case "WRITE":
		response.Type = "write"
	}

	babashka.WriteInvokeResponse(message, response)
}

func watch(message *babashka.Message, path string, opts Opts) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				dispatchEvent(event, path, message)
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

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func ProcessMessage(message *babashka.Message) {
	switch message.Op {
	case "describe":
		babashka.WriteDescribeResponse(&babashka.DescribeResponse{
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
					},
				},
			},
		})
	case "invoke":
		switch message.Var {
		case "pod.babashka.filewatcher/watch*":
			args := []json.RawMessage{}
			err := json.Unmarshal([]byte(message.Args), &args)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
			}

			opts := Opts{DelayMs: 2000}
			err = json.Unmarshal([]byte(args[1]), &opts)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
			}

			var path string
			json.Unmarshal(args[0], &path)

			watch(message, path, opts)
		default:
			babashka.WriteErrorResponse(message, fmt.Errorf("Unknown var %s", message.Var))
		}
	default:
		babashka.WriteErrorResponse(message, fmt.Errorf("Unknown op %s", message.Op))
	}
}
