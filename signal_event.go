package websocket

import "net"

type Event struct {
    OnOpen    func(net.Conn)
	OnMessage func(Packet)
	OnClose   func(net.Conn)
	OnError   func(string)
}

type Signal struct {
	OnOpen    chan net.Conn
	OnMessage chan Packet
	OnClose   chan net.Conn
	OnError   chan string
	Close     chan string
}

func EventListen(e Event) Signal {
	var s = Signal{
		OnOpen:    make(chan net.Conn),
		OnMessage: make(chan Packet),
		OnClose:   make(chan net.Conn),
		OnError:   make(chan string),
		Close:     make(chan string),
	}
	go func() {
		for {
			select {
			case E := <-s.OnOpen:
				e.OnOpen(E)
			case E := <-s.OnMessage:
				e.OnMessage(E)
			case E := <-s.OnClose:
				e.OnClose(E)
			case E := <-s.OnError:
				e.OnError(E)
			case <-s.Close:
				return
			}
		}
	}()
	return s
}
