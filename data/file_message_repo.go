package data

import (
	"bytes"
	"encoding/json"
	"github.com/sunkit02/filete/logging"
	"fmt"
	"io"
	"log"
	"os"
)

type FileMessageRepo struct {
	path             string
	file             *os.File
	nextId           MessageId
	messageReadCache map[MessageId]Message
}

func NewFileMessageRepo(path string) FileMessageRepo {
	return FileMessageRepo{path: path}
}

func (repo *FileMessageRepo) Get(id MessageId) (Message, bool, error) {
	if repo.messageReadCache == nil {
		repo.populateCache()
	}

	if message, exists := repo.messageReadCache[id]; exists {
		return message, true, nil
	} else {
		return Message{}, false, nil
	}
}

func (repo *FileMessageRepo) GetAll() ([]Message, error) {
	if repo.messageReadCache == nil {
		repo.populateCache()
	}

	// Create new copy of data independent of cached values
	messages := make([]Message, 0, len(repo.messageReadCache))
	for _, message := range repo.messageReadCache {
		messages = append(messages, Message{
			Id:       message.Id,
			Body:     message.Body,
			TimeSent: message.TimeSent,
		})
	}
	return messages, nil
}

// Returns DuplicateEntryError if there is a duplicate key entry exists
// NOTE: This repo automatically overrides the Id field of Message.
func (repo *FileMessageRepo) Add(message Message) error {
	// TODO: Extract this out and have a safe Add func that checks this and an
	// internal unsafe add that gets called by the safe Add after the check
	if repo.messageReadCache == nil {
		repo.populateCache()
	}

	repo.nextId++
	message.Id = repo.nextId
	err := repo.appendToFile(message)
	if err != nil {
		return nil
	}

	repo.messageReadCache[message.Id] = message

	return nil
}

// Returns DuplicateEntryError if there is a duplicate key entry exists.
// This method is atomic and aborts if there is a single duplicate entry
// NOTE: This repo automatically overrides the Id field of Message.
func (repo *FileMessageRepo) AddAll(messages []Message) error {
	for _, message := range messages {
		repo.Add(message)
	}

	return nil
}

func (repo *FileMessageRepo) Delete(id MessageId) {
	if repo.messageReadCache == nil {
		repo.populateCache()
	}

	if _, exists := repo.messageReadCache[id]; exists {
		delete(repo.messageReadCache, id)
		repo.file.Close()
		_ = os.Truncate(repo.path, 0)

		file, err := os.OpenFile(repo.path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			log.Fatal("Failed to open file: ", err)
		}
		repo.file = file

		for _, message := range repo.messageReadCache {
			err := repo.appendToFile(message)
			if err != nil {
				logging.Error.Fatal("Failed to write back to file after deletion, possible data corruption.")
			}
		}

	}
}

func (repo *FileMessageRepo) DeleteAll() {
	if repo.messageReadCache != nil {
		repo.messageReadCache = nil
		repo.file.Close()
	}

	os.Truncate(repo.path, 0)
}

func (repo *FileMessageRepo) populateCache() {
	if repo.file == nil {
		file, err := os.OpenFile(repo.path, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logging.Error.Fatal("Failed to open file: ", err)
		}
		repo.file = file
	}

	messages, err := parseMessagesFromFile(repo.file)
	if err != nil {
		logging.Error.Printf("error reading from message file: %v\n", err)
	}
	messageMap := make(map[MessageId]Message, len(messages))
	for _, message := range messages {
		messageMap[message.Id] = message
		if message.Id > repo.nextId {
			repo.nextId = message.Id
		}
	}
	repo.messageReadCache = messageMap
}

func (repo *FileMessageRepo) appendToFile(message Message) error {
	if repo.file == nil {
		file, err := os.OpenFile(repo.path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logging.Error.Fatal("Failed to open file: ", err)
		}
		repo.file = file
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		logging.Error.Fatal("Failed to marshal Message:", err)
	}

	_, err = repo.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = repo.file.Write(bytes)

	if err != nil {
		return err
	}

	_, err = repo.file.WriteString("\n")
	if err != nil {
		return err
	}

	return nil
}

func parseMessagesFromFile(file *os.File) ([]Message, error) {
	bytesRead, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	messageBytesList := bytes.Split(bytesRead, []byte{'\n'})
	messages := make([]Message, 0, len(messageBytesList))

	for i, messageBytes := range messageBytesList {
		if len(messageBytes) == 0 {
			continue
		}
		var message Message
		err := json.Unmarshal(messageBytes, &message)
		if err != nil {
			logging.Error.Printf("Failed to unmarshal Message. Line %d: '%s'", i+1, string(messageBytes))
			continue
		}
		messages = append(messages, message)
	}

	return messages, nil
}

type UploadedFilesRepo struct {
	basePath string
}

type DuplicateEntryError[K any] struct {
	duplicateKey K
}

func (e *DuplicateEntryError[K]) Error() string {
	return fmt.Sprintf("Duplicate entry error: duplicate key '%v'", e.duplicateKey)
}
