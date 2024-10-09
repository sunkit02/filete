package services

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
)

const (
	File = iota
	Directory
)

var sharedRootDirs map[string]SharedRootDir

type SharedRootDir struct {
	Id   string
	Path string
}

type SharedFile struct {
	FType       int
	Name        string
	Path        string
	Size        int64
	RootDirHash string

	// This is not nil only if Ftype == Directory
	Children []SharedFile
}

type DownloadServiceConfig struct {
	SharedDirectories []string
}

func InitDownloadService(c DownloadServiceConfig) {
	sharedRootDirs = make(map[string]SharedRootDir)
	for _, path := range c.SharedDirectories {
		id := hashSHA256(path)
		sharedRootDirs[id] = SharedRootDir{
			Id:   id,
			Path: path,
		}
	}
}

func ReadRootDirs(depth int) ([]SharedFile, error) {
	sharedDirs := make([]SharedFile, 0, len(sharedRootDirs))
	for hash := range sharedRootDirs {
		dir, err := ReadDir("", hash, depth)
		if err != nil {
			return nil, err
		}

		sharedDirs = append(sharedDirs, dir)
	}

	return sharedDirs, nil
}

func ReadDir(path, rootDirHash string, depth int) (SharedFile, error) {
	rootDir, ok := sharedRootDirs[rootDirHash]
	if !ok {
		return SharedFile{}, errors.New("Invalid rootDirHash")
	}
	divider := "/"
	// To avoid double '/' when path is empty (reading a Root Directory)
	if path == "" {
		divider = ""
	}
	path = rootDir.Path + divider + path

	return readDir(path, rootDirHash, depth)
}

func readDir(path, rootDirHash string, depth int) (SharedFile, error) {
	if depth < 1 {
		return SharedFile{}, fmt.Errorf("Depth must be >= 1. Got %d", depth)
	}

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
			childPath := path + "/" + name
			child, err := readDir(childPath, rootDirHash, depth-1)
			children = append(children, child)
			if err != nil {
				return SharedFile{}, nil
			}
		} else {
			children = append(children, SharedFile{
				FType:       File,
				Name:        name,
				Path:        path + "/" + name,
				Size:        info.Size(),
				RootDirHash: rootDirHash,
				Children:    nil,
			})
		}
	}

	sharedDir := SharedFile{
		FType:       Directory,
		Name:        dirName,
		Path:        path,
		Size:        dirSize,
		RootDirHash: rootDirHash,
		Children:    children,
	}
	return sharedDir, nil
}

func hashSHA256(s string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
