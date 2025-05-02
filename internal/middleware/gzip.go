package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GzipCompressWriter struct {
	rw         http.ResponseWriter
	gzipWriter *gzip.Writer
}

func NewCompressWriter(rw http.ResponseWriter) *GzipCompressWriter {
	return &GzipCompressWriter{
		rw:         rw,
		gzipWriter: gzip.NewWriter(rw),
	}
}

func (cw *GzipCompressWriter) Close() error {
	return cw.gzipWriter.Close()
}

func (cw *GzipCompressWriter) Write(p []byte) (int, error) {
	return cw.gzipWriter.Write(p)
}

func (cw *GzipCompressWriter) Header() http.Header {
	if cw.rw.Header().Get("Content-Encoding") != "gzip" {
		cw.rw.Header().Add("Content-Encoding", "gzip")
	}
	return cw.rw.Header()
}

func (cw *GzipCompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		cw.rw.Header().Add("Content-Encoding", "gzip")
	}
	cw.rw.WriteHeader(statusCode)
}

type GzipDecompressReader struct {
	body       io.ReadCloser
	gzipReader *gzip.Reader
}

func NewDecompressReader(r io.ReadCloser) (*GzipDecompressReader, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &GzipDecompressReader{
		body:       r,
		gzipReader: gz,
	}, nil
}

func (dr *GzipDecompressReader) Read(p []byte) (int, error) {
	return dr.gzipReader.Read(p)
}

func (dr *GzipDecompressReader) Close() error {
	err := dr.body.Close()
	if err != nil {
		fmt.Printf("failed to close original reader: %v\n", err)
	}
	dr.gzipReader.Close()
	return err
}

func CompressGzipMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				compWriter := NewCompressWriter(w)
				defer compWriter.Close()

				w = compWriter
			}

			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				decompReader, err := NewDecompressReader(r.Body)
				if err != nil {
					http.Error(w, "failed to decompress request body", http.StatusInternalServerError)
					return
				}
				defer decompReader.Close()

				r.Body = decompReader
			}

			h.ServeHTTP(w, r)
		})
	}
}

// var gzipWriterPool = sync.Pool{
// 	New: func() interface{} {
// 		return gzip.NewWriter(nil)
// 	},
// }
