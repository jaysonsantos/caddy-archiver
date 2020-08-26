package archiver

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/mholt/archiver/v3"
)

const (
	zipContentType     = "application/zip"
	zipExtension       = "zip"
	tarContentType     = "application/tar"
	tarExtension       = "tar"
	tarGzipContentType = "application/tar+gzip"
	tarGzipExtension   = "tar.gz"

	acceptHeader             = "Accept"
	contentTypeHeader        = "Content-Type"
	contentDispositionHeader = "Content-Disposition"
)

var (
	extensionToContentType = map[string]string{
		zipExtension:     zipContentType,
		tarExtension:     tarContentType,
		tarGzipExtension: tarGzipContentType,
	}

	contentTypeToExtension = map[string]string{
		zipContentType:     zipExtension,
		tarContentType:     tarExtension,
		tarGzipContentType: tarGzipExtension,
	}
)

func (a CaddyArchiver) ServeHTTP(writer http.ResponseWriter, request *http.Request, next caddyhttp.Handler) error {
	contentType, ok := parseAcceptHeader(request)
	if !ok {
		return next.ServeHTTP(writer, request)
	}

	a.logger.Info(fmt.Sprintf("Downloading as %v", contentType))

	downloadPath := request.URL.Path
	pathInsideRoot := a.pathInsideRoot(downloadPath)
	a.logger.Debug(fmt.Sprintf("path %v inside root %v = %v", downloadPath, a.Root, pathInsideRoot))

	if !pathInsideRoot {
		a.logger.Info(fmt.Sprintf("path %v is not inside root %v", downloadPath, a.Root))
		return next.ServeHTTP(writer, request)
	}

	extension := contentTypeToExtension[contentType]
	return a.streamFolderAsArchive(downloadPath, extension, writer)
}

func parseAcceptHeader(request *http.Request) (string, bool) {
	accept, ok := request.Header[acceptHeader]
	if !ok {
		return "", false
	}
	var validContentType string
	for _, contentType := range accept {
		if contentTypeToExtension[contentType] != "" {
			validContentType = contentType
		}
	}
	if validContentType == "" {
		return "", false
	}

	return validContentType, true
}

func (a *CaddyArchiver) pathInsideRoot(path string) bool {
	finalPath := filepath.Join(a.Root, path)
	finalPath, err := filepath.Abs(finalPath)
	if finalPath == a.Root {
		return true
	}
	if err != nil {
		a.logger.Debug(fmt.Sprintf("failed to get absolute path root: %v path: %v %v", a.Root, path, err))
	}

	finalPathDir := filepath.Dir(finalPath)
	return strings.HasPrefix(finalPathDir, a.Root)
}

func (a *CaddyArchiver) streamFolderAsArchive(downloadFolderName, extension string, w http.ResponseWriter) error {
	baseFolderName := path.Base(downloadFolderName)
	if err := validateExtension(extension); err != nil {
		return writeUnsupportedMediaType(w, err)
	}
	contentType := extensionToContentType[extension]

	writer, err := a.getArchiveWriter(contentType)
	if err != nil {
		return writeUnsupportedMediaType(w, err)
	}
	defer writer.Close()

	w.Header().Set(contentTypeHeader, contentType)
	w.Header().Set(contentDispositionHeader, fmt.Sprintf("attachment; filename=\"%s.%s\"", path.Base(baseFolderName), extension))
	err = writer.Create(w)
	if err != nil {
		return err
	}

	downloadFolderInfo, err := os.Stat(downloadFolderName)
	if err != nil {
		if os.IsNotExist(err) {
			return a.notFound()
		}
		return err
	}

	err = filepath.Walk(downloadFolderName, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil {
			return fmt.Errorf("nil file info")
		}

		// open the file, if it has any content
		var file io.ReadCloser
		if info.Mode().IsRegular() {
			file, err = os.Open(fpath)
			if err != nil {
				return fmt.Errorf("%s: opening: %v", fpath, err)
			}
			defer file.Close()
		}

		// make its archive-internal name
		internalName, err := archiver.NameInArchive(downloadFolderInfo, downloadFolderName, fpath)
		if err != nil {
			return fmt.Errorf("making internal archive name for %s: %v", fpath, err)
		}

		// write the file to the archive
		err = writer.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   info,
				CustomName: internalName,
			},
			ReadCloser: file,
		})
		if err != nil {
			return fmt.Errorf("writing file to archive: %v", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walking %s: %v", downloadFolderName, err)
	}

	return nil
}

func (a *CaddyArchiver) getArchiveWriter(contentType string) (archiver.Writer, error) {
	switch contentType {
	default:
		return nil, fmt.Errorf("A file format with content type %v is not supported", contentType)
	case zipContentType:
		return &archiver.Zip{
			CompressionLevel:       0,
			MkdirAll:               true,
			SelectiveCompression:   true,
			ImplicitTopLevelFolder: true,
		}, nil
	case tarContentType:
		return &archiver.Tar{MkdirAll: true, ImplicitTopLevelFolder: true}, nil
	case tarGzipContentType:
		return &archiver.TarGz{Tar: &archiver.Tar{MkdirAll: true, ImplicitTopLevelFolder: true}}, nil
	}
}

func writeUnsupportedMediaType(w http.ResponseWriter, err error) error {
	w.WriteHeader(415)
	_, err = w.Write([]byte(fmt.Sprint(err)))
	return err
}

func (a *CaddyArchiver) notFound() error {
	return caddyhttp.Error(http.StatusNotFound, nil)
}
