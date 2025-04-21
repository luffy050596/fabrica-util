// Package compress provides tools for compressing and decompressing data using deflate and gzip
package compress

import (
	"bytes"
	"compress/zlib"
	"slices"
	"sync"

	"github.com/go-pantheon/fabrica-util/errors"
)

var (
	compressMutex          sync.RWMutex
	defaultWeakThreshold   = 10 << 10  // 10KB
	defaultStrongThreshold = 512 << 10 // 512KB
	defaultWeakLevel       = zlib.BestSpeed
	defaultStrongLevel     = zlib.DefaultCompression
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	once = sync.Once{}
)

// Init init compress params
// weak: weak compress threshold, compress when data length is greater than this value
// strong: strong compress threshold, use higher compression rate when data length is greater than this value
func Init(weak, strong int) {
	compressMutex.Lock()
	defer compressMutex.Unlock()

	once.Do(func() {
		if weak > 0 {
			defaultWeakThreshold = weak
		}

		if strong > 0 {
			defaultStrongThreshold = strong
		}
	})
}

// Compress auto select compress strategy based on data length
// return compressed data, whether compression is performed, error info
func Compress(data []byte) (ret []byte, didCompress bool, err error) {
	dataLen := len(data)
	if dataLen == 0 {
		return []byte{}, false, nil
	}

	compressMutex.RLock()
	weakThreshold := defaultWeakThreshold
	strongThreshold := defaultStrongThreshold
	compressMutex.RUnlock()

	if dataLen < weakThreshold {
		return data, false, nil
	}

	level := defaultWeakLevel
	if dataLen >= strongThreshold {
		level = defaultStrongLevel
	}

	if level < zlib.BestSpeed || level > zlib.BestCompression {
		level = zlib.DefaultCompression
	}

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		bufferPool.Put(buffer)
	}()

	writer, err := zlib.NewWriterLevel(buffer, level)
	if err != nil {
		return nil, false, errors.Wrapf(err, "create zlib writer failed (level %d)", level)
	}

	if _, err = writer.Write(data); err != nil {
		return nil, false, errors.Wrap(err, "write to compressor failed")
	}

	if err = writer.Close(); err != nil {
		return nil, false, errors.Wrap(err, "close compressor failed")
	}

	ret = slices.Clone(buffer.Bytes())
	didCompress = true

	return ret, didCompress, err
}

// Decompress decompress data
func Decompress(data []byte) (ret []byte, err error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		err = errors.Wrap(err, "create zlib reader failed")

		return nil, err
	}

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		bufferPool.Put(buffer)
	}()

	if _, err = buffer.ReadFrom(reader); err != nil {
		err = errors.Wrap(err, "read from decompressor failed")

		return nil, err
	}

	if err = reader.Close(); err != nil {
		err = errors.Wrap(err, "close decompressor failed")

		return nil, err
	}

	ret = slices.Clone(buffer.Bytes())

	return ret, nil
}
