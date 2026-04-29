package traceid

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidTraceID = errors.New("invalid TraceID")
)

// 前八字节为时间戳，剩下的为 nonce
type TraceID []byte

func New(nonceSize int) *TraceID {
	id, err := NewTraceID(rand.Reader, nonceSize)
	if err != nil {
		panic(err)
	}

	return id
}

func NewTraceID(rd io.Reader, nonceSize int) (*TraceID, error) {
	timestamp := time.Now().Unix()

	id := make([]byte, nonceSize+8)
	nonce := make([]byte, nonceSize)

	_, err := rd.Read(nonce)
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint64(id[:8], uint64(timestamp))

	copy(id[8:], nonce)

	tuid := TraceID(id)
	return &tuid, nil
}

func (id TraceID) Timestamp() int64 {
	if len(id) < 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(slices.Clone(id[:8])))
}

func (id TraceID) Nonce() []byte {
	return slices.Clone(id[8:])
}

func (id TraceID) NonceString() string {
	nonce := &big.Int{}
	nonce = nonce.SetBytes(id.Nonce())
	return nonce.Text(16)
}

func (id TraceID) String() string {
	timestamp := id.Timestamp()
	prefix := strconv.FormatInt(timestamp, 10)
	suffix := id.NonceString()

	return prefix + "-" + suffix
}

func ParseTraceID(idStr string) (TraceID, error) {
	parts := strings.Split(idStr, "-")
	if len(parts) != 2 {
		return nil, ErrInvalidTraceID
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, ErrInvalidTraceID
	}

	id := TraceID{}
	binary.BigEndian.PutUint64(id, uint64(timestamp))

	nonce, err := hex.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidTraceID
	}

	id = append(id, nonce...)

	return id, nil
}
