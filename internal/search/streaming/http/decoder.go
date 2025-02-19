package http

import (
	"bufio"
	"bytes"
	"io"

	"github.com/cockroachdb/errors"
)

// Decoder decodes streaming events from a Server Sent Event stream. We only
// support streams which are generated by Sourcegraph. IE this is not a fully
// compliant Server Sent Events decoder.
type Decoder struct {
	scanner *bufio.Scanner
	event   []byte
	data    []byte
	err     error
}

func NewDecoder(r io.Reader) *Decoder {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 4096), maxPayloadSize)
	// bufio.ScanLines, except we look for two \n\n which separate events.
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
			return i + 2, data[:i], nil
		}
		// If we're at EOF, we have a final, non-terminated event. This should
		// be empty.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
	scanner.Split(split)
	return &Decoder{
		scanner: scanner,
	}
}

// Scan advances the decoder to the next event in the stream. It returns
// false when it either hits the end of the stream or an error.
func (d *Decoder) Scan() bool {
	if !d.scanner.Scan() {
		d.err = d.scanner.Err()
		return false
	}

	// event: $event\n
	// data: json($data)\n\n
	data := d.scanner.Bytes()
	nl := bytes.Index(data, []byte("\n"))
	if nl < 0 {
		d.err = errors.Errorf("malformed event, no newline: %s", data)
		return false
	}

	eventK, event := splitColon(data[:nl])
	dataK, data := splitColon(data[nl+1:])

	if !bytes.Equal(eventK, []byte("event")) {
		d.err = errors.Errorf("malformed event, expected event: %s", eventK)
		return false
	}
	if !bytes.Equal(dataK, []byte("data")) {
		d.err = errors.Errorf("malformed event %s, expected data: %s", eventK, dataK)
		return false
	}

	d.event = event
	d.data = data
	return true
}

// Event returns the event name of the last decoded event
func (d *Decoder) Event() []byte {
	return d.event
}

// Event returns the event data of the last decoded event
func (d *Decoder) Data() []byte {
	return d.data
}

// Err returns the last encountered error
func (d *Decoder) Err() error {
	return d.err
}

func splitColon(data []byte) ([]byte, []byte) {
	i := bytes.Index(data, []byte(":"))
	if i < 0 {
		return bytes.TrimSpace(data), nil
	}
	return bytes.TrimSpace(data[:i]), bytes.TrimSpace(data[i+1:])
}
