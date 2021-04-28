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
