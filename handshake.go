package websocket

import (
    "bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

var WSErrorKeyNULL = errors.New("WS Error: NULL, Sec-WebSocket-Key")

func Sec_WebSocket_Accept(key string) string {

	const GUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	key = strings.TrimSpace(key)

	var sha1 = sha1.New()

	sha1.Write([]byte(key + GUID))

	key = base64.StdEncoding.EncodeToString(sha1.Sum(nil))

	return key
}

func FormatHttpHeader(p string, s int, h http.Header) *bytes.Buffer {

	var b = new(bytes.Buffer)

	b.WriteString(fmt.Sprintf("HTTP/%s %d %s\r\n", p, s, http.StatusText(s)))

	for i, v := range h {

		for _, vv := range v {

			b.WriteString(fmt.Sprintf("%s: %s\r\n", i, vv))
		}
	}

	b.WriteString("\r\n")

	return b
}

func HandShake(c net.Conn) error {

	var r, e = http.ReadRequest(bufio.NewReader(c))

	if e != nil {

		return e
	}

	var k = r.Header.Get("Sec-WebSocket-Key")

	if k == "" {

		return WSErrorKeyNULL
	}

	k = Sec_WebSocket_Accept(k)

	c.Write(FormatHttpHeader("1.1", 101, http.Header{
		"Upgrade":              []string{"websocket"},
		"Connection":           []string{"Upgrade"},
		"Sec-WebSocket-Accept": []string{k},
	}).Bytes())

	return nil
}
