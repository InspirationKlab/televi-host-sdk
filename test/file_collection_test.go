package test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"televi_host_sdk"
	"testing"
)

func TestTraverseFolderAndUnpackFolder(t *testing.T) {
	// Create a temporary directory and some files for testing
	tempDir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	file1 := filepath.Join(tempDir, "file1.txt")
	data1 := []byte("hello world")
	if err := ioutil.WriteFile(file1, data1, 0644); err != nil {
		t.Fatal(err)
	}
	file2 := filepath.Join(tempDir, "file2.txt")
	data2 := []byte("goodbye")
	if err := ioutil.WriteFile(file2, data2, 0644); err != nil {
		t.Fatal(err)
	}
	// Test the traverseFolder and UnpackFolder functions
	info, err := televi_host_sdk.WrapFolder(tempDir, televi_host_sdk.Info{})
	if err != nil {
		t.Fatal(err)
	}
	if err := televi_host_sdk.UnpackFolder(filepath.Join(tempDir, "newdir"), info.Root, info.Contents); err != nil {
		t.Fatal(err)
	}
	// Check that the unpacked files match the original files
	newFile1 := filepath.Join(tempDir, "newdir", "file1.txt")
	newData1, err := ioutil.ReadFile(newFile1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data1, newData1) {
		t.Error("data1 does not match")
	}
	newFile2 := filepath.Join(tempDir, "newdir", "file2.txt")
	newData2, err := ioutil.ReadFile(newFile2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data2, newData2) {
		t.Error("data2 does not match")
	}
}
