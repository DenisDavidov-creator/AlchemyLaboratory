package server

import (
	alchemyHandler "alla/db-service/internal/alchemy/handler"
	brewingHandler "alla/db-service/internal/brewing/handler"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	alchemyHandler *alchemyHandler.AlchemyHandler
	brewingHandler *brewingHandler.BrewingHandler
}

func NewServer(alchemyHandler *alchemyHandler.AlchemyHandler, brewingHandler *brewingHandler.BrewingHandler) *Server {
	return &Server{
		alchemyHandler: alchemyHandler,
		brewingHandler: brewingHandler,
	}
}

func (s *Server) Run() error {
	r := mux.NewRouter()

	internal := r.PathPrefix("/internal").Subrouter()

	internal.HandleFunc("/ingredients", s.alchemyHandler.BuyNewIngredients).Methods(http.MethodPost)
	internal.HandleFunc("/ingredients", s.alchemyHandler.GetIngredients).Methods(http.MethodGet)
	internal.HandleFunc("/ingredients/{id:[0-9]+}", s.alchemyHandler.BuyExistIngredient).Methods(http.MethodPatch)
	internal.HandleFunc("/recipes", s.alchemyHandler.CreateRecipe).Methods(http.MethodPost)
	internal.HandleFunc("/recipes", s.alchemyHandler.ShowRecipes).Methods(http.MethodGet)

	internal.HandleFunc("/brew", s.brewingHandler.CreateJob).Methods(http.MethodPost)
	internal.HandleFunc("/brew", s.brewingHandler.GetJobByUUID).Methods(http.MethodGet)
	internal.HandleFunc("/brew/status", s.brewingHandler.ChangeStatus).Methods(http.MethodPatch)
	internal.HandleFunc("/brew/status", s.brewingHandler.GetStatus).Methods(http.MethodGet)

	server := http.Server{
		Addr:    "0.0.0.0:8081",
		Handler: r,
	}

	go func() {
		log.Println("Starting server...")
		err := server.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error start server: %v", err)
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	log.Println("Shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	return nil
}
