package bandwidthController

import (
	"io"
	"sync/atomic"

	ratelimitedreader "github.com/idanmadmon/rate-limited-reader"
)

type FileReader struct {
	reader    *ratelimitedreader.RateLimitedReader
	bytesRead int64
	rateLimit int64
	callback  func() // called on Close
}

func NewFileReader(r io.Reader, limit int64, callback func()) *FileReader {
	return &FileReader{
		reader:   ratelimitedreader.NewRateLimitedReader(r, limit),
		callback: callback,
	}
}

func NewFileReadCloser(r io.ReadCloser, limit int64, callback func()) *FileReader {
	return &FileReader{
		reader:   ratelimitedreader.NewRateLimitedReadCloser(r, limit),
		callback: callback,
	}
}

func (fr *FileReader) Read(p []byte) (n int, err error) {
	n, err = fr.reader.Read(p)
	atomic.AddInt64(&fr.bytesRead, int64(n))
	return n, err
}

func (fr *FileReader) Close() error {
	err := fr.reader.Close()

	if fr.callback != nil {
		fr.callback()
	}

	return err
}

func (fr *FileReader) UpdateRateLimit(newLimit int64) {
	atomic.StoreInt64(&fr.rateLimit, newLimit)
	fr.reader.UpdateLimit(newLimit)
}

func (fr *FileReader) GetRateLimit() int64 {
	return atomic.LoadInt64(&fr.rateLimit)
}

func (fr *FileReader) BytesRead() int64 {
	return atomic.LoadInt64(&fr.bytesRead)
}
