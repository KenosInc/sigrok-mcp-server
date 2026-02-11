package serial

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	goserial "go.bug.st/serial"
)

// mockPort implements go.bug.st/serial.Port for testing.
type mockPort struct {
	readData  []byte
	readPos   int
	readErr   error
	writeErr  error
	written   []byte
	closed    bool
	closeErr  error
	timeout   time.Duration
}

func (p *mockPort) Read(buf []byte) (int, error) {
	if p.readErr != nil && p.readPos >= len(p.readData) {
		return 0, p.readErr
	}
	if p.readPos >= len(p.readData) {
		return 0, nil
	}
	n := copy(buf, p.readData[p.readPos:])
	p.readPos += n
	return n, nil
}

func (p *mockPort) Write(data []byte) (int, error) {
	if p.writeErr != nil {
		return 0, p.writeErr
	}
	p.written = append(p.written, data...)
	return len(data), nil
}

func (p *mockPort) SetReadTimeout(t time.Duration) error {
	p.timeout = t
	return nil
}

func (p *mockPort) Close() error {
	p.closed = true
	return p.closeErr
}

// Unused methods to satisfy the goserial.Port interface.
func (p *mockPort) ResetInputBuffer() error                          { return nil }
func (p *mockPort) ResetOutputBuffer() error                         { return nil }
func (p *mockPort) SetDTR(dtr bool) error                            { return nil }
func (p *mockPort) SetRTS(rts bool) error                            { return nil }
func (p *mockPort) GetModemStatusBits() (*goserial.ModemStatusBits, error) { return nil, nil }
func (p *mockPort) SetMode(mode *goserial.Mode) error                { return nil }
func (p *mockPort) Break(time.Duration) error                        { return nil }
func (p *mockPort) Drain() error                                     { return nil }

func TestQuery(t *testing.T) {
	tests := []struct {
		name     string
		opts     QueryOptions
		port     *mockPort
		wantResp string
		wantCmd  string
	}{
		{
			name: "basic SCPI IDN query",
			opts: QueryOptions{
				Port:      "/dev/ttyUSB0",
				Command:   "*IDN?",
				BaudRate:  115200,
				DataBits:  8,
				Parity:    "none",
				StopBits:  "1",
				TimeoutMs: 1000,
			},
			port: &mockPort{
				readData: []byte("OWON,XDM1241,1234567,V1.0\n"),
			},
			wantResp: "OWON,XDM1241,1234567,V1.0",
			wantCmd:  "*IDN?\n",
		},
		{
			name: "response with CR+LF",
			opts: QueryOptions{
				Port:      "/dev/ttyUSB0",
				Command:   "MEAS:VOLT:DC?",
				BaudRate:  9600,
				DataBits:  8,
				Parity:    "none",
				StopBits:  "1",
				TimeoutMs: 1000,
			},
			port: &mockPort{
				readData: []byte("+1.234E+00\r\n"),
			},
			wantResp: "+1.234E+00",
			wantCmd:  "MEAS:VOLT:DC?\n",
		},
		{
			name: "command already has newline",
			opts: QueryOptions{
				Port:      "/dev/ttyUSB0",
				Command:   "*IDN?\n",
				BaudRate:  115200,
				DataBits:  8,
				Parity:    "none",
				StopBits:  "1",
				TimeoutMs: 1000,
			},
			port: &mockPort{
				readData: []byte("OWON,XDM1241,1234567,V1.0\n"),
			},
			wantResp: "OWON,XDM1241,1234567,V1.0",
			wantCmd:  "*IDN?\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &PortQuerier{
				opener: func(_ string, _ *goserial.Mode) (goserial.Port, error) {
					return tt.port, nil
				},
			}

			result, err := q.Query(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Response != tt.wantResp {
				t.Errorf("response = %q, want %q", result.Response, tt.wantResp)
			}
			if string(tt.port.written) != tt.wantCmd {
				t.Errorf("written = %q, want %q", string(tt.port.written), tt.wantCmd)
			}
			if !tt.port.closed {
				t.Error("port was not closed")
			}
		})
	}
}

