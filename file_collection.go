package televi_host_sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileCollection struct {
	Root     Info
	Contents ContentsSection
}

func (fileCollection *FileCollection) FromBinary(reader *bytes.Reader) error {
	preambleLength, err := ReadUint64(reader)
	if err != nil {
		return err
	}
	jsonPart := make([]byte, preambleLength)
	err = json.Unmarshal(jsonPart, &fileCollection.Root)
	return DeserializeSlice(reader, &fileCollection.Contents)
}

func (fileCollection *FileCollection) ToBinary(writer *bytes.Buffer) error {
	jsonPart, err := json.Marshal(fileCollection.Root)
	if err != nil {
		return err
	}
	writer.Write(LenToBytes(jsonPart))
	writer.Write(jsonPart)
	return SerializeSlice(fileCollection.Contents, writer)
}

type Info struct {
	Entries    map[string]*Info
	IsFolder   bool
	EntryIndex int
	ModifiedAt time.Time
	IsDeleted  bool
}

func traverseFolder(path string, previous *Info, content *[]Contents) (Info, error) {
	file, err := os.Stat(path)
	info := Info{
		IsFolder:   file.IsDir(),
		ModifiedAt: file.ModTime(),
	}
	if err != nil {
		// check if previous entry exists and was not deleted
		if previous != nil {
			info.IsDeleted = true
		}
		return info, nil
	}

	if !file.IsDir() {
		if !previous.ModifiedAt.IsZero() && info.ModifiedAt.Before(previous.ModifiedAt) {
			return info, nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return Info{}, err
		}
		*content = append(*content, data)
		info.EntryIndex = len(*content) - 1
		return info, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return Info{}, err
	}
	info.Entries = make(map[string]*Info, len(entries))
	for _, entry := range entries {
		subPath := filepath.Join(path, entry.Name())
		subPrevious, ok := previous.Entries[entry.Name()]
		subInfo, err := traverseFolder(subPath, subPrevious, content)
		if err != nil {
			return Info{}, err
		}
		if ok && subInfo.ModifiedAt.Before(subPrevious.ModifiedAt) {
			subInfo.ModifiedAt = subPrevious.ModifiedAt
		}
		info.Entries[entry.Name()] = &subInfo
	}
	// check for deleted entries
	for name, entry := range previous.Entries {
		if _, ok := info.Entries[name]; !ok {
			entry.IsDeleted = true
			info.Entries[name] = entry
		}
	}
	return info, nil
}

func WrapFolder(path string, previous Info) (FileCollection, error) {
	fileCollection := FileCollection{
		Root:     Info{},
		Contents: make([]Contents, 0),
	}
	var err error
	fileCollection.Root, err = traverseFolder(path, &previous, &fileCollection.Contents)
	return fileCollection, err
}

func UnpackFolder(path string, info *Info, content ContentsSection) error {
	if info.IsFolder {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	for name, subInfo := range info.Entries {
		subPath := filepath.Join(path, name)
		if err := UnpackFolder(subPath, subInfo, content); err != nil {
			return err
		}
	}
	if !info.IsFolder {
		if len(content) <= info.EntryIndex {
			return fmt.Errorf("content index out of range")
		}
		data := content[info.EntryIndex]
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	return nil
}
