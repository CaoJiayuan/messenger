package messenger

import (
	socketio "github.com/googollee/go-socket.io"
	"net/http"
)

var ConnectionHandler = func(s socketio.Conn) error {
	s.SetContext("")
	return nil
}

type Server struct {
	io *socketio.Server
}

func (s *Server) Serve(addr string, path ...string) error {
	s.io.OnConnect("/", ConnectionHandler)
	go s.io.Serve()
	defer s.io.Close()
	var p string
	if len(path) < 1 {
		p = "/socket.io/"
	} else {
		p = path[0]
	}

	http.HandleFunc(p, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", request.Header.Get("Origin"))
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		request.Header.Del("Origin")
		s.io.ServeHTTP(writer, request)
	})

	return http.ListenAndServe(addr, nil)
}

func NewServer() (*Server, error) {
	io, err := socketio.NewServer(nil)

	if err != nil {
		return nil, err
	}

	return &Server{io: io}, nil
}
