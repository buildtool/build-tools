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

package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFilesForTarget(t *testing.T) {
	type args struct {
		dir    string
		target string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "non existing directory",
			args: args{
				dir:    "testdata/not_existing",
				target: "local",
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "only common files",
			args: args{
				dir:    "testdata/only_common_files",
				target: "local",
			},
			want:    []string{"config.yaml", "deploy.yaml"},
			wantErr: false,
		},
		{
			name: "specific config files",
			args: args{
				dir:    "testdata/specific_config_files",
				target: "local",
			},
			want:    []string{"config-local.yaml", "deploy.yaml"},
			wantErr: false,
		},
		{
			name: "specific and common config files - specific",
			args: args{
				dir:    "testdata/specific_and_common_config_files",
				target: "local",
			},
			want:    []string{"config-local.yaml", "deploy.yaml"},
			wantErr: false,
		},
		{
			name: "specific and common config files - common",
			args: args{
				dir:    "testdata/specific_and_common_config_files",
				target: "prod",
			},
			want:    []string{"config.yaml", "deploy.yaml"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindFilesForTarget(tt.args.dir, tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindFilesForTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			files := make([]string, len(got))
			for i, f := range got {
				files[i] = f.Name()
			}
			assert.Equal(t, tt.want, files)
		})
	}
}

func TestFindScriptsForTarget(t *testing.T) {
	type args struct {
		dir    string
		target string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "non existing directory",
			args: args{
				dir:    "testdata/not_existing",
				target: "local",
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "only common files",
			args: args{
				dir:    "testdata/only_common_files",
				target: "local",
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "specific config files",
			args: args{
				dir:    "testdata/specific_config_files",
				target: "local",
			},
			want:    []string{"setup-local.sh"},
			wantErr: false,
		},
		{
			name: "specific and common config files - specific",
			args: args{
				dir:    "testdata/specific_and_common_config_files",
				target: "local",
			},
			want:    []string{"setup-local.sh"},
			wantErr: false,
		},
		{
			name: "specific and common config files - common",
			args: args{
				dir:    "testdata/specific_and_common_config_files",
				target: "prod",
			},
			want:    []string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindScriptsForTarget(tt.args.dir, tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindScriptsForTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			files := make([]string, len(got))
			for i, f := range got {
				files[i] = f.Name()
			}
			assert.Equal(t, tt.want, files)
		})
	}
}
