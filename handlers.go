package antipaste

import (
	"io"
)

var protocolHandlers map[string]ProtocolHandler = make(map[string]ProtocolHandler)

type ProtocolHandler interface {
	Prefix() string
	ReadPaste(url string) (io.ReadCloser, error)
	WritePaste(r io.Reader) (string, error)
}
