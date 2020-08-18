package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/holmes89/lectionary/internal/handlers/rest"
	"github.com/holmes89/lectionary/internal/verses"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

//Default values for application -> move to config?
const (
	defaultPort = ":8080"
	defaultCORS = "*"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	app := NewApp()
	app.Run()
	logrus.WithField("error", <-app.Done()).Error("terminated")
}

// NewApp will create new FX application which houses the server configuration and loading
func NewApp() *fx.App {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return fx.New(
		fx.Provide(
			verses.NewService,
			NewMux,
		),
		fx.Invoke(
			rest.NewVerseHandler,
		),
		fx.Logger(
			logger,
		),
	)
}

// NewMux handler will create new routing layer and base http server
func NewMux(lc fx.Lifecycle) *mux.Router {
	logrus.Info("creating mux")

	router := mux.NewRouter()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{defaultCORS})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"})
	cors := handlers.CORS(originsOk, headersOk, methodsOk)

	router.Use(cors)
	handler := (cors)(router)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logrus.Infof("starting server on %s", defaultPort)
			go func() {
				if err := http.ListenAndServe(defaultPort, handler); err != nil {
					logrus.WithError(err).Fatal("http server failure")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Info("stopping server")
			return errors.New("exited")
		},
	})

	return router
}
