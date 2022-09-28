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
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
)

func Test_New_Stdout(t *testing.T) {
	handler := New(os.Stdout)
	assert.Equal(t, os.Stdout, handler.Writer)
	assert.IsType(t, &os.File{}, handler.Writer)
}

func Test_New_NoFileWriter(t *testing.T) {
	w := &Writer{}
	handler := New(w)

	assert.Equal(t, w, handler.Writer)
	assert.IsType(t, w, handler.Writer)
}

func Test_HandleLog(t *testing.T) {
	mu := &checkLocker{}
	buff := &bytes.Buffer{}
	handler := Handler{
		mu:     mu,
		Writer: buff,
	}
	_ = handler.HandleLog(&log.Entry{
		Logger:    nil,
		Fields:    nil,
		Level:     log.WarnLevel,
		Timestamp: time.Time{},
		Message:   "<green>msg</green>",
	})
	assert.Equal(t, 1, mu.lockCalled)
	assert.Equal(t, 1, mu.unlockCalled)
	assert.Equal(t, "\x1b[0m\x1b[32mmsg\x1b[39m\x1b[0m", buff.String())
}

func Test_Writer_New(t *testing.T) {
	assert.NotNil(t, NewWriter(log.Log))
	assert.Nil(t, NewWriter(invalidLog{}))
}

func Test_Writer_Write(t *testing.T) {
	w := NewWriter(log.Log)
	logMsg := "test"
	written, err := w.Write([]byte(logMsg))

	assert.NoError(t, err)
	assert.Equal(t, len(logMsg), written)
}

func Test_Verbose(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	assert.True(t, Verbose(log.Log))

	log.SetLevel(log.InfoLevel)
	assert.False(t, Verbose(log.Log))
	assert.False(t, Verbose(invalidLog{}))
}

type checkLocker struct {
	lockCalled   int
	unlockCalled int
}

func (c *checkLocker) Lock() {
	c.lockCalled = c.lockCalled + 1
}

func (c *checkLocker) Unlock() {
	c.unlockCalled = c.unlockCalled + 1
}

type invalidLog struct {
}

func (i invalidLog) WithFields(fielder log.Fielder) *log.Entry {
	panic("implement me")
}

func (i invalidLog) WithField(s string, i2 interface{}) *log.Entry {
	panic("implement me")
}

func (i invalidLog) WithDuration(duration time.Duration) *log.Entry {
	panic("implement me")
}

func (i invalidLog) WithError(err error) *log.Entry {
	panic("implement me")
}

func (i invalidLog) Debug(s string) {
	panic("implement me")
}

func (i invalidLog) Info(s string) {
	panic("implement me")
}

func (i invalidLog) Warn(s string) {
	panic("implement me")
}

func (i invalidLog) Error(s string) {
	panic("implement me")
}

func (i invalidLog) Fatal(s string) {
	panic("implement me")
}

func (i invalidLog) Debugf(s string, i2 ...interface{}) {
	panic("implement me")
}

func (i invalidLog) Infof(s string, i2 ...interface{}) {
	panic("implement me")
}

func (i invalidLog) Warnf(s string, i2 ...interface{}) {
	panic("implement me")
}

func (i invalidLog) Errorf(s string, i2 ...interface{}) {
	panic("implement me")
}

func (i invalidLog) Fatalf(s string, i2 ...interface{}) {
	panic("implement me")
}

func (i invalidLog) Trace(s string) *log.Entry {
	panic("implement me")
}
