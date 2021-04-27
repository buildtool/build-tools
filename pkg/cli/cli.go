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
	mu     sync.Mutex
	Writer io.Writer
}

// New handler.
func New(w io.Writer) *Handler {
	if f, ok := w.(*os.File); ok {
		return &Handler{
			Writer: f,
		}
	}

	return &Handler{
		Writer: w,
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
	return nil
}

// Write implementation.
func (w *Writer) Write(b []byte) (int, error) {
	s := bufio.NewScanner(bytes.NewReader(b))

	for s.Scan() {
		if err := w.write(s.Text()); err != nil {
			return 0, err
		}
	}

	if err := s.Err(); err != nil {
		return 0, err
	}

	return len(b), nil
}

func (w *Writer) write(s string) error {
	w.log.Infof("%s\n", s)
	return nil
}
