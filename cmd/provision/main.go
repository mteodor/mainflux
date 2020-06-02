package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	mfSDK "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/provision"
	"github.com/mainflux/mainflux/provision/api"
)

const (
	defLogLevel        = "debug"
	defConfigFile      = "config.toml"
	defTLS             = "false"
	defServerCert      = ""
	defServerKey       = ""
	defThingsLocation  = "http://localhost"
	defUsersLocation   = "http://localhost"
	defMQTTURL         = "localhost:1883"
	defHTTPPort        = "8091"
	defMfUser          = "test@example.com"
	defMfPass          = "test"
	defMfAPIKey        = ""
	defMfBSURL         = "http://localhost:8202/things/configs"
	defMfWhiteListURL  = "http://localhost:8202/things/state"
	defMfCertsURL      = "http://localhost/certs"
	defProvisionCerts  = "false"
	defProvisionBS     = "true"
	defBSAutoWhitelist = "true"
	defBSContent       = ""

	defCACerts        = ""
	defCertsCAPrivKey = ""
	defCertsDaysValid = "2400h"
	defCertsRsaBits   = 4096

	envConfigFile = "MF_PROVISION_CONFIG_FILE"
	envLogLevel   = "MF_PROVISION_LOG_LEVEL"
	envHTTPPort   = "MF_PROVISION_HTTP_PORT"
	envTLS        = "MF_PROVISION_ENV_CLIENTS_TLS"

	envServerCert       = "MF_PROVISION_SERVER_CERT"
	envServerKey        = "MF_PROVISION_SERVER_KEY"
	envMQTTURL          = "MF_PROVISION_MQTT_URL"
	envUsersLocation    = "MF_PROVISION_USERS_LOCATION"
	envThingsLocation   = "MF_PROVISION_THINGS_LOCATION"
	envMfUser           = "MF_PROVISION_USER"
	envMfPass           = "MF_PROVISION_PASS"
	envMfAPIKey         = "MF_PROVISION_API_KEY"
	envMfCertsURL       = "MF_PROVISION_CERTS_SVC_URL"
	envProvisionCerts   = "MF_PROVISION_X509_PROVISIONING"
	envMfBSURL          = "MF_PROVISION_BS_SVC_URL"
	envMfBSWhiteListURL = "MF_PROVISION_BS_SVC_WHITELIST_URL"
	envProvisionBS      = "MF_PROVISION_BS_CONFIG_PROVISIONING"
	envBSAutoWhiteList  = "MF_PROVISION_BS_AUTO_WHITELIST"
	envBSContent        = "MF_PROVISION_BS_CONTENT"

	envCertsDaysValid = "MF_PROVISION_CERTS_DAYS_VALID"
	envCertsRsaBits   = "MF_PROVISION_CERTS_RSA_BITS"
	envCertsCA        = "MF_PROVISION_CERTS_CA"
	envCertsCAPrivKey = "MF_PROVISION_CERTS_PRIV_KEY"
)

var (
	errMissingConfigFile        = errors.New("missing config file setting")
	errFailedToLoadConfigFile   = errors.New("failed to load config from file")
	errFailedToGetAutoWhiteList = errors.New("failed to get auto whitelist setting")
	errFailGettingCertSettings  = errors.New("failed to get certificate file setting")
	errFailGettingTLSConf       = errors.New("failed to get TLS setting")
	errFailGettingProvBS        = errors.New("failed to get BS url setting")
	errFailedCertLoading        = errors.New("failed to load certificate")
	errFailedCertDecode         = errors.New("failed to decode certificate")
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}
	logger, err := logger.New(os.Stdout, cfg.Server.LogLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}
	if cfgFromFile, err := loadConfigFromFile(cfg.File); err != nil {
		logger.Warn(fmt.Sprintf("Continue with settings from env, failed to load from: %s: %s", cfg.File, err))
	} else {
		// Merge environment variables and file settings.
		mergeConfigs(&cfgFromFile, &cfg)
		cfg = cfgFromFile
		logger.Info("Continue with settings from file:" + cfg.File)
	}
	tlsCert, x509Cert, err := loadCertificates(cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Problem loading certificates %v", err))
	}

	cfg.Certs.CA = x509Cert
	cfg.Certs.Cert = tlsCert

	SDKCfg := mfSDK.Config{
		BaseURL:           cfg.Server.ThingsLocation,
		BootstrapURL:      cfg.Server.MfBSURL,
		CertsURL:          cfg.Server.MfCertsURL,
		HTTPAdapterPrefix: "http",
		MsgContentType:    "application/json",
		TLSVerification:   cfg.Server.TLS,
	}
	SDK := mfSDK.NewSDK(SDKCfg)

	svc := provision.New(cfg, SDK, logger)
	svc = api.NewLoggingMiddleware(svc, logger)

	errs := make(chan error, 2)

	go startHTTPServer(svc, cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Provision service terminated: %s", err))
}

func startHTTPServer(svc provision.Service, cfg provision.Config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.Server.HTTPPort)
	if cfg.Server.ServerCert != "" || cfg.Server.ServerKey != "" {
		logger.Info(fmt.Sprintf("Provision service started using https on port %s with cert %s key %s",
			cfg.Server.HTTPPort, cfg.Server.ServerCert, cfg.Server.ServerKey))
		errs <- http.ListenAndServeTLS(p, cfg.Server.ServerCert, cfg.Server.ServerKey, api.MakeHandler(svc))
		return
	}
	logger.Info(fmt.Sprintf("Provision service started using http on port %s", cfg.Server.HTTPPort))
	errs <- http.ListenAndServe(p, api.MakeHandler(svc))
}