func TestQueryOpenError(t *testing.T) {
	q := &PortQuerier{
		opener: func(_ string, _ *goserial.Mode) (goserial.Port, error) {
			return nil, errors.New("permission denied")
		},
	}

	_, err := q.Query(context.Background(), QueryOptions{
		Port:      "/dev/ttyUSB0",
		Command:   "*IDN?",
		BaudRate:  115200,
		DataBits:  8,
		Parity:    "none",
		StopBits:  "1",
		TimeoutMs: 1000,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "open port") {
		t.Errorf("error = %q, want to contain 'open port'", err.Error())
	}
}

func TestQueryWriteError(t *testing.T) {
	q := &PortQuerier{
		opener: func(_ string, _ *goserial.Mode) (goserial.Port, error) {
			return &mockPort{writeErr: errors.New("broken pipe")}, nil
		},
	}

	_, err := q.Query(context.Background(), QueryOptions{
		Port:      "/dev/ttyUSB0",
		Command:   "*IDN?",
		BaudRate:  115200,
		DataBits:  8,
		Parity:    "none",
		StopBits:  "1",
		TimeoutMs: 1000,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "write command") {
		t.Errorf("error = %q, want to contain 'write command'", err.Error())
	}
}

func TestQueryNoResponse(t *testing.T) {
	q := &PortQuerier{
		opener: func(_ string, _ *goserial.Mode) (goserial.Port, error) {
			return &mockPort{
				readData: []byte{},
			}, nil
		},
	}

	_, err := q.Query(context.Background(), QueryOptions{
		Port:      "/dev/ttyUSB0",
		Command:   "*IDN?",
		BaudRate:  115200,
		DataBits:  8,
		Parity:    "none",
		StopBits:  "1",
		TimeoutMs: 100,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no response") {
		t.Errorf("error = %q, want to contain 'no response'", err.Error())
	}
}

func TestQueryCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	q := &PortQuerier{
		opener: func(_ string, _ *goserial.Mode) (goserial.Port, error) {
			return &mockPort{readData: []byte("data\n")}, nil
		},
	}

	_, err := q.Query(ctx, QueryOptions{
		Port:      "/dev/ttyUSB0",
		Command:   "*IDN?",
		BaudRate:  115200,
		DataBits:  8,
		Parity:    "none",
		StopBits:  "1",
		TimeoutMs: 1000,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseParity(t *testing.T) {
	tests := []struct {
		input string
		want  goserial.Parity
		err   bool
	}{
		{"none", goserial.NoParity, false},
		{"n", goserial.NoParity, false},
		{"", goserial.NoParity, false},
		{"odd", goserial.OddParity, false},
		{"o", goserial.OddParity, false},
		{"even", goserial.EvenParity, false},
		{"e", goserial.EvenParity, false},
		{"mark", goserial.MarkParity, false},
		{"m", goserial.MarkParity, false},
		{"space", goserial.SpaceParity, false},
		{"s", goserial.SpaceParity, false},
		{"NONE", goserial.NoParity, false},
		{"ODD", goserial.OddParity, false},
		{"invalid", goserial.NoParity, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseParity(tt.input)
			if (err != nil) != tt.err {
				t.Errorf("parseParity(%q) error = %v, wantErr %v", tt.input, err, tt.err)
			}
			if got != tt.want {
				t.Errorf("parseParity(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseStopBits(t *testing.T) {
	tests := []struct {
		input string
		want  goserial.StopBits
		err   bool
	}{
		{"1", goserial.OneStopBit, false},
		{"", goserial.OneStopBit, false},
		{"1.5", goserial.OnePointFiveStopBits, false},
		{"2", goserial.TwoStopBits, false},
		{"3", goserial.OneStopBit, true},
		{"invalid", goserial.OneStopBit, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseStopBits(tt.input)
			if (err != nil) != tt.err {
				t.Errorf("parseStopBits(%q) error = %v, wantErr %v", tt.input, err, tt.err)
			}
			if got != tt.want {
				t.Errorf("parseStopBits(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestQueryInvalidParity(t *testing.T) {
	q := &PortQuerier{
		opener: func(_ string, _ *goserial.Mode) (goserial.Port, error) {
			return &mockPort{readData: []byte("data\n")}, nil
		},
	}

	_, err := q.Query(context.Background(), QueryOptions{
		Port:      "/dev/ttyUSB0",
		Command:   "*IDN?",
		BaudRate:  115200,
		DataBits:  8,
		Parity:    "invalid",
		StopBits:  "1",
		TimeoutMs: 1000,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid parity") {
		t.Errorf("error = %q, want to contain 'invalid parity'", err.Error())
	}
}

func TestQueryInvalidStopBits(t *testing.T) {
	q := &PortQuerier{
		opener: func(_ string, _ *goserial.Mode) (goserial.Port, error) {
			return &mockPort{readData: []byte("data\n")}, nil
		},
	}

	_, err := q.Query(context.Background(), QueryOptions{
		Port:      "/dev/ttyUSB0",
		Command:   "*IDN?",
		BaudRate:  115200,
		DataBits:  8,
		Parity:    "none",
		StopBits:  "3",
		TimeoutMs: 1000,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid stop bits") {
		t.Errorf("error = %q, want to contain 'invalid stop bits'", err.Error())
	}
}
