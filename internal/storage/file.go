package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

const maxCapacity = 1024

type FileStorage struct {
	file *os.File
}

type Row struct {
	Key   string
	Value string
}

func NewFileStorage(filename string) *FileStorage {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}

	return &FileStorage{file: file}
}

func (s *FileStorage) Close() error {
	return s.file.Close()
}

func (s *FileStorage) Load() (map[string]string, error) {
	scanner := bufio.NewScanner(s.file)

	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	data := make(map[string]string)

	for scanner.Scan() {
		rawRow := scanner.Bytes()
		var row Row
		err := json.Unmarshal(rawRow, &row)
		if err != nil {
			return nil, err
		}
		data[row.Key] = row.Value
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return data, nil
}

func (s *FileStorage) Add(key string, val string) error {
	// row := Row{Key: key, Value: val}
	// data, err := json.Marshal(row)
	// if err != nil {
	// 	return err
	// }
	// data = append(data, '\n')
	// _, err = s.file.Write(data)
	// if err != nil {
	// 	return err
	// }

	return nil
}
