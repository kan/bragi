package admin

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/kan/bragi/config"
	"github.com/pkg/errors"
)

//go:embed app/dist/*
var adminFiles embed.FS

type AdminServer struct {
	Config     *config.Config
	ConfigPath string
}

func (a *AdminServer) saveConfig(conf *config.Config) error {
	buf, err := toml.Marshal(conf)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := os.WriteFile(a.ConfigPath, buf, 0644); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (a *AdminServer) Serve() error {
	subfs, err := fs.Sub(adminFiles, "app/dist")
	if err != nil {
		return errors.WithStack(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFileFS(w, r, subfs, "index.html")
			return
		}

		http.FileServerFS(subfs).ServeHTTP(w, r)
	})

	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			if err := json.NewEncoder(w).Encode(a.Config); err != nil {
				http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			}
			return
		case http.MethodPost:
			var conf config.Config
			if err := json.NewDecoder(r.Body).Decode(&conf); err != nil {
				http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			}

			if err := a.saveConfig(&conf); err != nil {
				http.Error(w, "Error save config", http.StatusInternalServerError)
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(a.Config); err != nil {
				http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			}
			return
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{Addr: ":" + a.Config.AdminPort}
	log.Printf("Starting web server on port %s...", a.Config.AdminPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.WithStack(err)
	}

	return nil
}

func LoadServer(conf *config.Config, path string) *AdminServer {
	return &AdminServer{Config: conf, ConfigPath: path}
}