func loadConfigFromFile(file string) (provision.Config, error) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return provision.Config{}, errors.Wrap(errMissingConfigFile, err)
	}
	c, err := provision.Read(file)
	if err != nil {
		return provision.Config{}, errors.Wrap(errFailedToLoadConfigFile, err)
	}
	return c, nil
}

func loadCertificates(conf provision.Config) (tls.Certificate, *x509.Certificate, error) {
	var tlsCert tls.Certificate
	var caCert *x509.Certificate

	tlsCert, err := tls.LoadX509KeyPair(conf.Certs.CAPath, conf.Certs.CAPrivKeyPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(errFailedCertLoading, err)
	}

	b, err := ioutil.ReadFile(conf.Certs.CAPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(errFailedCertLoading, err)
	}

	block, _ := pem.Decode(b)
	if block == nil {
		log.Fatalf("No PEM data found, failed to decode CA")
	}

	caCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(errFailedCertDecode, err)
	}

	return tlsCert, caCert, nil
}

func loadConfig() (provision.Config, error) {
	tls, err := strconv.ParseBool(mainflux.Env(envTLS, defTLS))
	if err != nil {
		return provision.Config{}, errors.Wrap(errFailGettingTLSConf, err)
	}
	provisionX509, err := strconv.ParseBool(mainflux.Env(envProvisionCerts, defProvisionCerts))
	if err != nil {
		return provision.Config{}, errors.Wrap(errFailGettingCertSettings, err)
	}
	provisionBS, err := strconv.ParseBool(mainflux.Env(envProvisionBS, defProvisionBS))
	if err != nil {
		return provision.Config{}, errors.Wrap(errFailGettingProvBS, fmt.Errorf(" for %s", envProvisionBS))
	}

	autoWhiteList, err := strconv.ParseBool(mainflux.Env(envBSAutoWhiteList, defBSAutoWhitelist))
	if err != nil {
		return provision.Config{}, errors.Wrap(errFailedToGetAutoWhiteList, fmt.Errorf(" for %s", envBSAutoWhiteList))
	}
	if autoWhiteList && !provisionBS {
		return provision.Config{}, errors.New("Can't auto whitelist if auto config save is off")
	}

	cfg := provision.Config{
		Server: provision.ServiceConf{
			LogLevel:       mainflux.Env(envLogLevel, defLogLevel),
			ServerCert:     mainflux.Env(envServerCert, defServerCert),
			ServerKey:      mainflux.Env(envServerKey, defServerKey),
			HTTPPort:       mainflux.Env(envHTTPPort, defHTTPPort),
			MfBSURL:        mainflux.Env(envMfBSURL, defMfBSURL),
			MfWhiteListURL: mainflux.Env(envMfBSWhiteListURL, defMfWhiteListURL),
			MfCertsURL:     mainflux.Env(envMfCertsURL, defMfCertsURL),
			MfUser:         mainflux.Env(envMfUser, defMfUser),
			MfPass:         mainflux.Env(envMfPass, defMfPass),
			MfAPIKey:       mainflux.Env(envMfAPIKey, defMfAPIKey),
			ThingsLocation: mainflux.Env(envThingsLocation, defThingsLocation),
			UsersLocation:  mainflux.Env(envUsersLocation, defUsersLocation),
			TLS:            tls,
		},
		Certs: provision.Certs{
			CAPath: mainflux.Env(envCertsCA, defCACerts),
		},
		Bootstrap: provision.Bootstrap{
			X509Provision: provisionX509,
			Provision:     provisionBS,
			AutoWhiteList: autoWhiteList,
			Content:       mainflux.Env(envBSContent, defBSContent),
		},

		// This is default conf for provision if there is no config file
		Channels: []provision.Channel{
			{
				Name:     "control-channel",
				Metadata: map[string]interface{}{"type": "control"},
			}, {
				Name:     "data-channel",
				Metadata: map[string]interface{}{"type": "data"},
			},
		},
		Things: []provision.Thing{
			{
				Name:     "thing",
				Metadata: map[string]interface{}{"external_id": "xxxxxx"},
			},
		},
	}

	cfg.File = mainflux.Env(envConfigFile, defConfigFile)
	return cfg, nil
}

func mergeConfigs(dst, src interface{}) interface{} {
	d := reflect.ValueOf(dst).Elem()
	s := reflect.ValueOf(src).Elem()

	for i := 0; i < d.NumField(); i++ {
		dField := d.Field(i)
		sField := s.Field(i)
		switch dField.Kind() {
		case reflect.Struct:
			dst := dField.Addr().Interface()
			src := sField.Addr().Interface()
			m := mergeConfigs(dst, src)
			val := reflect.ValueOf(m).Elem().Interface()
			dField.Set(reflect.ValueOf(val))
		case reflect.Slice:
		case reflect.Bool:
			if dField.Interface() == false {
				dField.Set(reflect.ValueOf(sField.Interface()))
			}
		case reflect.String:
			if dField.Interface() == "" {
				dField.Set(reflect.ValueOf(sField.Interface()))
			}
		}
	}
	return dst
}
