package storage

import (
	"os"
	"reflect"
	"testing"
)

func TestFileStorage_test0(t *testing.T) {
	filename := "/tmp/go-shortener-test-123.txt"
	os.Remove(filename)
	defer os.Remove(filename)
	persistentStorage := NewFileStorage(filename)

	expected := map[string]string{"1": "1", "2": "2", "3": "3"}
	for _, ch := range "123" {
		err := persistentStorage.Add(string(ch), string(ch))
		if err != nil {
			t.Errorf("%s %s", string(ch), err)
		}
	}

	result, err := persistentStorage.Load()
	if err != nil {
		t.Errorf("%s", err)
	}

	persistentStorage.Close()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Invalid result 1 %s, %s", result, expected)
	}

	// close storage and load again
	persistentStorage.Close()
	persistentStorage = NewFileStorage(filename)

	result, err = persistentStorage.Load()
	if err != nil {
		t.Errorf("%s", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Invalid result 2 %s, %s", result, expected)
	}
	persistentStorage.Close()
}
