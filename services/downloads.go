package services

import (
	"archive/zip"
	"crypto/sha256"
	"errors"
	"filete/logging"
	"filete/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
	logging.Debug.Println("ReadDir Path: "+path, "Root hash: "+rootDirHash, "depth:", depth)
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

// Returns the bytes of a file, file name, and an error if there is any.
// If the file is a directory it will be returned in the form of a zip file.
func GetFileForDownload(path, rootDirHash string) (io.Reader, string, bool, error) {
	logging.Debug.Printf("GetFileBytes(%v, %v)", path, rootDirHash)

	rootDir, ok := sharedRootDirs[rootDirHash]
	if !ok {
		return nil, "", false, errors.New("Invalid rootDirHash")
	}

	fullPath := rootDir.Path
	// Only add '/' separator when path is not empty.
	// Unnecessary currently but for good measure.
	if path != "" {
		fullPath += "/" + path
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, "", false, err
	}

	if info.IsDir() {
		tmpFilePath := "/tmp/filete-downloads-" + utils.GenerateRandomString(10) + ".zip"
		zipFile, err := os.Create(tmpFilePath)
		if err != nil {
			return nil, "", false, err
		}
		ZipDirectory(fullPath, zipFile)
		zipFile.Close()

		zipFile, err = os.Open(tmpFilePath)
		if err != nil {
			return nil, "", false, err
		}

		return zipFile, info.Name(), true, nil
	} else {
		file, err := os.Open(fullPath)
		if err != nil {
			return nil, "", false, err
		}

		return file, info.Name(), false, nil
	}
}

func readDir(path, rootDirHash string, depth int) (SharedFile, error) {
	logging.Debug.Println("readDir Path: "+path, "Root hash: "+rootDirHash, "depth:", depth)
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
		logging.Trace.Println("Found entry ", entry.Name(), " isDir:", entry.IsDir())

		info, err := entry.Info()
		if err != nil {
			logging.Debug.Println("Failed to get entry info for", entry.Name(), err)
			return SharedFile{}, err
		}

		// set file type to `File` initially
		childName := info.Name()
		childPath := path + "/" + childName
		childSize := info.Size()
		var childChildren []SharedFile

		if info.IsDir() && depth > 1 {
			child, err := readDir(childPath, rootDirHash, depth-1)
			if err != nil {
				return SharedFile{}, err
			}
			children = append(children, child)
		} else {
			childFType := File
			if info.IsDir() {
				childFType = Directory
			}
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

	if children != nil {
		sort.Sort(AsFiles(children))
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

// TODO: Understand how zip and compression works at a deeper level
func ZipDirectory(source string, destination io.Writer) error {
	// Create a new zip writer
	zipWriter := zip.NewWriter(destination)
	defer zipWriter.Close()

	// Walk through the source directory
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the zip file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set the correct path in the zip archive (relative to the source directory)
		header.Name = strings.TrimPrefix(path, filepath.Dir(source)+"/")

		// If it's a directory, we need to mark it as a directory in the header
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate // Use Deflate compression method for files
		}

		// Create the writer in the zip archive
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// If it's a file, copy the file data to the zip archive
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
