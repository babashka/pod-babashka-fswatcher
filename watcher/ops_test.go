package watcher

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/babashka/pod-babashka-fswatcher/babashka"
)

func TestStartWatcher(t *testing.T) {
	message := babashka.Message{
		Op: "invoke", Id: "101", Var: "pod.babashka.fswatcher/-create-watcher", Args: "[101]"}

	startMessage := babashka.Message{
		Op: "invoke", Id: "2", Var: "pod.babashka.fswatcher/-start-watcher", Args: "[102]"}

	opts := Opts{DelayMs: 100, Recursive: false}
	watcherInfo, err := createWatcher(&message, "ops_test.go", opts)

	if err != nil {
		t.Errorf("Error starting watcher: %s", err)
	}

	idx := watcherInfo.WatcherId //starts at 1 and is incremented everyime create watcher is called

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

	time.Sleep(3 * time.Second)

	w.Close()
	out, _ := io.ReadAll(r)

	want := "Hello, World!\n"
	got := string(out)
	if got != want {
		t.Errorf("main() = %v, want %v", got, want)
	}

	//"d2:id1:26:statusl4:donee5:value37:{\"type\":\"chmod\",\"path\":\"ops_test.go\"}e"
	fmt.Println("out:", got)

}
