package auth

import (
	"net/http"
	"log"
)

type Server struct {
	Port string
	C chan string
}

// TODO should the client be responsible for closing the server?

func (s *Server) Start() {
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		s.C <- r.URL.String()
		w.Write([]byte("The application is now authenticated. You may close this page."))
	})

	defer func() {
		close(s.C)
	}()


	log.Fatalln("server error", http.ListenAndServe(":" + s.Port, nil))
}
