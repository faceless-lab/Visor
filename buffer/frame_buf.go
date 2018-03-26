package buffer

import (
	"sync"
	"time"
)

var (
	fps          time.Duration = 60
	currentFrame *frame
	once         sync.Once
	ticker       = time.NewTicker(time.Second / fps)
)

type frame struct {
	Buffer chan []byte
	Data   []byte
}

func GetInstance() *frame {
	once.Do(func() {
		currentFrame = &frame{
			Buffer: make(chan []byte, fps),
		}

		go func() {
			for range ticker.C {
				currentFrame.Data = <-currentFrame.Buffer
			}
		}()
	})

	return currentFrame
}

func (f *frame) Close() {
	ticker.Stop()
}

func (f *frame) Push(buf []byte) {
	f.Buffer <- buf
}

func (f *frame) Get() []byte {
	return f.Data
}
