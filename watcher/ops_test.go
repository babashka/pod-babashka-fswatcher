package watcher

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/babashka/pod-babashka-fswatcher/babashka"
	"github.com/jackpal/bencode-go"
)

func TestStartWatcher(t *testing.T) {

	watchFolder := "."
	thisFile := "ops_test.go"
	delay := 50
	recursive := true

	createMessage := babashka.Message{
		Op: "invoke", Id: "101", Var: "pod.babashka.fswatcher/-create-watcher", Args: "[101]"}

	startMessage := babashka.Message{
		Op: "invoke", Id: "2", Var: "pod.babashka.fswatcher/-start-watcher", Args: "[102]"}

	opts := Opts{DelayMs: uint64(delay), Recursive: recursive, Dedup: true}

	watcherInfo, err := createWatcher(&createMessage, watchFolder, opts)

	if err != nil {
		t.Errorf("Error starting watcher: %s", err)
	}

	idx := watcherInfo.WatcherId

	if e := startWatcher(&startMessage, idx); e != nil {
		fmt.Printf("Error starting watcher: %s", e)
	}

	//touch this file and capture babashka output
	fsNotifications, err := captureBabashkaOutput(func() error {

		// trying to recreate test/script.clj test
		time.Sleep(100 * time.Millisecond)

		if ee := os.Chtimes(thisFile, time.Now(), time.Now()); ee != nil {
			return ee
		}

		//events within delay should be ignored.
		time.Sleep(49 * time.Millisecond)

		if ee := os.Chtimes(thisFile, time.Now(), time.Now()); ee != nil {
			return ee
		}

		time.Sleep(51 * time.Millisecond)

		return nil
	})

	if err != nil {
		t.Errorf("Error Capturing output: %s", err)
	}

	fmt.Println(fsNotifications)

	if len(fsNotifications) != 1 {
		t.Errorf("Expected 1 notification, but got %d", len(fsNotifications))
	}

	if fsNotifications[0].Path != "./ops_test.go" {
		t.Errorf("Expected notification Path to be './ops_test.go', but got %s", fsNotifications[0].Path)
	}

}

func captureBabashkaOutput(f func() error) ([]Response, error) {

	defer func(orig *os.File) {
		os.Stdout = orig
	}(os.Stdout)

	r, w, _ := os.Pipe()
	os.Stdout = w
	if err := f(); err != nil {
		fmt.Print("Failed to touch file and capture output: ", err)
	}
	w.Close()
	out, _ := io.ReadAll(r)

	outString := string(out)

	bencodeStrings := strings.Split(outString, "}e")

	//return bencodeStrings

	responses := []Response{}

	// Process each bencode message
	for _, bencodeString := range bencodeStrings {

		if len(bencodeString) == 0 {
			continue
		}

		reader := strings.NewReader(bencodeString + "}e")

		// Parse the bencode message into an InvokeResponse struct
		var babashkaMessage babashka.InvokeResponse

		if err := bencode.Unmarshal(reader, &babashkaMessage); err != nil {
			return nil, err
		}

		jsonString := babashkaMessage.Value

		var fsnotifyMsg Response

		if err := json.Unmarshal([]byte(jsonString), &fsnotifyMsg); err != nil {
			return nil, err
		}

		responses = append(responses, fsnotifyMsg)
	}

	return responses, nil
}
