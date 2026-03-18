package server

import (
	alchemyHandler "alla/api-service/internal/alchemy/handlers"
	brewingHandler "alla/api-service/internal/brewing/handlers"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	alchemyHandler *alchemyHandler.GuildHandler
	brewingHandler *brewingHandler.BrewingHandler
}

func NewServer(alchemyHandler *alchemyHandler.GuildHandler, brewingHandler *brewingHandler.BrewingHandler) *Server {
	return &Server{
		alchemyHandler: alchemyHandler,
		brewingHandler: brewingHandler,
	}
}

func (s *Server) Run() error {
	r := mux.NewRouter()

	r.HandleFunc("/ingredients", s.alchemyHandler.BuyNewIngredients).Methods(http.MethodPost)
	r.HandleFunc("/ingredients", s.alchemyHandler.ShowIngredients).Methods(http.MethodGet)
	r.HandleFunc("/ingredients/{id}", s.alchemyHandler.BuyExistIngredient).Methods(http.MethodPatch)
	r.HandleFunc("/recipes", s.alchemyHandler.CreateRecipe).Methods(http.MethodPost)
	r.HandleFunc("/recipes", s.alchemyHandler.ShowRecipes).Methods(http.MethodGet)
	r.HandleFunc("/brew", s.brewingHandler.Brew).Methods(http.MethodPost)
	r.HandleFunc("/brew/status", s.brewingHandler.StatusBrew).Methods(http.MethodGet)
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: corsMiddleware(r),
	}

	go func() {
		log.Println("Starting server...")
		err := server.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	log.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	log.Println()
	return nil
}
