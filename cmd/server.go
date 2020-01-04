package cmd

import (
	"fmt"
	"net/http"

	provider "github.com/dpfg/kinohub-core/provider"
	"github.com/dpfg/kinohub-core/provider/kinopub"
	"github.com/dpfg/kinohub-core/provider/tmdb"
	"github.com/dpfg/kinohub-core/provider/trakt"
	"github.com/dpfg/kinohub-core/services"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
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
		Trakt   OAuthGroup  `group:"trakt" namespace:"trakt" env-namespace:"TRAKT" description:"Trakt OAuth"`
		TMBD    APIKeyGroup `group:"tmdb" namespace:"tmdb" env-namespace:"TMDB" description:"TMDB API Auth"`
		KinoPub OAuthGroup  `group:"kinopub" namespace:"kinopub" env-namespace:"KINOPUB" description:"KinoPub OAuth"`
	} `group:"auth" namespace:"auth" env-namespace:"AUTH"`
}

// OAuthGroup defines options group for oauth params
type OAuthGroup struct {
	CID  string `long:"cid" env:"CID" description:"OAuth client ID"`
	CSEC string `long:"csec" env:"CSEC" description:"OAuth client secret"`
}

// APIKeyGroup defines auth options that reliy on a single API Key.
type APIKeyGroup struct {
	Key string `long:"key" env:"KEY" description:"API key"`
}

// Execute starts web server on specified port. Called by flags parser
func (cmd *ServerCommand) Execute(args []string) error {
	logger := newLogger()

	logger.Debugf("%v", cmd)

	cacheFactory, err := cmd.makeCacheFactory(logger)
	if err != nil {
		return fmt.Errorf("Cannot initialize cache factory. %s", err.Error())
	}

	tmdbc := cmd.makeTMDBClient(logger, cacheFactory)
	kpc := cmd.makeKinoPubClient(cacheFactory, logger)
	trakt := cmd.makeTraktIntegration(logger)

	server := Server{
		port:         cmd.Port,
		logger:       logger,
		cacheFactory: cacheFactory,
		trakt:        trakt,
		tmdb:         tmdbc,
		kinopub:      kpc,
		search:       cmd.makeContentSearch(kpc, tmdbc, logger),
		feedService:  cmd.makeFeed(trakt.Client, kpc, tmdbc, logger),
		infoService:  cmd.makeContentBrowser(kpc, tmdbc, logger),
	}

	server.serve()

	return nil
}

func (cmd *ServerCommand) makeCacheFactory(logger *logrus.Logger) (provider.CacheFactory, error) {
	return provider.NewCacheFactory(cmd.DataLocation, logger)
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
		PreferenceStorage: provider.JSONPreferenceStorage{
			Path: cmd.DataLocation,
		},
		Logger: logger.WithField("prefix", "trakt"),
	}}
}

func (cmd *ServerCommand) makeContentSearch(kpc kinopub.KinoPubClient, tmdbc tmdb.Client, logger *logrus.Logger) *services.ContentSearch {
	return &services.ContentSearch{
		Kinopub: kpc,
		TMDB:    tmdbc,
		Logger:  logger.WithField("prefix", "content-search"),
	}
}

func (cmd *ServerCommand) makeFeed(trakt *trakt.Client, kinopub kinopub.KinoPubClient, tmdbc tmdb.Client, logger *logrus.Logger) services.Feed {
	return services.NewFeed(trakt, kinopub, tmdbc, logger.WithField("prefix", "feed"))
}

func (cmd *ServerCommand) makeContentBrowser(kinopub kinopub.KinoPubClient, tmdbc tmdb.Client, logger *logrus.Logger) services.ContentBrowser {
	return services.NewContentBrowser(kinopub, tmdbc)
}

func (cmd *ServerCommand) makeKinoPubClient(cf provider.CacheFactory, logger *logrus.Logger) kinopub.KinoPubClient {
	return kinopub.KinoPubClientImpl{
		ClientID:     cmd.Auth.KinoPub.CID,
		ClientSecret: cmd.Auth.KinoPub.CSEC,
		PreferenceStorage: provider.JSONPreferenceStorage{
			Path: cmd.DataLocation,
		},
		CacheFactory: cf,
		Logger:       logger.WithField("prefix", "kinopub"),
	}
}

func (cmd *ServerCommand) makeTMDBClient(logger *logrus.Logger, cf provider.CacheFactory) tmdb.Client {
	return tmdb.ClientImpl{
		APIKey: cmd.Auth.TMBD.Key,
		PreferenceStorage: provider.JSONPreferenceStorage{
			Path: cmd.DataLocation,
		},
		Cache:  cf,
		Logger: logger.WithField("prefix", "tmdb"),
	}
}

// Server with all available dependencies
type Server struct {
	port   int
	logger *logrus.Logger

	cacheFactory provider.CacheFactory
	trakt        *trakt.Integration
	kinopub      kinopub.KinoPubClient
	tmdb         tmdb.Client
	search       *services.ContentSearch
	infoService  services.ContentBrowser
	feedService  services.Feed
}

func (server *Server) serve() {
	router := chi.NewRouter()

	router.Use(cors.AllowAll().Handler)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("kinohub v0.0.2"))
	})

	router.Mount("/trakt", server.trakt.Handler())
	router.Mount("/search", server.search.Handler())

	router.Group(server.infoService.Handler())
	router.Group(server.feedService.Handler())

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
