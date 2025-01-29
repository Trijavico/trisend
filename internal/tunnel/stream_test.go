package tunnel

import "testing"

func TestSetStream(t *testing.T) {
	key := "testKey"
	stream := make(chan Stream)

	SetStream(key, stream)

	if _, ok := GetStream(key); !ok {
		t.Errorf("expected stream to be set for key %s, but it was not found", key)
	}
}

func TestGetStream(t *testing.T) {
	key := "testKey"
	streamChan := make(chan Stream)

	SetStream(key, streamChan)

	retrievedChan, ok := GetStream(key)
	if !ok {
		t.Fatalf("expected to retrieve a stream for key %s, but it was not found", key)
	}

	if retrievedChan != streamChan {
		t.Errorf("expected to retrieve the correct stream for key %s, but got a different one", key)
	}
}

func TestDeleteStream(t *testing.T) {
	key := "testKey"
	streamChan := make(chan Stream)

	SetStream(key, streamChan)
	DeleteStream(key)

	if _, ok := GetStream(key); ok {
		t.Errorf("expected stream for key %s to be deleted, but it was found", key)
	}
}

func TestGetStreamNotFound(t *testing.T) {
	key := "nonExistentKey"
	if _, ok := GetStream(key); ok {
		t.Errorf("expected no stream for key \"%s\", but found one", key)
	}
}
