package data

import (
	"bytes"
	"encoding/json"
	"floader/logging"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

const testFilePath = "/tmp/floader_messages"

func init() {
	logging.InitializeLoggers(os.Stdout)
}

func initNewFileRepo() FileMessageRepo {
	dataFilePath := fmt.Sprintf("%s-%d.dat", testFilePath, rand.Uint64())

	return NewFileMessageRepo(dataFilePath)
}

func TestAdd(t *testing.T) {
	repo := initNewFileRepo()

	message1 := Message{Id: 1, Title: "Hello", Body: "World", TimeSent: time.Now().UTC()}
	message2 := Message{Id: 2, Title: "Foo", Body: "Bar", TimeSent: time.Now().UTC()}

	encodedMsg1, err := json.Marshal(message1)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	encodedMsg2, err := json.Marshal(message2)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	err = repo.AddAll([]Message{message1, message2})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	// Check if file content is as expected
	bytesRead, _ := os.ReadFile(repo.path)
	var buffer bytes.Buffer
	buffer.Write(encodedMsg1)
	buffer.Write([]byte{'\n'})
	buffer.Write(encodedMsg2)
	buffer.Write([]byte{'\n'})
	expectedBytes := buffer.Bytes()

	if !bytes.Equal(expectedBytes, bytesRead) {
		t.Logf(
			"Expected:\n%v\nGot:\n%v\n",
			string(expectedBytes),
			string(bytesRead),
		)
		t.FailNow()
	}

}

func TestGet(t *testing.T) {
	repo := initNewFileRepo()

	expectedMsg1 := Message{Id: 1, Title: "Hello", Body: "World", TimeSent: time.Now().UTC()}
	expectedMsg2 := Message{Id: 2, Title: "Foo", Body: "Bar", TimeSent: time.Now().UTC()}

	err := repo.AddAll([]Message{expectedMsg1, expectedMsg2})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	for _, expected := range []Message{expectedMsg1, expectedMsg2} {
		message, exists, err := repo.Get(expected.Id)
		if exists != true {
			t.Logf("Failed to find message with expected id: %d\n", expected.Id)
			t.FailNow()
		}
		if err != nil {
			t.Log("Error:", err)
			t.FailNow()
		}
		if message != expected {
			t.Logf("Assertion Error:\nExpected:\n%v\nGot:\n%v", expected, message)
			t.FailNow()
		}
	}
}

func TestGetAll(t *testing.T) {
	repo := initNewFileRepo()

	expectedMessages := []Message{
		{Id: 1, Title: "Hello", Body: "World", TimeSent: time.Now().UTC()},
		{Id: 2, Title: "Foo", Body: "Bar", TimeSent: time.Now().UTC()},
	}

	err := repo.AddAll(expectedMessages)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	messages, err := repo.GetAll()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if !slicesEqual(messages, expectedMessages) {
		t.Logf("Assertion Error:\nExpected:\n%v\nGot:\n%v", expectedMessages, messages)
		t.FailNow()
	}
}

func TestDelete(t *testing.T) {
	repo := initNewFileRepo()

	expectedMessages := []Message{
		{Title: "Hello", Body: "World", TimeSent: time.Now().UTC()},
		{Title: "Foo", Body: "Bar", TimeSent: time.Now().UTC()},
	}

	err := repo.AddAll(expectedMessages)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	messages1, err := repo.GetAll()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	repo.Delete(messages1[0].Id)

	messages2, err := repo.GetAll()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if len(messages2) != 1 {
		t.Log("Expected length to be 1, Got:", len(messages2))
		t.FailNow()
	}

	if messages2[0] != messages1[1] {
		t.Logf("Expected:\n%+v\nGot:\n%+v\n", messages1[1], messages2[0])
		t.FailNow()
	}

	repo.Delete(messages2[0].Id)

	messages3, err := repo.GetAll()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if len(messages3) != 0 {
		t.Log("Expected length to be 0, Got:", len(messages3))
		t.FailNow()
	}
}

func TestDeleteAll(t *testing.T) {
	repo := initNewFileRepo()

	expectedMessages := []Message{
		{Title: "Hello", Body: "World", TimeSent: time.Now().UTC()},
		{Title: "Foo", Body: "Bar", TimeSent: time.Now().UTC()},
	}

	err := repo.AddAll(expectedMessages)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	repo.DeleteAll()

	messages, err := repo.GetAll()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if len(messages) != 0 {
		t.Log("Expected length to be 0, Got:", len(messages))
		t.FailNow()
	}
}

func slicesEqual[t comparable](s1, s2 []t) bool {
	if len(s1) != len(s2) {
		return false
	}

	m1 := make(map[t]bool)
	m2 := make(map[t]bool)

	mapsEqual(m1, m2)

	return true
}

// Checks for equality between two maps
func mapsEqual[K comparable, V comparable](m1, m2 map[K]V) bool {
	if len(m1) != len(m2) {
		return false
	}

	for key, value := range m1 {
		if v, ok := m2[key]; !ok || v != value {
			return false
		}
	}

	return true
}
