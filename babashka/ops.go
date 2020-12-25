package babashka

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/jackpal/bencode-go"
)

type Message struct {
	Op   string
	Id   string
	Args string
	Var  string
}

type Namespace struct {
	Name string "name"
	Vars []Var  "vars"
}

type Var struct {
	Name string "name"
	Code string `bencode:"code,omitempty"`
}

type DescribeResponse struct {
	Format     string      "format"
	Namespaces []Namespace "namespaces"
}

type InvokeResponse struct {
	Id     string   "id"
	Value  string   "value" // stringified json response
	Status []string "status"
}

type ErrorResponse struct {
	Id        string   "id"
	Status    []string "status"
	ExMessage string   "ex-message"
	ExData    string   "ex-data"
}

func ReadMessage() *Message {
	reader := bufio.NewReader(os.Stdin)
	message := &Message{}
	err := bencode.Unmarshal(reader, &message)
	if err != nil {
		log.Fatalln("Could not decode bencode message", err)
	}
	log.Printf("Received Message: %+v\n", message)

	return message
}

func WriteDescribeResponse(describeResponse *DescribeResponse) {
	writeResponse(*describeResponse)
}

func WriteInvokeResponse(inputMessage *Message, value interface{}) {
	resultValue, err := json.Marshal(value)
	if err != nil {
		log.Fatalln("Could not marshall value to json", err)
	}
	response := InvokeResponse{Id: inputMessage.Id, Status: []string{"done"}, Value: string(resultValue)}
	writeResponse(response)
}

func WriteErrorResponse(inputMessage *Message, err error) {
	errorResponse := ErrorResponse{Id: inputMessage.Id, Status: []string{"done", "error"}, ExMessage: err.Error()}
	writeResponse(errorResponse)
}

func writeResponse(response interface{}) {
	log.Printf("Writing response: %+v\n", response)
	writer := bufio.NewWriter(os.Stdout)
	err := bencode.Marshal(writer, response)

	if err != nil {
		log.Fatalln("Couldn't write response", err)
	}

	writer.Flush()
}
