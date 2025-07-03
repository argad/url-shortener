package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Compress
type gzipWriter struct {
	w      http.ResponseWriter
	Writer *gzip.Writer
}

func (g *gzipWriter) Header() http.Header {
	return g.w.Header()
}

func (g *gzipWriter) Write(p []byte) (int, error) {
	return g.Writer.Write(p)
}

func (g *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		g.w.Header().Set("Content-Encoding", "gzip")
	}
	g.w.WriteHeader(statusCode)
}

func (g *gzipWriter) Close() error {
	return g.Writer.Close()
}

// Decompress
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		newWriter := w

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		contentType := r.Header.Get("Content-Type")
		shouldCompress := strings.HasPrefix(contentType, "application/json") ||
			strings.HasPrefix(contentType, "text/html")

		if sendsGzip && shouldCompress {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {

			cw := &gzipWriter{
				w:      newWriter,
				Writer: gzip.NewWriter(newWriter),
			}
			newWriter = cw
			defer cw.Close()
		}

		next.ServeHTTP(newWriter, r)

	})
}
