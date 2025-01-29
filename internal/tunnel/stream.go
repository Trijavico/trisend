package tunnel

import (
	"net/http"
	"sync"
	"time"
)

type Stream struct {
	Writer http.ResponseWriter
	Done   chan struct{}
	Error  chan struct{}
}

var (
	streamings    = map[string]chan Stream{}
	streamDetails sync.Map
	mutex         sync.RWMutex
)

type StreamDetails struct {
	Username string
	Pfp      string
	Filename string
	Expires  time.Time
}

func SetStream(key string, stream chan Stream, value *StreamDetails) {
	streamDetails.Store(key, value)

	mutex.Lock()
	defer mutex.Unlock()
	streamings[key] = stream
}

func GetStream(key string) (chan Stream, bool) {
	details, ok := GetStreamDetails(key)
	if !ok {
		return nil, ok
	} else if time.Now().After(details.Expires) {
		return nil, false
	}

	mutex.Lock()
	defer mutex.Unlock()
	stream, ok := streamings[key]

	return stream, ok
}

func GetStreamDetails(key string) (*StreamDetails, bool) {
	value, ok := streamDetails.Load(key)
	if !ok {
		return nil, ok
	}

	details := value.(*StreamDetails)
	if time.Now().After(details.Expires) {
		return nil, false
	}

	return details, ok
}

func DeleteStream(key string) {
	streamDetails.Delete(key)

	mutex.Lock()
	defer mutex.Unlock()
	delete(streamings, key)
}
