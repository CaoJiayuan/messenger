package messenger

import socketio "github.com/googollee/go-socket.io"

type evHandler func(conn socketio.Conn, msg interface{}) string
