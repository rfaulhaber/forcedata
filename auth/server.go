package auth

import (
	"net/http"
	"log"
)

type Server struct {
	Port string
	C chan string

	srv *http.Server
}

func (s *Server) Start() {
	srv := &http.Server{Addr: s.Port}
	s.srv = srv

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		s.C <- r.URL.String()
		w.Write([]byte("You are now connected. You may close this page."))
	})

	srv.ListenAndServe()

	log.Fatalln("server error", http.ListenAndServe(":" + s.Port, nil))
}

func (s *Server) Close() {
	err := s.srv.Close()

	if err != nil {
		log.Println("server close error was not nil, panicking")
		panic(err)
	}
}
