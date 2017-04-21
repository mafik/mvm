package mvm

import "sync"

type FanOut struct {
	in       chan string
	mutex    sync.Mutex
	channels []chan string
}

func MakeFanOut() *FanOut {
	fo := &FanOut{
		in: make(chan string),
	}
	go func() {
		for x := range fo.in {
			fo.mutex.Lock()
			for _, c := range fo.channels {
				c <- x
			}
			fo.mutex.Unlock()
		}
	}()
	return fo
}

func (fo *FanOut) Open() chan string {
	fo.mutex.Lock()
	defer fo.mutex.Unlock()
	c := make(chan string)
	fo.channels = append(fo.channels, c)
	return c
}

func (fo *FanOut) Close(c chan string) {
	fo.mutex.Lock()
	defer fo.mutex.Unlock()
	s := fo.channels
	for i, elem := range s {
		if elem == c {
			last := len(s) - 1
			s[i] = s[last]
			s[last] = nil
			s = s[:last]
			break
		}
	}
	fo.channels = s
}
