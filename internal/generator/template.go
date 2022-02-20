package generator

import "text/template"

var tpl = template.Must(
	template.New("generated").
		Parse(`
{{- "// Code generated with github.com/kolosovi/periodic, DO NOT EDIT" }}

package {{ .Package }}

import (
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"
	{{- if .TypePackage }}
	usertype "{{ .TypePackage }}"
	{{ end }}
)

type {{ .Name }}Cache struct {
	data   *unsafe.Pointer
	stopCh chan struct{}
	doneCh chan struct{}

	provider {{ .Name }}ValueProvider
	tracer   {{ .Name }}Tracer
	ticks    <-chan time.Time
}

type {{ .Name }}Options struct {
	Provider {{ .Name }}ValueProvider
	Tracer   {{ .Name }}Tracer
	Ticks    <-chan time.Time
}

type {{ .Name }}ValueProvider interface {
	NewValue() ({{ .FullTypeName }}, error)
}

type {{ .Name }}Tracer interface {
	OnUpdateError(err error)
	OnUpdateSuccess()
}

func New{{ .Name }}Cache(options {{ .Name }}Options) (*{{ .Name }}Cache, error) {
	if options.Provider == nil {
		return nil, fmt.Errorf("provider can't be nil")
	}
	if options.Ticks == nil {
		return nil, fmt.Errorf("ticks can't be nil")
	}
	ptr := unsafe.Pointer(nil)
	return &{{ .Name }}Cache{
		data:   &ptr,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),

		provider: options.Provider,
		tracer:   options.Tracer,
		ticks:    options.Ticks,
	}, nil
}

func (c *{{ .Name }}Cache) Get() ({{ .FullTypeName }}, bool) {
	ptr := atomic.LoadPointer(c.data)
	userTypePtr := (*{{ .FullTypeName }})(ptr)
	if userTypePtr == nil {
		var fallback {{ .FullTypeName }}
		return fallback, false
	}
	return *userTypePtr, true
}

func (c *{{ .Name }}Cache) Update() error {
	newValue, err := c.provider.NewValue()
	if err != nil {
		return fmt.Errorf("cannot update cache: %v", err)
	}
	atomic.StorePointer(c.data, unsafe.Pointer(&newValue))
	return nil
}

func (c *{{ .Name }}Cache) Close() {
	close(c.stopCh)
}

func (c *{{ .Name }}Cache) Start() {
	go c.run()
}

func (c *{{ .Name }}Cache) run() {
	for {
		select {
		case <-c.ticks:
			err := c.Update()
			if err != nil {
				c.onUpdateError(err)
			} else {
				c.onUpdateSuccess()
			}
		case <-c.stopCh:
			close(c.doneCh)
			return
		}
	}
}

func (c *{{ .Name }}Cache) onUpdateError(err error) {
	if c.tracer == nil {
		return
	}
	c.tracer.OnUpdateError(err)
}

func (c *{{ .Name }}Cache) onUpdateSuccess() {
	if c.tracer == nil {
		return
	}
	c.tracer.OnUpdateSuccess()
}

func (c *{{ .Name }}Cache) Done() <-chan struct{} {
	return c.doneCh
}
`,
		))
