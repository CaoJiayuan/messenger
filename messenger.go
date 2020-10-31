package messenger

import (
	socketio "github.com/googollee/go-socket.io"
	"net/http"
)

var (
	DefaultNameSpace     = ""
	CORSAllowOrigins     = "*"
	CORSAllowCredentials = "true"
	CORSAllowHeaders     = "x-client-id,authorization"
)

type Server struct {
	io   *socketio.Server
	cors bool
}

func (s *Server) Cors(cors ...bool) *Server {
	if len(cors) > 0 {
		s.cors = cors[0]
	} else {
		s.cors = true
	}
	return s
}

func (s *Server) Serve(addr string, path ...string) error {
	s.RegisterEvents().ServeIo()
	defer s.Close()
	var p string
	if len(path) < 1 {
		p = "/socket.io/"
	} else {
		p = path[0]
	}

	http.Handle(p, s)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) RegisterEvents() *Server {
	s.io.OnConnect(DefaultNameSpace, ConnectionHandler(s))
	s.io.OnEvent(DefaultNameSpace, EvSubscribe, SubscribeHandler(s))
	s.io.OnEvent(DefaultNameSpace, EvUnsubscribe, UnsubscribeHandler(s))
	s.io.OnEvent(DefaultNameSpace, EvBroadcast, BroadcastHandler(s))
	s.io.OnDisconnect(DefaultNameSpace, DisconnectHandler(s))
	return s
}

//ServeIo only start io server
func (s *Server) ServeIo() *Server {
	go s.io.Serve()

	return s
}

//Close io server
func (s *Server) Close() error {
	return s.io.Close()
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	if s.cors {
		if CORSAllowOrigins == "*" {
			writer.Header().Set("Access-Control-Allow-Origin", request.Header.Get("Origin"))
		} else {
			writer.Header().Set("Access-Control-Allow-Origin", CORSAllowOrigins)
		}
		writer.Header().Set("Access-Control-Allow-Credentials", CORSAllowCredentials)
		request.Header.Del("Origin")
		writer.Header().Set("Access-Control-Allow-Headers", CORSAllowHeaders)
		if request.Method == http.MethodOptions {
			writer.WriteHeader(204)
			return
		}
	}

	s.io.ServeHTTP(writer, request)
}

func (s *Server) Broadcast(ev string, payload interface{}, channels ...string) {

	if len(channels) < 1 || hasWildcard(channels) {
		s.io.BroadcastToRoom(DefaultNameSpace, "", "*::"+ev, payload)
	} else {
		for _, ch := range channels {
			s.io.BroadcastToRoom(DefaultNameSpace, ch, ch+"::"+ev, payload)
		}
	}
}

func hasWildcard(channels []string) bool {
	for _, ch := range channels {
		if ch == "*" {
			return true
		}
	}
	return false
}

func NewServer() (*Server, error) {
	io, err := socketio.NewServer(nil)

	if err != nil {
		return nil, err
	}

	return &Server{io: io}, nil
}
