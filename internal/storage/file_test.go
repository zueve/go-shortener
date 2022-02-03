package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileStorage_test0(t *testing.T) {
	file, err := os.CreateTemp("", "go_shortener")
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.Remove(file.Name())
	persistentStorage, err := NewFileStorage(file.Name())
	assert.Nil(t, err)

	expected := map[string]string{"1": "1", "2": "2", "3": "3"}
	for _, ch := range "123" {
		err := persistentStorage.Add(string(ch), string(ch))
		assert.Nil(t, err)
	}

	result, err := persistentStorage.Load()
	assert.Nil(t, err)
	assert.Equal(t, result, expected, "they should be equal")

	// close storage and load again
	persistentStorage.Close()
	persistentStorage, _ = NewFileStorage(file.Name())
	defer persistentStorage.Close()

	result, err = persistentStorage.Load()
	assert.Nil(t, err)
	assert.Equal(t, result, expected, "they should be equal")
}
