package watcher

import (
	//"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackpal/bencode-go"

	"github.com/babashka/pod-babashka-fswatcher/babashka"
)

type FileAction struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

func TestStartWatcher(t *testing.T) {
	message := babashka.Message{
		Op: "invoke", Id: "101", Var: "pod.babashka.fswatcher/-create-watcher", Args: "[101]"}

	startMessage := babashka.Message{
		Op: "invoke", Id: "2", Var: "pod.babashka.fswatcher/-start-watcher", Args: "[102]"}

	opts := Opts{DelayMs: 50, Recursive: true}
	watcherInfo, err := createWatcher(&message, ".", opts)

	if err != nil {
		t.Errorf("Error starting watcher: %s", err)
	}

	idx := watcherInfo.WatcherId

	defer func(orig *os.File) {
		os.Stdout = orig
	}(os.Stdout)

	r, w, _ := os.Pipe()
	os.Stdout = w
	e := startWatcher(&startMessage, idx)
	if e != nil {
		fmt.Printf("Error starting watcher: %s", e)
	}

	//touch this file
	err = os.Chtimes("ops_test.go", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to touch file: %s", err)
	}

	time.Sleep(100 * time.Millisecond)

	err = os.Chtimes("ops_test.go", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to touch file: %s", err)
	}

	time.Sleep(100 * time.Millisecond)

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)
	fmt.Println("out:", got)
	messages := strings.Split(got, "}e")

	// Process each bencode message
	for _, message := range messages {
		if len(message) == 0 {
			continue
		}

		reader := strings.NewReader(message + "}e")

		// Parse the bencode message into an InvokeResponse struct
		var response babashka.InvokeResponse
		err = bencode.Unmarshal(reader, &response)
		if err != nil {
			fmt.Printf("Error parsing bencode message: %v\n", err)
			continue
		}

		// Process the parsed response as needed
		fmt.Println("Id:", response)
	}

	//want := "Hello, World!\n"
	dataFromFsnotify := babashka.InvokeResponse{}
	err = bencode.Unmarshal(r, &dataFromFsnotify)
	if err != nil {
		t.Fatalf("Failed to touch file: %s", err)
	}

	//data := whatsthis.(map[string]interface{})["value"]
	//dataString := data.(string)

	//var action FileAction

	// Parse the JSON string into the FileAction struct
	//err = json.Unmarshal([]byte(dataString), &action)
	//if err != nil {
	//		fmt.Println("Error parsing JSON:", err)
	//		return
	//	}

	//	path := action.Path
	//msg := json.Decoder()

	//if got != want {
	//	t.Errorf("main() = %v, want %v", got, want)
	//	}

	//"d2:id1:26:statusl4:donee5:value37:{\"type\":\"chmod\",\"path\":\"ops_test.go\"}e"
	fmt.Println("out:", dataFromFsnotify)

}
