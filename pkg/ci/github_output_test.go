// MIT License
//
// Copyright (c) 2018 buildtool
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

package ci

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteGitHubOutput_NotInGitHubActions(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "")
	tmpFile := t.TempDir() + "/output"
	_ = os.WriteFile(tmpFile, []byte{}, 0o644)
	t.Setenv("GITHUB_OUTPUT", tmpFile)

	WriteGitHubOutput("digest", "sha256:abc123")

	content, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)
	assert.Empty(t, string(content))
}

func TestWriteGitHubOutput_NoOutputFile(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_OUTPUT", "")

	// Should not panic
	WriteGitHubOutput("digest", "sha256:abc123")
}

func TestWriteGitHubOutput_WritesOutput(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	tmpFile := t.TempDir() + "/output"
	_ = os.WriteFile(tmpFile, []byte{}, 0o644)
	t.Setenv("GITHUB_OUTPUT", tmpFile)

	WriteGitHubOutput("digest", "sha256:abc123")

	content, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, "digest=sha256:abc123\n", string(content))
}

func TestWriteGitHubOutput_AppendsToExistingContent(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	tmpFile := t.TempDir() + "/output"
	_ = os.WriteFile(tmpFile, []byte("existing=value\n"), 0o644)
	t.Setenv("GITHUB_OUTPUT", tmpFile)

	WriteGitHubOutput("digest", "sha256:abc123")

	content, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, "existing=value\ndigest=sha256:abc123\n", string(content))
}

func TestWriteGitHubOutput_InvalidPath(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_OUTPUT", "/nonexistent/path/output")

	// Should not panic
	WriteGitHubOutput("digest", "sha256:abc123")
}
