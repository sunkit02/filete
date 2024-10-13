package services

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

var DownloadServiceConfigs DownloadServiceConfig

func initializeService() string {
	tmpDirPath, err := os.MkdirTemp("", "filete_downloads_test-*")
	if err != nil {
		panic(err)
	}

	DownloadServiceConfigs = DownloadServiceConfig{SharedDirectories: []string{tmpDirPath}}

	InitDownloadService(DownloadServiceConfigs)

	return tmpDirPath
}

func TestReadDir(t *testing.T) {
	tmpDir := initializeService()

	err := os.MkdirAll(tmpDir+"/foo", 0700)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	rootDirHash := hashSHA256(tmpDir)

	expectedChildren := []SharedFile{
		{FType: Directory, Name: "foo", Path: "foo", Size: 0, RootDirHash: rootDirHash, Children: []SharedFile{}},
	}

	// Test function
	dir, err := readDir(tmpDir, rootDirHash, 2)
	if err != nil {
		t.Fatalf("Failed to readDir: %v", err)
	}

	fmt.Println(dir)

	tmpDirName := strings.Split(tmpDir, "/")[2]
	if dir.Name != tmpDirName {
		t.Fatalf("Expected %s. Got %s", tmpDirName, dir.Name)
	}

	if dir.Path != "" {
		t.Fatalf("Expected %s. Got %s", tmpDir, dir.Path)
	}

	if dir.FType != Directory {
		t.Fatalf("Expected %v. Got %v", Directory, dir.FType)
	}

	if dir.RootDirHash != rootDirHash {
		t.Fatalf("Expected %v. Got %v", rootDirHash, dir.RootDirHash)
	}

	if !reflect.DeepEqual(dir.Children, expectedChildren) {
		t.Fatalf("Expected:\n%+v.\nGot:\n%+v", expectedChildren, dir.Children)
	}
}
