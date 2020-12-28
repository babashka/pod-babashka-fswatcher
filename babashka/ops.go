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
	Name string `json:"name"`
	Vars []Var  `json:"vars"`
}

type Var struct {
	Name string `json:"name"`
	Code string `bencode:"code,omitempty"`
}

type DescribeResponse struct {
	Format     string      `json:"format"`
	Namespaces []Namespace `json:"namespaces"`
}

type InvokeResponse struct {
	Id     string   `json:"id"`
	Value  string   `json:"value"` // stringified json response
	Status []string `json:"status"`
}

type ErrorResponse struct {
	Id        string   `json:"id"`
	Status    []string `json:"status"`
	ExMessage string   `json:"ex-message"`
	ExData    string   `json:"ex-data"`
}

func ReadMessage() (*Message, error) {
	reader := bufio.NewReader(os.Stdin)
	message := &Message{}
	err := bencode.Unmarshal(reader, &message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func WriteDescribeResponse(describeResponse *DescribeResponse) {
	writeResponse(*describeResponse)
}

func WriteInvokeResponse(inputMessage *Message, value interface{}) error {
	resultValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	response := InvokeResponse{Id: inputMessage.Id, Status: []string{"done"}, Value: string(resultValue)}
	writeResponse(response)

	return nil
}

func WriteErrorResponse(inputMessage *Message, err error) {
	errorResponse := ErrorResponse{Id: inputMessage.Id, Status: []string{"done", "error"}, ExMessage: err.Error()}
	writeResponse(errorResponse)
}

func writeResponse(response interface{}) error {
	writer := bufio.NewWriter(os.Stdout)
	err := bencode.Marshal(writer, response)

	if err != nil {
		return err
	}

	writer.Flush()

	return nil
}
