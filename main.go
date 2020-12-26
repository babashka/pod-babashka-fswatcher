package main

import (
	"github.com/babashka/pod-fswatcher/babashka"
	"github.com/babashka/pod-fswatcher/watcher"
)

func main() {
	for {
		message, err := babashka.ReadMessage()
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			continue
		}

		res, err := watcher.ProcessMessage(message)
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			continue
		}

		describeRes, ok := res.(*babashka.DescribeResponse)
		if ok {
			babashka.WriteDescribeResponse(describeRes)
			continue
		}

		babashka.WriteInvokeResponse(message, res)
	}
}
