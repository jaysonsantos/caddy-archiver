package archiver

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mholt/archiver/v3"
)

func Test_parseAcceptHeader(t *testing.T) {
	type args struct {
		request *http.Request
	}
	type testCaseType struct {
		name        string
		args        args
		contentType string
		ok          bool
	}
	missingHeaderRequest := httptest.NewRequest("get", "/a", nil)
	invalidContentTypeRequest := httptest.NewRequest("get", "/a", nil)
	invalidContentTypeRequest.Header.Set("accept", "invalid")

	tests := []testCaseType{
		{"missing-header", args{request: missingHeaderRequest}, "", false},
		{"invalid-content-type", args{request: invalidContentTypeRequest}, "", false},
	}

	for contentType := range contentTypeToExtension {
		request := httptest.NewRequest("get", "/a", nil)
		request.Header.Set("accept", contentType)
		testCase := testCaseType{contentType, args{request: request}, contentType, true}
		tests = append(tests, testCase)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentType, ok := parseAcceptHeader(tt.args.request)
			if contentType != tt.contentType {
				t.Errorf("parseAcceptHeader() contentType = %v, contentType %v", contentType, tt.contentType)
			}
			if ok != tt.ok {
				t.Errorf("parseAcceptHeader() ok = %v, contentType %v", ok, tt.ok)
			}
		})
	}
}

func TestFileServer_getArchiveWriter(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        interface{}
		wantErr     bool
	}{
		{name: "zip", contentType: "application/zip", want: new(archiver.Zip)},
		{name: "tar", contentType: "application/tar", want: new(archiver.Tar)},
		{name: "err", contentType: "application/wrong", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caddyArchiver := new(CaddyArchiver)
			got, err := caddyArchiver.getArchiveWriter(tt.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("getArchiveWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("getArchiveWriter() got = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("make sure all archives are implemented", func(y *testing.T) {
		caddyArchiver := new(CaddyArchiver)
		for _, mimeType := range extensionToContentType {
			if _, err := caddyArchiver.getArchiveWriter(mimeType); err != nil {
				t.Errorf("archive writer is for content type %v is not implemented yet: %v", mimeType, err)
			}
		}
	})
}

func Test_pathInsideRoot(t *testing.T) {
	type args struct {
		root string
		path string
	}
	root, _ := filepath.Abs(".")
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"dir-one-level-above", args{root, ".."}, false},
		{"file-one-level-above", args{root, "../test.txt"}, false},
		{"file-on-root", args{root, "../test.txt"}, false},
		{"same-level-dir", args{root, "."}, true},
		{"same-level-dir", args{root, "./"}, true},
		{"same-level-file", args{root, "./test.txt"}, true},
		{"dir-one-level-below", args{root, "./b"}, true},
		{"file-one-level-below", args{root, "./b/test.txt"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dumb := new(CaddyArchiver)
			dumb.Root = tt.args.root
			if got := dumb.pathInsideRoot(tt.args.path); got != tt.want {
				t.Errorf("pathInsideRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}
