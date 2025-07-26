package filestorage

import (
	"fmt"
	"io"

	"github.com/google/uuid"
	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
)

// Get file identifier for storing.
func CreateFileId(_ *pb.FileInfo) string {
	return uuid.NewString()
}

// Utils for converting io.reader to stream
type StreamSender interface {
	Send(*pb.FileStream) error
}

// Write file from reader to stream.
func FromReader2FileStream(reader io.Reader, stream StreamSender) error {
	buf := make([]byte, ChunkSize)

	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error when read file: %w", err)
		}
		stream.Send(&pb.FileStream{Data: &pb.FileStream_ChunkData{ChunkData: buf[:n]}})
		if err == io.EOF {
			break
		}
	}
	return nil
}

// Utils for converting stream to io.reader.
type StreamReciever interface {
	Recv() (*pb.FileStream, error)
}

// Read from stream to reader.
type FileStreamReader struct {
	stream     StreamReciever
	buffer     []byte
	bufferSize *int
	pos        *int
}

// New reader from stream to reader.
func NewFileStreamReader(stream StreamReciever) *FileStreamReader {
	pos := new(int)
	*pos = 0
	bufferSize := new(int)
	*bufferSize = 0
	return &FileStreamReader{stream: stream, buffer: make([]byte, ChunkSize), pos: pos, bufferSize: bufferSize}
}

// Define Read function for io.reader.
func (w FileStreamReader) Read(p []byte) (int, error) {
	for {
		if *w.pos < *w.bufferSize {
			n := copy(p, w.buffer[*w.pos:*w.bufferSize])
			*w.pos += n
			return n, nil
		}

		resp, err := w.stream.Recv()
		if err != nil {
			if err == io.EOF {
				return 0, io.EOF
			}
			return 0, fmt.Errorf("failed to upload: %w", err)
		}

		*w.bufferSize = copy(w.buffer, resp.GetChunkData())
		*w.pos = 0
	}
}
