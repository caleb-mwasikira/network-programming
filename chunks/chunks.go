package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

var (
	Bytes     uint32 = 8    // 1 byte -> 8 bits
	Kilobytes uint32 = 1024 // 1 kb -> 1024 bytes
	Megabytes uint32 = 1024 * 1024
	Gigabytes uint32 = 1024 * 1024 * 1024
)

type Chunk struct {
	fpath       string
	cursor_pos  uint64
	buffer_size uint64
	md5sum      string
}

func (chunk Chunk) SaveToFile(buffer []byte) error {
	fmt.Printf("saving chunk to file %v...\n", path.Base(chunk.fpath))

	if _, err := os.Stat(chunk.fpath); err == nil {
		// filepath exists
		fmt.Printf("chunk file %v already exists\n", path.Base(chunk.fpath))
		return nil
	}

	file, err := os.OpenFile(chunk.fpath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := file.Write(buffer)
	if err != nil {
		return err
	}

	if n != len(buffer) {
		return fmt.Errorf("only wrote %v bytes out of %v bytes", n, len(buffer))
	}

	return nil
}

func (chunk Chunk) LoadFromFile() ([]byte, error) {
	fmt.Printf("loading buffer from chunk file %v...\n", chunk.fpath)
	buffer, err := os.ReadFile(chunk.fpath)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func ChunkFile(fpath string) ([]Chunk, error) {
	fmt.Printf("chunking file %v...\n", fpath)
	file, err := os.OpenFile(fpath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	chunks := []Chunk{}
	chunk_size := 69 * Megabytes
	cursor_pos := 0
	for {
		buffer := make([]byte, chunk_size)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read chunk; %v", err)
		}
		if n == 0 {
			// reached end of file
			break
		}
		buffer = buffer[:n]

		hash := md5.Sum(buffer)
		md5_hash := hex.EncodeToString(hash[:])
		parent_fname := path.Base(fpath)
		chunk_fname := fmt.Sprintf("%v_%v_%v", md5_hash, cursor_pos, parent_fname)
		chunk_fpath := path.Join(path.Dir(fpath), chunk_fname)

		chunk := Chunk{
			fpath:       chunk_fpath,
			cursor_pos:  uint64(cursor_pos),
			buffer_size: uint64(len(buffer)),
			md5sum:      md5_hash,
		}
		err = chunk.SaveToFile(buffer)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
		cursor_pos += n
	}

	return chunks, nil
}

func main() {
	fpath := "test_data/file.txt"
	chunks, err := ChunkFile(fpath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("saved %v chunks to disk\n", len(chunks))
}
