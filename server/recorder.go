package server

import (
	"github.com/dan-j/go-wasm/core"
	"log"
	"sync"
)

type Recorder struct {
	events map[string][]core.Event
	mu sync.Mutex
}

func (r *Recorder) Record(id string, event core.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[id] = append(r.events[id], event)
	return nil
}

func (r *Recorder) Hydrate(id string, thing *core.Thing) error {
	for i, e := range r.events[id] {
		log.Printf("hydrating event: %d: %T", i, e)
		if err := thing.ApplyEvent(e); err != nil {
			log.Println("failed to apply event: ", err)
			return err
		}
	}

	return nil
}