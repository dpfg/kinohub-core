package cmd

import (
	"fmt"
	"net/http"

	"github.com/dpfg/kinohub-core/providers"
	"github.com/dpfg/kinohub-core/providers/trakt"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"golang.org/x/oauth2"
)

// ServerCommand with available flags and env. variables
type ServerCommand struct {
	Port         int    `long:"port" env:"KINOHUB_PORT" default:"8090" description:"port"`
	SiteName     string `long:"site-name" default:"localhost" description:"Site name used by 3rd parties"`
	DataLocation string `long:"data-location" env:"KINOHUB_DATA_LOCATION" default:".data/" description:"path to folder to store application data"`
	Auth         struct {
		Trakt   AuthGroup `group:"trakt" namespace:"trakt" env-namespace:"TRAKT" description:"Trakt OAuth"`
		TMBD    AuthGroup `group:"tmdb" namespace:"tmdb" env-namespace:"TMDB" description:"TMDB OAuth"`
		KinoPub AuthGroup `group:"kinopub" namespace:"kinopub" env-namespace:"KINO_PUB" description:"KinoPub OAuth"`
	}
}

// AuthGroup defines options group for auth params
type AuthGroup struct {
	CID  string `long:"cid" env:"CID" description:"OAuth client ID"`
	CSEC string `long:"csec" env:"CSEC" description:"OAuth client secret"`
}

// Execute starts web server on specified port. Called by flags parser
func (cmd *ServerCommand) Execute(args []string) error {

	logger := newLogger()

	server := Server{
		port:   cmd.Port,
		logger: logger,
		trakt:  cmd.makeTraktIntegration(logger),
	}

	server.serve()

	return nil
}

func (cmd *ServerCommand) makeTraktIntegration(logger *logrus.Logger) *trakt.Integration {
	return &trakt.Integration{Client: &trakt.Client{
		Config: oauth2.Config{
			ClientID:     cmd.Auth.Trakt.CID,
			ClientSecret: cmd.Auth.Trakt.CSEC,
			Scopes:       []string{},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://api.trakt.tv/oauth/authorize",
				TokenURL: "https://api.trakt.tv/oauth/token",
			},
			RedirectURL: fmt.Sprintf("http://%s:%d/trakt/exchange", cmd.SiteName, cmd.Port),
		},
		PreferenceStorage: providers.JSONPreferenceStorage{
			Path: cmd.DataLocation,
		},
		Logger: logger.WithFields(logrus.Fields{"prefix": "trakt"}),
	}}
}

// Server with all available dependencies
type Server struct {
	port   int
	logger *logrus.Logger

	trakt *trakt.Integration
}

func (server *Server) serve() {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	router.Mount("/trakt", server.trakt.Handler())

	server.logger.Infof("Starting KinuHub server on localhost:%d", server.port)
	http.ListenAndServe(fmt.Sprintf(":%d", server.port), router)
}

func newLogger() *logrus.Logger {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)

	logfmt := new(prefixed.TextFormatter)
	// logfmt.DisableColors = true
	logfmt.FullTimestamp = true
	logfmt.TimestampFormat = "2006/01/02 15:04:05"

	logger.Formatter = logfmt
	return logger
}
