package main

import (
	"github.com/babashka/pod-fswatcher/babashka"
	"github.com/babashka/pod-fswatcher/watcher"
)

func main() {
	for {
		message := babashka.ReadMessage()
		watcher.ProcessMessage(message)
	}
}
