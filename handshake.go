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

var WSErrOR = errors.New("websocket error: index out of range(0 - 127)")

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

func (sv *Service) Connect(conn net.Conn) {

	var res http.Response

	var req *http.Request

	var er error

	sv.Log.Printf("%s 开始握手\n", conn.RemoteAddr())

	// 客户端握手请求的除错, 对于违反规则的请求, 关闭链接, 记录错误, 返回 http.Response 告知出错

	if req, er = http.ReadRequest(bufio.NewReader(conn)); er != nil {

		sv.Log.Println(er)

		goto fail
	}

	// HTTP 版本 应 >= 1.1

	if !(req.ProtoMajor >= 1 && req.ProtoMinor >= 1) {

		sv.Log.Printf("WebSocket: 请求的HTTP版本低于 1.1, 为 %d.%d\n", req.ProtoMajor, req.ProtoMinor)

		goto fail

	} else {

		res.ProtoMajor = req.ProtoMajor

		res.ProtoMinor = req.ProtoMinor
	}

	// URL 包含的协议 应为 ws 或者 wss

	switch strings.ToLower(req.URL.Scheme) {

	case "ws":

	case "wss":

	default:

		sv.Log.Printf("WebSocket: URI 协议名称不规范, 未使用 ws 或 wss (不区分大小写的ASCII值), 为 ％s\n", []byte(req.URL.Scheme)[:10])

		goto fail
	}

	if k := req.Header.Get("Connection"); !(strings.ToLower(k) == "upgrade") {

		sv.Log.Print("WebSocket: 用户请求 Request.Header 不规范, Connection 未使用 upgrade (不区分大小写的ASCII值)\n")

		goto fail

	} else {

		res.Header.Set("Connection", "Upgrade")
	}

	if k := req.Header.Get("Upgrade"); !(strings.ToLower(k) == "websocket") {

		sv.Log.Print("WebSocket: 用户请求 Request.Header 不规范, Upgrade 未使用 websocket (不区分大小写的ASCII值)\n")

		goto fail

	} else {

		res.Header.Set("Upgrade", "websocket")
	}

	if k := req.Header.Get("Sec-WebSocket-Version"); !(strings.Contains(k, "13")) {

		sv.Log.Print("WebSocket: 用户请求 Request.Header 不规范, Sec-WebSocket-Version 未包含 13\n")

		goto fail

	} else {

		res.Header.Set("Sec-WebSocket-Version", k)
	}

	if k := req.Header.Get("Sec-WebSocket-Key"); k == "" {

		sv.Log.Print("WebSocket: 用户请求 Request.Header 不规范, Sec-WebSocket-Key 缺失\n")

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

	if er = res.Write(conn); er != nil && er != io.EOF {

		sv.Log.Println(er)

		goto fail

	} else {

		sv.Log.Printf("%s 握手成功\n", conn.RemoteAddr())

		return
	}

fail:

	sv.Log.Printf("%s 握手失败, 关闭连接\n", conn.RemoteAddr())

	conn.Close()
}
