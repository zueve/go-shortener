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

	expected := make([]Row, 3)
	for i, ch := range "123" {
		ch := string(ch)
		row := Row{Key: ch, OriginURL: ch, UserID: ch}
		err := persistentStorage.Add(row)
		assert.Nil(t, err)
		expected[i] = row
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
