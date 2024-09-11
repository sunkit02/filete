package data

import (
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Repository[T any] interface {
	Get(key uint) (T, error)
	GetAll() ([]T, error)
	Add(item T) (uint, error)
	AddAll(items []T) (uint, error)
	Delete(key uint) error
	DeleteAll() error
}

const MessageTombstoneStr = "\000tombstone\000"

var MessageTombstone = Message{
	Id:       0,
	Title:    MessageTombstoneStr,
	Body:     MessageTombstoneStr,
	TimeSent: time.Unix(0, 0),
}

type FileMessageRepo struct {
	file  *os.File
	cache []Message
	sep   string
}

func (repo *FileMessageRepo) Get(key uint) (Message, bool, error) {
	if repo.cache == nil {
		repo.populateCache()
	}

	if key >= uint(len(repo.cache)) {
		return Message{}, false, nil
	}

	if message := repo.cache[key]; message != MessageTombstone {
		return message, true, nil
	} else {
		return Message{}, false, nil
	}
}

func (repo *FileMessageRepo) GetAll() ([]Message, error) {
	if repo.cache == nil {
		repo.populateCache()
	}

	// Create new copy of data independent of cached values
	messages := make([]Message, len(repo.cache))
	for _, record := range repo.cache {
		if record != MessageTombstone {
			messages = append(messages, Message{
				Id:       record.Id,
				Title:    record.Title,
				Body:     record.Body,
				TimeSent: record.TimeSent,
			})
		}
	}
	return messages, nil
}

func (repo *FileMessageRepo) populateCache() {
	messages, err := parseMessagesFromFile(repo.file, repo.sep)
	if err != nil {
		log.Fatalf("error reading from message file: %v\n", err)
	}
	repo.cache = messages
}

func parseMessagesFromFile(file *os.File, sep string) ([]Message, error) {
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	messageStrings := strings.Split(string(bytes), "\n")
	messages := make([]Message, len(messageStrings))

	for i, messageString := range messageStrings {
		splits := strings.SplitN(messageString, sep, 3)
		id, err := strconv.ParseUint(splits[0], 10, 64)
		if err != nil {
			log.Println("Failed to parse id for entry", i)
			return nil, err
		}
		title := splits[1]
		body := splits[2]
		timeStamp, err := time.Parse("2006-01-02T15:04:05Z07:00", splits[3])
		if err != nil {
			log.Println("Failed to parse timestamp for entry", i)
			return nil, err
		}

		messages = append(messages, Message{id, title, body, timeStamp})
	}

	return messages, nil
}

type UploadedFilesRepo struct {
	basePath string
}
