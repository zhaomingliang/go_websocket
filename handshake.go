package websocket

import (
    "bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type Frame struct {
	fin     byte
	rsv1    byte
	rsv2    byte
	rsv3    byte
	opcode  byte
	masked  byte
	u7      uint8
	u16     uint16
	u64     uint64
	masking [4]byte
	memory  []byte
	disk    []*os.File
}

type Service struct {
	Log *log.Logger
}

func (sv *Service) Handshake(conn net.Conn, log *log.Logger) {

	var res http.Response

	log.Printf("%s 开始握手\n", conn.RemoteAddr())

	//默认回复

	res.StatusCode = 101

	res.ProtoMajor = 1
	res.ProtoMinor = 1

	res.Header.Set("Connection", "Upgrade")

	res.Header.Set("Upgrade", "websocket")

	res.Header.Set("Sec-WebSocket-Version", "13")

	//读取客户端请求

	var req *http.Request

	var er error

	if req, er = http.ReadRequest(bufio.NewReader(conn)); er != nil {

		log.Println(er)

		goto fail
	}

	//握手请求的除错

	//除错: URL协议 ws(80) 或 wss(443)

	switch strings.ToLower(req.URL.Scheme) {

	case "ws":

	case "wss":

	default:

		log.Printf("WebSocket: URI 协议名称不规范, 未使用 ws 或 wss (不区分大小写的ASCII值), 为 %s\n", []byte(req.URL.Scheme)[:10])

		res.StatusCode = 400

		goto fail
	}

	//除错: HTTP (版本 >= 1.1)

	if !req.ProtoAtLeast(1, 1) {

		log.Printf("WebSocket: 请求的HTTP版本低于 1.1, 为 %d.%d\n", req.ProtoMajor, req.ProtoMinor)

		res.StatusCode = 400

		goto fail

	} else {

		res.ProtoMajor = req.ProtoMajor

		res.ProtoMinor = req.ProtoMinor
	}

	//除错: 协议升级

	if k := req.Header.Get("Connection"); !(strings.ToLower(k) == "upgrade") {

		log.Print("WebSocket: 用户请求 Request.Header 不规范, Connection 未使用 upgrade (不区分大小写的ASCII值)\n")

		res.StatusCode = 400

		goto fail
	}

	if k := req.Header.Get("Upgrade"); !(strings.ToLower(k) == "websocket") {

		log.Print("WebSocket: 用户请求 Request.Header 不规范, Upgrade 未使用 websocket (不区分大小写的ASCII值)\n")

		res.StatusCode = 400

		goto fail
	}

	//除错: WebSocket (版本 >= 13)

	if k := strings.Join(req.Header["Sec-WebSocket-Version"], ""); !(strings.Contains(k, "13")) {

		log.Print("WebSocket: 用户请求 Request.Header 不规范, Sec-WebSocket-Version 未包含 13\n")

		res.StatusCode = 400

		goto fail

	} else {

		res.Header.Set("Sec-WebSocket-Version", k)
	}

	//除错: Key是否缺失

	if k := req.Header.Get("Sec-WebSocket-Key"); k == "" {

		log.Print("WebSocket: 用户请求 Request.Header 不规范, Sec-WebSocket-Key 缺失\n")

		res.StatusCode = 400

		goto fail

	} else {

		// (客户端 Sec-WebSocket-Key + GUID), SHA1, Base64

		const GUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

		k = strings.TrimSpace(k)

		h := sha1.New()

		h.Write([]byte(k + GUID))

		k = base64.StdEncoding.EncodeToString(h.Sum(nil))

		res.Header.Set("Sec-WebSocket-Accept", k)
	}

fail:

	if er = res.Write(conn); er != nil && er != io.EOF {

		log.Println(er)

	} else if res.StatusCode == 101 {

		log.Printf("%s 握手成功\n", conn.RemoteAddr())

		return
	}

	log.Printf("%s 握手失败, 关闭连接\n", conn.RemoteAddr())

	conn.Close()
}
