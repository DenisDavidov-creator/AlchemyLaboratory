package server

import (
	"alla/worker-service/internal/boiler-worker/handler"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	handlerBrewing handler.HandlerBrewing
}

func NewServer(handlerBrewing handler.HandlerBrewing) *Server {
	return &Server{
		handlerBrewing: handlerBrewing,
	}
}

func (s *Server) Run() error {
	r := mux.NewRouter()
	internal := r.PathPrefix("/internal").Subrouter()
	internal.HandleFunc("/brew", s.handlerBrewing.StartBrewing).Methods(http.MethodPost)

	server := http.Server{
		Addr:    "0.0.0.0:8082",
		Handler: r,
	}

	log.Println("Starting server...")
	err := server.ListenAndServe()
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
