package storage

import (
	"reflect"
	"testing"
)

func TestFileStorage_test0(t *testing.T) {
	persistentStorage := NewFileStorage("/tmp/go-shortener-test-123.txt")
	var err error

	expected := map[string]string{"1": "1", "2": "2", "3": "3"}
	var result map[string]string
	err = persistentStorage.Add("1", "1")
	if err != nil {
		t.Errorf("%s", err)
	}
	err = persistentStorage.Add("2", "2")
	if err != nil {
		t.Errorf("%s", err)
	}
	err = persistentStorage.Add("3", "3")
	if err != nil {
		t.Errorf("%s", err)
	}
	result, err = persistentStorage.Load()
	if err != nil {
		t.Errorf("%s", err)
	}

	if reflect.DeepEqual(result, expected) {
		t.Errorf("Invalid result")
	}

	// close storage and load again
	persistentStorage.Close()
	persistentStorage = NewFileStorage("/tmp/go-shortener-test-123.txt")

	result, err = persistentStorage.Load()
	if err != nil {
		t.Errorf("%s", err)
	}

	if reflect.DeepEqual(result, expected) {
		t.Errorf("Invalid result")
	}
}
