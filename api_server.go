// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/pprof"

	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

func corsHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeader(w)
	w.WriteHeader(http.StatusOK)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	WriteHTTPError(w, http.StatusNotFound, elemental.NewError("Not Found", "Unable to find the requested resource", "http", http.StatusNotFound))
}

// an apiServer is the structure serving the api routes.
type apiServer struct {
	config      APIServerConfig
	address     string
	multiplexer *bone.Mux
}

// newAPIServer returns a new apiServer.
func newAPIServer(config APIServerConfig, multiplexer *bone.Mux) *apiServer {

	return &apiServer{
		config:      config,
		multiplexer: multiplexer,
	}
}

// isTLSEnabled checks if the current configuration contains sufficient information
// to estabish of TLS connection.
//
// It basically checks that the TLSCAPath, TLSCertificatePath and TLSKeyPath all correctly defined.
func (a *apiServer) isTLSEnabled() bool {

	return a.config.TLSCAPath != "" && a.config.TLSCertificatePath != "" && a.config.TLSKeyPath != ""
}

// createSecureHTTPServer returns a secure HTTP Server.
//
// It will return an error if any.
func (a *apiServer) createSecureHTTPServer(address string) (*http.Server, error) {

	caCert, err := ioutil.ReadFile(a.config.TLSCAPath)
	if err != nil {
		return nil, err
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  clientCertPool,
	}

	tlsConfig.BuildNameToCertificate()

	return &http.Server{
		Addr:      address,
		TLSConfig: tlsConfig,
	}, nil
}

// createSecureHTTPServer returns a insecure HTTP Server.
//
// It will return an error if any.
func (a *apiServer) createUnsecureHTTPServer(address string) (*http.Server, error) {

	return &http.Server{
		Addr: address,
	}, nil
}

// installRoutes installs all the routes declared in the APIServerConfig.
func (a *apiServer) installRoutes() {

	for _, route := range a.config.Routes {

		if route.Method == http.MethodHead {
			a.multiplexer.Head(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodGet {
			a.multiplexer.Get(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodPost {
			a.multiplexer.Post(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodPut {
			a.multiplexer.Put(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodDelete {
			a.multiplexer.Delete(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodPatch {
			a.multiplexer.Patch(route.Pattern, http.HandlerFunc(route.Handler))
		}

		log.WithFields(log.Fields{
			"pattern": route.Pattern,
			"method":  route.Method,
			"package": "bahamut",
		}).Debug("API route installed.")
	}

	a.multiplexer.Options("*", http.HandlerFunc(corsHandler))

	a.multiplexer.Get("/", http.HandlerFunc(corsHandler))

	a.multiplexer.NotFound(http.HandlerFunc(notFoundHandler))

	if a.config.EnableProfiling {
		a.multiplexer.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		a.multiplexer.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		a.multiplexer.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		a.multiplexer.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		a.multiplexer.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

		log.WithFields(log.Fields{
			"package": "bahamut",
		}).Warn("Profiling routes installed.")
	}

	log.WithFields(log.Fields{
		"routes":  len(a.multiplexer.Routes),
		"package": "bahamut",
	}).Info("All routes installed.")
}

// start starts the apiServer.
func (a *apiServer) start() {

	a.installRoutes()

	if a.isTLSEnabled() {

		if a.config.HealthHandler != nil {

			log.WithFields(log.Fields{
				"address":  a.config.HealthListenAddress,
				"endpoint": a.config.HealthEndpoint,
				"package":  "bahamut",
			}).Info("Creating health check server.")

			srv, err := a.createUnsecureHTTPServer(a.config.HealthListenAddress)
			if err != nil {
				log.WithFields(log.Fields{
					"error":   err,
					"package": "bahamut",
				}).Fatal("unable to create health check server")
			}

			mux := bone.New()
			mux.Get(a.config.HealthEndpoint, a.config.HealthHandler)
			srv.Handler = mux
			go srv.ListenAndServe()
		}

		log.WithFields(log.Fields{
			"address": a.config.ListenAddress,
			"package": "bahamut",
		}).Info("Creating secure http server.")

		server, err := a.createSecureHTTPServer(a.config.ListenAddress)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"package": "bahamut",
			}).Fatal("unable to create secure http server")
		}

		server.Handler = a.multiplexer
		err = server.ListenAndServeTLS(a.config.TLSCertificatePath, a.config.TLSKeyPath)

		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"package": "bahamut",
			}).Fatal("unable to start secure http server")
		}

	} else {

		if a.config.HealthHandler != nil {
			log.WithFields(log.Fields{
				"address":  a.config.ListenAddress,
				"endpoint": a.config.HealthEndpoint,
				"package":  "bahamut",
			}).Info("Registering health check handler.")

			a.multiplexer.Get(a.config.HealthEndpoint, a.config.HealthHandler)
		}

		log.WithFields(log.Fields{
			"address": a.config.ListenAddress,
			"package": "bahamut",
		}).Info("Creating unsecure http server.")

		server, err := a.createUnsecureHTTPServer(a.config.ListenAddress)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"package": "bahamut",
			}).Fatal("unable to create unsecure http server")
		}

		server.Handler = a.multiplexer
		err = server.ListenAndServe()

		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"package": "bahamut",
			}).Fatal("unable to start unsecure http server")
		}
	}
}

// stop stops the apiServer.
//
// In reality right now, it does nothing :).
func (a *apiServer) stop() {

}