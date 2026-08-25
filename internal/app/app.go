package app

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/kyotomin/ab-router/internal/config"
	"github.com/kyotomin/ab-router/internal/storage"
)

type App struct {
	routerServer *router.Server
	adminServer  *admin.Server
	store        storage.Storage
}

func New(cfg *config.Config) *App {
	var store storage.Storage
	var err error

	connString := cfg.GetDBConnString()
	migrateString := cfg.GetMigrationsConnString()

	store, err = storage.NewPostgresStorage(
		connString,
		migrateString,
		cfg.DB.PingRetries,
	)

	if err != nil {
		slog.Info(
			"error connecting to postgres",
			"error", err,
			"Using MemoryStorage (data will not store)",
		)
		store = storage.NewMemoryStorage()
	} else {
		slog.Info("Connected to PostgreSQL")
	}

	routerServer := router.NewServer()
	adminServer := admin.NewServer(store)

	return &App{
		routerServer: routerServer,
		adminServer:  adminServer,
		store:        store,
	}
}

func (a *App) Run(cfg *config.Config) error {
	routerPort := cfg.App.RouterPort
	adminPort := cfg.App.AdminPort

	go func() {
		log.Printf("Router launched on :%d", routerPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", routerPort), a.routerServer.Handler()); err != nil {
			log.Printf("router err: %v", err)
		}
	}()

	go func() {
		log.Printf("Admin panel launched on :%d", adminPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", adminPort), a.adminServer.Handler()); err != nil {
			log.Printf("admin error: %v", err)
		}
	}()

	log.Println("App started correctly")
	select {}
}
