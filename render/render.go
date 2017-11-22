// Package render allows the rendering of different template formats through a given interface
package render

import (
	"io"
	"sync"

	"github.com/unrolled/render"
)

var (
	// Since the unrolled renderer can lead to race conditions with
	// concurrent requests, lock resource access until template has
	// been sucessfully rendered
	hMutex = &sync.Mutex{}
	jMutex = &sync.Mutex{}
)

type renderer interface {
	HTML(w io.Writer, status int, name string, binding interface{}, htmlOpt ...render.HTMLOptions) error
	JSON(w io.Writer, status int, v interface{}) error
}

// Renderer provides an instance of the renderer interface used to allow the rendering of
// HTML and JSON templates
var Renderer renderer

// HTML controls the rendering of an HTML template with a given name and template parameters to an io.Writer
func HTML(w io.Writer, status int, name string, binding interface{}, htmlOpt ...render.HTMLOptions) error {
	hMutex.Lock()
	defer hMutex.Unlock()
	return Renderer.HTML(w, status, name, binding, htmlOpt...)
}

// JSON controls the rendering of a JSON template with a given name and template parameters to an io.Writer
func JSON(w io.Writer, status int, v interface{}) error {
	jMutex.Lock()
	defer jMutex.Unlock()
	return Renderer.JSON(w, status, v)
}
