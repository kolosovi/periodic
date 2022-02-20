package main

import (
	"encoding/json"
	"example/caches"
	"example/types"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

type stringValueProvider struct {
	counter int64
}

func (p *stringValueProvider) NewValue() (string, error) {
	value := fmt.Sprintf("cached %v", p.counter)
	atomic.AddInt64(&p.counter, 1)
	return value, nil
}

type compositeValueProvider struct {
}

func (p *compositeValueProvider) NewValue() (types.Foo, error) {
	return types.Foo{
		Bar: rand.Int(),
		Baz: rand.Int(),
	}, nil
}

type Handler struct {
	stringCache *caches.StringCache
	fooCache    *caches.FooCache
}

type stringResponseDTO struct {
	Value string `json:"value"`
	OK    bool   `json:"ok"`
}

func (h *Handler) HandleString(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	value, ok := h.stringCache.Get()
	dto := stringResponseDTO{
		Value: value,
		OK:    ok,
	}
	body, err := json.Marshal(dto)
	if err != nil {
		log.Printf("cannot marshal response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	header := w.Header()
	header["Content-Type"] = []string{"application/json"}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		log.Printf("cannot write response: %v", err)
	}
}

type compositeResponseDTO struct {
	Value types.Foo `json:"value"`
	OK    bool      `json:"ok"`
}

func (h *Handler) HandleComposite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	value, ok := h.fooCache.Get()
	dto := compositeResponseDTO{
		Value: value,
		OK:    ok,
	}
	body, err := json.Marshal(dto)
	if err != nil {
		log.Printf("cannot marshal response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	header := w.Header()
	header["Content-Type"] = []string{"application/json"}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		log.Printf("cannot write response: %v", err)
	}
}

func main() {
	stringCache, err := caches.NewStringCache(caches.StringOptions{
		Provider: &stringValueProvider{},
		Ticks:    time.Tick(time.Second),
	})
	if err != nil {
		log.Fatalf("cannot create string cache: %v", err)
	}
	fooCache, err := caches.NewFooCache(caches.FooOptions{
		Provider: &compositeValueProvider{},
		Ticks:    time.Tick(time.Second),
	})
	if err != nil {
		log.Fatalf("cannot create composite cache: %v", err)
	}
	handler := &Handler{stringCache: stringCache, fooCache: fooCache}
	stringCache.Start()
	fooCache.Start()
	http.HandleFunc("/cached/string", handler.HandleString)
	http.HandleFunc("/cached/composite", handler.HandleComposite)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("got serve error: %v", err)
	}
	stringCache.Close()
	<-stringCache.Done()
	fooCache.Close()
	<-fooCache.Done()
}
