package services

import (
	"errors"
	"os"
)

var sharedDirectories []string

type DownloadServiceConfig struct {
	SharedDirectories []string
}

const (
	File = iota
	Directory
)

type SharedFile struct {
	FType int
	Name  string
	Size  int64

	// This is not nil only if Ftype == Directory
	Children []SharedFile
}

func InitDownloadService(c DownloadServiceConfig) {
	sharedDirectories = c.SharedDirectories
}

func GetSharedFiles() ([]SharedFile, error) {
	sharedDirs := make([]SharedFile, 0, len(sharedDirectories))
	for _, dir := range sharedDirectories {
		dir, err := readDir(dir)
		if err != nil {
			return nil, err
		}
		sharedDirs = append(sharedDirs, dir)
	}

	return sharedDirs, nil
}

func readDir(path string) (SharedFile, error) {
	stat, err := os.Stat(path)
	if !stat.IsDir() {
		return SharedFile{}, errors.New(path + " is not a directory")
	}
	dirName := stat.Name()
	dirSize := stat.Size()

	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return SharedFile{}, err
	}

	children := make([]SharedFile, 0)
	for _, entry := range dirEntries {
		info, err := entry.Info()
		if err != nil {
			return SharedFile{}, err
		}
		name := info.Name()

		if info.IsDir() {
			child, err := readDir(path + "/" + name)
			children = append(children, child)
			if err != nil {
				return SharedFile{}, nil
			}
		} else {
			children = append(children, SharedFile{
				FType:    File,
				Name:     name,
				Size:     info.Size(),
				Children: nil,
			})
		}
	}

	sharedDir := SharedFile{
		FType:    Directory,
		Name:     dirName,
		Size:     dirSize,
		Children: children,
	}
	return sharedDir, nil
}
