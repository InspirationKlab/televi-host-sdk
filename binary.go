package televi_host_sdk

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Contents []byte

//goland:noinspection GoMixedReceiverTypes
func (contents *Contents) FromBinary(reader *bytes.Reader) error {
	length, err := ReadUint64(reader)
	if err != nil {
		return err
	}
	*contents = make(Contents, length)
	_, err = reader.Read(*contents)
	return err
}

func ReadUint64(b *bytes.Reader) (uint64, error) {
	buff := [8]byte{}
	_, err := b.Read(buff[:])
	if err != nil {
		return 0, err
	}
	value := binary.LittleEndian.Uint64(buff[:])
	return value, nil
}

//goland:noinspection GoMixedReceiverTypes
func (contents Contents) ToBinary(writer *bytes.Buffer) error {
	writer.Write(LenToBytes(contents))
	writer.Write(contents)
	return nil
}

func LenToBytes[T any](source []T) []byte {
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffer, uint64(len(source)))
	return buffer
}

type ContentsSection = []Contents

func SerializeSlice[T BinarySerializable](slice []T, writer *bytes.Buffer) error {
	writer.Write(LenToBytes(slice))
	for _, entry := range slice {
		err := entry.ToBinary(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeserializeSlice[T any](reader *bytes.Reader, slice *[]T) error {
	length, err := ReadUint64(reader)
	if err != nil {
		return err
	}
	*slice = make([]T, length)
	for i := 0; i < int(length); i++ {
		ptrToElement := (any)(&((*slice)[i]))
		binDeser, isBinaryDeserializable := ptrToElement.(BinaryDeserializable)
		if !isBinaryDeserializable {
			var t T
			return fmt.Errorf("type %T is not binary deserializable", t)
		}
		err = binDeser.FromBinary(reader)

		if err != nil {
			return err
		}
	}
	return nil
}

type BinarySerializable interface {
	ToBinary(writer *bytes.Buffer) error
}

type BinaryDeserializable interface {
	FromBinary(reader *bytes.Reader) error
}
