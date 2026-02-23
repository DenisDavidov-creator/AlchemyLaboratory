package server

import (
	"alchemicallabaratory/handlers"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	handler *handlers.GuildHandler
}

func NewServer(handler *handlers.GuildHandler) *Server {
	return &Server{
		handler: handler,
	}
}

func (s *Server) Run() error {
	r := mux.NewRouter()

	r.HandleFunc("/ingredients", s.handler.BuyNewIngredients).Methods(http.MethodPost)
	r.HandleFunc("/ingredients", s.handler.ShowIngredients).Methods(http.MethodGet)
	r.HandleFunc("/ingredients/{id}", s.handler.BuyExistIngredient).Methods(http.MethodPatch)
	r.HandleFunc("/recipes", s.handler.CreateRecipe).Methods(http.MethodPost)
	r.HandleFunc("/recipes", s.handler.ShowRecipes).Methods(http.MethodGet)
	r.HandleFunc("/brew", s.handler.Brew).Methods(http.MethodPost)
	r.HandleFunc("/brew/status", s.handler.StatusBrew).Methods(http.MethodGet)

	server := http.Server{
		Addr:    ":9090",
		Handler: r,
	}

	log.Println("Starting server...")
	err := server.ListenAndServe()
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
