package babashka

import (
	"bufio"
	"encoding/json"
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

func ReadMessage() (*Message, error) {
	reader := bufio.NewReader(os.Stdin)
	message := &Message{}
	if err := bencode.Unmarshal(reader, &message); err != nil {
		return nil, err
	}

	return message, nil
}

func WriteDescribeResponse(describeResponse *DescribeResponse) {
	writeResponse(*describeResponse)
}

func WriteInvokeResponse(inputMessage *Message, value any, done bool) error {
	if value == nil {
		return nil
	}
	resultValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var status []string
	if (done) {
		status = []string{"done"}
	} else {
		status = []string{}
	}
	response := InvokeResponse{Id: inputMessage.Id, Status: status, Value: string(resultValue)}
	writeResponse(response)

	return nil
}

func WriteErrorResponse(inputMessage *Message, err error) {
	errorResponse := ErrorResponse{Id: inputMessage.Id, Status: []string{"done", "error"}, ExMessage: err.Error()}
	writeResponse(errorResponse)
}

func writeResponse(response any) error {
	writer := bufio.NewWriter(os.Stdout)
	if err := bencode.Marshal(writer, response); err != nil {
		return err
	}

	writer.Flush()

	return nil
}
