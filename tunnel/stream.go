package tunnel

import (
	"fmt"
	"io"
	"sync"
)

type Stream struct {
	Writer io.Writer
	Done   chan struct{}
	Error  chan struct{}
}

var (
	streamings = map[string]chan Stream{}
	details    sync.Map
	mutex      sync.RWMutex
)

type StreamDetails struct {
	Username string
	Pfp      string
	Filename string
}

func SetStream(key string, stream chan Stream, value *StreamDetails) {
	details.Store(key, value)

	mutex.Lock()
	defer mutex.Unlock()
	streamings[key] = stream
}

func GetStream(key string) (chan Stream, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	stream, ok := streamings[key]
	return stream, ok
}

func GetStreamDetails(key string) (*StreamDetails, error) {
	value, ok := details.Load(key)
	if !ok {
		return nil, fmt.Errorf("value not found")
	}

	return value.(*StreamDetails), nil
}

func DeleteStream(key string) {
	details.Delete(key)

	mutex.Lock()
	defer mutex.Unlock()
	delete(streamings, key)
}
