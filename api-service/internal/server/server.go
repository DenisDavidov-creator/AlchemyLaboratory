package server

import (
	alchemyHandler "alla/api-service/internal/alchemy/handlers"
	brewingHandler "alla/api-service/internal/brewing/handlers"
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	logger         *slog.Logger
}

func NewServer(alchemyHandler *alchemyHandler.GuildHandler, brewingHandler *brewingHandler.BrewingHandler, logger *slog.Logger) *Server {
	return &Server{
		alchemyHandler: alchemyHandler,
		brewingHandler: brewingHandler,
		logger:         logger,
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
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

	r.Use(s.loggingMiddleware)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: corsMiddleware(r),
	}

	go func() {
		s.logger.Info("Starting server...")
		err := server.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {

			s.logger.Error("Start server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	s.logger.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	return nil
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		s.logger.Info("HTTP request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
