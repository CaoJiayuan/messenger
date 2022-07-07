package messenger

import (
	"errors"
	"strings"

	socketio "github.com/googollee/go-socket.io"
)

var (
	EvSubscribe   = "subscribe"
	EvUnsubscribe = "unsubscribe"
	EvBroadcast   = "broadcast"
	EvErr         = "_error"
	MasterChannel = "master"
)

var ConnectionHandler = func(server *Server) func(s socketio.Conn) error {
	return func(s socketio.Conn) error {
		s.SetContext("")
		return nil
	}
}

var SubscribeHandler = func(server *Server) evHandler {
	return func(conn socketio.Conn, msg interface{}) string {
		conn.SetContext(msg)
		return mustOrError(conn, func() error {
			if m, ok := msg.(map[string]interface{}); ok {
				if chs, isArray := m["channels"].([]interface{}); isArray {
					for _, ch := range chs {
						conn.Join(ch.(string))
					}
				} else {
					conn.Join(m["channels"].(string))
				}
			} else {
				return errors.New("invalid payload")
			}

			return nil
		})
	}
}

var BroadcastHandler = func(server *Server) evHandler {
	return func(conn socketio.Conn, msg interface{}) string {
		conn.SetContext(msg)
		return mustOrError(conn, func() error {
			if m, ok := msg.(map[string]interface{}); ok {
				if ch, ok := m["channels"].(string); ok {
					chanEv := strings.Split(ch, "::")
					server.Broadcast(chanEv[1], m["payload"], chanEv[0])
				} else {
					return errors.New("invalid payload")
				}
			} else {
				return errors.New("invalid payload")
			}

			return nil
		})
	}
}

var UnsubscribeHandler = func(server *Server) evHandler {
	return func(conn socketio.Conn, msg interface{}) string {
		conn.SetContext(msg)
		return mustOrError(conn, func() error {
			if m, ok := msg.(map[string]interface{}); ok {
				if chs, isArray := m["channels"].([]interface{}); isArray {
					for _, ch := range chs {
						conn.Leave(ch.(string))
					}
				} else {
					conn.Leave(m["channels"].(string))
				}
			} else {
				return errors.New("invalid payload")
			}

			return nil
		})
	}
}

var DisconnectHandler = func(server *Server) func(s socketio.Conn, reason string) {
	return func(s socketio.Conn, reason string) {
		defer server.Broadcast("logout", s.ID(), MasterChannel)
		s.LeaveAll()
	}
}

func mustOrError(conn socketio.Conn, cb func() error) string {
	err := cb()

	if err != nil {
		conn.Emit("default::"+EvErr, err.Error())
		return err.Error()
	}

	return "ok"
}
