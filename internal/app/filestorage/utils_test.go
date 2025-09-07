package filestorage

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
)

type TestSender struct {
	t    *testing.T
	data []byte
	ind  *int
}

func (s TestSender) Send(stream *pb.FileStream) error {
	require.NotNil(s.t, stream.GetChunkData())
	chunkLen := len(stream.GetChunkData())
	require.LessOrEqual(s.t, chunkLen+*s.ind, len(s.data))
	require.Equal(s.t, s.data[*s.ind:*s.ind+chunkLen], stream.GetChunkData())
	*s.ind += chunkLen
	return nil
}

func TestFromReader2FileStream(t *testing.T) {
	const msgLength = 100000
	data := make([]byte, msgLength)
	rand.Read(data)
	sender := TestSender{t: t, data: data, ind: new(int)}
	*sender.ind = 0
	FromReader2FileStream(bytes.NewReader(data), sender)
	require.Equal(t, len(data), *sender.ind)
}

type TestReciever struct {
	data      []byte
	ind       *int
	chunkSize int
}

func (r TestReciever) Recv() (*pb.FileStream, error) {
	if *r.ind >= len(r.data) {
		return nil, io.EOF
	}
	chunkLen := min(len(r.data)-*r.ind, r.chunkSize)
	chunk := r.data[*r.ind : *r.ind+chunkLen]
	*r.ind += chunkLen
	return &pb.FileStream{Data: &pb.FileStream_ChunkData{ChunkData: chunk}}, nil
}

func TestFileStreamReader(t *testing.T) {
	const msgLength = 100000
	initialData := make([]byte, msgLength)
	rand.Read(initialData)
	streamReciever := TestReciever{data: initialData[:], ind: new(int), chunkSize: 1001}
	*streamReciever.ind = 0
	fileReader := NewFileStreamReader(streamReciever)
	getData, err := io.ReadAll(fileReader)
	require.NoError(t, err)
	require.Equal(t, initialData, getData)
}
