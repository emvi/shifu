package llms

import (
	"bytes"
	"net/http"
)

type bufferResponseWriter struct {
	header http.Header
	buffer bytes.Buffer
	status int
}

func newBufferResponseWriter() *bufferResponseWriter {
	return &bufferResponseWriter{
		header: make(http.Header),
		status: http.StatusOK,
	}
}

func (w *bufferResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferResponseWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *bufferResponseWriter) WriteHeader(status int) {
	w.status = status
}
