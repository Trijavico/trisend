package tunnel

import (
	"io"
	"sync"
)

type Stream struct {
	Writer io.Writer
	Done   chan struct{}
}

var (
	streamings = map[string]chan Stream{}
	mutex      sync.RWMutex
)

func SetStream(key string, stream chan Stream) {
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

func DeleteStream(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(streamings, key)
}
