package services

import (
	"crypto/sha256"
	"errors"
	"filete/logging"
	"fmt"
	"os"
	"strings"
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
	FType       int    `json:"fType"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	RootDirHash string `json:"rootDirHash"`

	// This is not nil only if FType == Directory, but it still can be nil even
	// if FType == Directory when the contents has yet to be fetched
	Children []SharedFile `json:"children"`
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
	logging.Debug.Println(path, "Root hash: "+rootDirHash, depth)
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

	rootDirPath := sharedRootDirs[rootDirHash].Path

	stat, err := os.Stat(path)
	if !stat.IsDir() {
		return SharedFile{}, errors.New(path + " is not a directory")
	}
	dirName := stat.Name()

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

		// set file type to `File` initially
		childFType := File
		childName := info.Name()
		childPath := path + "/" + childName
		childSize := info.Size()
		var childChildren []SharedFile

		if info.IsDir() {
			if depth > 1 {
				child, err := readDir(childPath, rootDirHash, depth-1)
				if err != nil {
					return SharedFile{}, err
				}
				children = append(children, child)
			} else {
				childFType = Directory
			}
		} else {
			children = append(children, SharedFile{
				FType:       childFType,
				Name:        childName,
				Path:        stripRootPath(childPath, rootDirPath),
				Size:        childSize,
				RootDirHash: rootDirHash,
				Children:    childChildren,
			})
		}
	}

	var dirSize int64 = 0

	for _, child := range children {
		dirSize += child.Size
	}

	sharedDir := SharedFile{
		FType:       Directory,
		Name:        dirName,
		Path:        stripRootPath(path, rootDirPath),
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

// Strips the given root path from a path and returns the new relative path.
// Does nothing if the root path is not part of the path.
func stripRootPath(path string, rootPath string) string {
	if !strings.HasPrefix(path, rootPath) {
		return path
	}

	path = strings.Split(path, rootPath)[1]
	// Strip leading '/' after stripping root path
	if strings.HasPrefix(path, "/") {
		path = strings.SplitN(path, "/", 2)[1]
	}

	return path
}
