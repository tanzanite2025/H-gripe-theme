package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var ErrRemoteObjectByteRangeUnavailable = errors.New("remote object byte range could not be verified")

type ObjectByteRangeMetadata struct {
	Size    int64
	ModTime time.Time
	ETag    string
}

// The caller must authorize the business record and exact object before opening
// this source. Metadata and ranges are server-side reads, never bearer URLs.
type ObjectByteRangeSource interface {
	io.ReadCloser
	Metadata() ObjectByteRangeMetadata
	OpenByteRange(context.Context, int64, int64) (io.ReadCloser, error)
}

type ObjectByteRangeSourceOpener interface {
	OpenObjectByteRangeSource(context.Context, string) (ObjectByteRangeSource, error)
}

type remoteObjectByteRangeSource struct {
	ctx        context.Context
	metadata   ObjectByteRangeMetadata
	openRange  func(context.Context, int64, int64) (io.ReadCloser, error)
	fullReader io.ReadCloser
	closed     bool
}

func validatePinnedRemoteObjectByteRangeMetadata(metadata ObjectByteRangeMetadata, encoding string) error {
	if metadata.Size < 0 || len(metadata.ETag) < 2 || len(metadata.ETag) > 1024 ||
		metadata.ETag[0] != '"' || metadata.ETag[len(metadata.ETag)-1] != '"' ||
		(encoding != "" && !strings.EqualFold(encoding, "identity")) {
		return ErrRemoteObjectByteRangeUnavailable
	}
	for _, character := range []byte(metadata.ETag[1 : len(metadata.ETag)-1]) {
		if character < 0x21 || character == '"' || character == 0x7f {
			return ErrRemoteObjectByteRangeUnavailable
		}
	}
	return nil
}

func (s *remoteObjectByteRangeSource) Metadata() ObjectByteRangeMetadata { return s.metadata }

func (s *remoteObjectByteRangeSource) OpenByteRange(ctx context.Context, start, length int64) (io.ReadCloser, error) {
	if s.closed || start < 0 || length <= 0 || start >= s.metadata.Size || length > s.metadata.Size-start {
		return nil, ErrRemoteObjectByteRangeUnavailable
	}
	if err := s.ctx.Err(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.openRange(ctx, start, length)
}

// Existing service callers can still read the full object. Opening the source
// alone never downloads a body, so an HTTP range does not first fetch a video.
func (s *remoteObjectByteRangeSource) Read(buffer []byte) (int, error) {
	if s.closed {
		return 0, io.ErrClosedPipe
	}
	if err := s.ctx.Err(); err != nil {
		return 0, err
	}
	if len(buffer) == 0 {
		return 0, nil
	}
	if s.metadata.Size == 0 {
		return 0, io.EOF
	}
	if s.fullReader == nil {
		reader, err := s.OpenByteRange(s.ctx, 0, s.metadata.Size)
		if err != nil {
			return 0, err
		}
		s.fullReader = reader
	}
	return s.fullReader.Read(buffer)
}

func (s *remoteObjectByteRangeSource) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	if s.fullReader != nil {
		return s.fullReader.Close()
	}
	return nil
}

type remoteObjectByteRangeExactLengthReadCloser struct {
	io.ReadCloser
	remaining int64
}

func (r *remoteObjectByteRangeExactLengthReadCloser) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	if r.remaining == 0 {
		return 0, io.EOF
	}
	if int64(len(buffer)) > r.remaining {
		buffer = buffer[:r.remaining]
	}
	count, err := r.ReadCloser.Read(buffer)
	r.remaining -= int64(count)
	if errors.Is(err, io.EOF) && r.remaining > 0 {
		err = io.ErrUnexpectedEOF
	}
	return count, err
}

func validateRemoteObjectByteRangeResponseBeforeExposingBody(body io.ReadCloser, metadata ObjectByteRangeMetadata, start, length, responseLength int64, contentRange, etag, encoding string) (io.ReadCloser, error) {
	if body == nil {
		return nil, ErrRemoteObjectByteRangeUnavailable
	}
	expected := fmt.Sprintf("bytes %d-%d/%d", start, start+length-1, metadata.Size)
	if responseLength != length || strings.TrimSpace(contentRange) != expected || etag != metadata.ETag ||
		(encoding != "" && !strings.EqualFold(encoding, "identity")) {
		_ = body.Close()
		return nil, ErrRemoteObjectByteRangeUnavailable
	}
	return &remoteObjectByteRangeExactLengthReadCloser{ReadCloser: body, remaining: length}, nil
}
