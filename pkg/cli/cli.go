// MIT License
//
// Copyright (c) 2021 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/apex/log"
	"github.com/liamg/tml"
)

// Handler implementation.
type Handler struct {
	mu     sync.Locker
	Writer io.Writer
}

// New handler.
func New(w io.Writer) *Handler {
	if f, ok := w.(*os.File); ok {
		return &Handler{
			Writer: f,
			mu:     &sync.Mutex{},
		}
	}

	return &Handler{
		Writer: w,
		mu:     &sync.Mutex{},
	}
}

// HandleLog implements cli.Handler.
func (h *Handler) HandleLog(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, _ = fmt.Fprint(h.Writer, tml.Sprintf(e.Message))
	return nil
}

type Writer struct {
	log   log.Interface
	level log.Level
}

func NewWriter(ctx log.Interface) *Writer {
	if logger, ok := ctx.(*log.Logger); ok {
		return &Writer{
			log:   ctx,
			level: logger.Level,
		}
	}
	// TODO Panic?
	return nil
}

// Write implementation.
func (w *Writer) Write(b []byte) (int, error) {
	s := bufio.NewScanner(bytes.NewReader(b))

	for s.Scan() {
		w.log.Infof("%s\n", s.Text())
	}

	if err := s.Err(); err != nil {
		return 0, err
	}

	return len(b), nil
}

func Verbose(ctx log.Interface) bool {
	if logger, ok := ctx.(*log.Logger); ok {
		return logger.Level <= log.DebugLevel
	}
	return false
}
