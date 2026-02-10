package serial

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	goserial "go.bug.st/serial"
)

// Querier abstracts serial port query execution for testing.
type Querier interface {
	Query(ctx context.Context, opts QueryOptions) (*QueryResult, error)
}

// QueryOptions holds the parameters for a serial port query.
type QueryOptions struct {
	Port      string
	Command   string
	BaudRate  int
	DataBits  int
	Parity    string
	StopBits  string
	TimeoutMs int
}

// QueryResult holds the response from a serial port query.
type QueryResult struct {
	Response string `json:"response"`
}

// PortOpener opens a serial port with the given mode. It exists as a seam for testing.
type PortOpener func(portName string, mode *goserial.Mode) (goserial.Port, error)

// PortQuerier implements Querier using a real serial port.
type PortQuerier struct {
	opener PortOpener
}

// NewPortQuerier creates a new PortQuerier that opens real serial ports.
func NewPortQuerier() *PortQuerier {
	return &PortQuerier{opener: goserial.Open}
}

// Query sends a command over serial and reads the response.
func (q *PortQuerier) Query(ctx context.Context, opts QueryOptions) (*QueryResult, error) {
	parity, err := parseParity(opts.Parity)
	if err != nil {
		return nil, fmt.Errorf("invalid parity: %w", err)
	}
	stopBits, err := parseStopBits(opts.StopBits)
	if err != nil {
		return nil, fmt.Errorf("invalid stop bits: %w", err)
	}

	mode := &goserial.Mode{
		BaudRate: opts.BaudRate,
		DataBits: opts.DataBits,
		Parity:   parity,
		StopBits: stopBits,
	}

	port, err := q.opener(opts.Port, mode)
	if err != nil {
		return nil, fmt.Errorf("open port %s: %w", opts.Port, err)
	}
	defer func() { _ = port.Close() }()

	timeout := time.Duration(opts.TimeoutMs) * time.Millisecond
	if err := port.SetReadTimeout(timeout); err != nil {
		return nil, fmt.Errorf("set read timeout: %w", err)
	}

	cmd := opts.Command
	if !strings.HasSuffix(cmd, "\n") {
		cmd += "\n"
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	_, err = port.Write([]byte(cmd))
	if err != nil {
		return nil, fmt.Errorf("write command: %w", err)
	}

	response, err := readUntilNewline(port, timeout)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &QueryResult{
		Response: strings.TrimRight(response, "\r\n"),
	}, nil
}

// readUntilNewline reads from the port until a newline is encountered or the timeout expires.
func readUntilNewline(port goserial.Port, timeout time.Duration) (string, error) {
	var buf bytes.Buffer
	oneByte := make([]byte, 1)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		n, err := port.Read(oneByte)
		if n > 0 {
			buf.Write(oneByte[:n])
			if oneByte[0] == '\n' {
				return buf.String(), nil
			}
		}
		if err != nil {
			if buf.Len() > 0 {
				return buf.String(), nil
			}
			return "", fmt.Errorf("no response from device: %w", err)
		}
		if n == 0 {
			// Read timeout with no data — check if we already have partial data.
			if buf.Len() > 0 {
				return buf.String(), nil
			}
			return "", fmt.Errorf("no response from device (timeout)")
		}
	}

	if buf.Len() > 0 {
		return buf.String(), nil
	}
	return "", fmt.Errorf("no response from device (timeout)")
}

// parseParity converts a string parity name to a goserial.Parity value.
func parseParity(s string) (goserial.Parity, error) {
	switch strings.ToLower(s) {
	case "none", "n", "":
		return goserial.NoParity, nil
	case "odd", "o":
		return goserial.OddParity, nil
	case "even", "e":
		return goserial.EvenParity, nil
	case "mark", "m":
		return goserial.MarkParity, nil
	case "space", "s":
		return goserial.SpaceParity, nil
	default:
		return goserial.NoParity, fmt.Errorf("unknown parity %q (valid: none, odd, even, mark, space)", s)
	}
}

// parseStopBits converts a string stop bits name to a goserial.StopBits value.
func parseStopBits(s string) (goserial.StopBits, error) {
	switch strings.ToLower(s) {
	case "1", "":
		return goserial.OneStopBit, nil
	case "1.5":
		return goserial.OnePointFiveStopBits, nil
	case "2":
		return goserial.TwoStopBits, nil
	default:
		return goserial.OneStopBit, fmt.Errorf("unknown stop bits %q (valid: 1, 1.5, 2)", s)
	}
}
