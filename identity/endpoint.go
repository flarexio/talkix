package identity

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/flarexio/talkix/config"
	"github.com/flarexio/talkix/user"
)

type DirectUser func(subject string) (*user.UserProfile, *Token, error)

type DirectUserResponse struct {
	User  *user.UserProfile `json:"user"`
	Token *Token            `json:"token"`
}

type Token struct {
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
}

func DirectUserEndpoint(path string, cfg config.IdentityConfig) DirectUser {
	baseURL := cfg.ServerURL

	certFile := filepath.Join(path, "certs", cfg.CertFile)
	keyFile := filepath.Join(path, "certs", cfg.KeyFile)
	caFile := filepath.Join(path, "certs", cfg.CaFile)

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		panic(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return func(subject string) (*user.UserProfile, *Token, error) {
		resource := "users/" + subject

		response, err := client.Get(baseURL + "/" + resource)
		if err != nil {
			return nil, nil, err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			errMsg, err := io.ReadAll(response.Body)
			if err != nil {
				return nil, nil, err
			}
			return nil, nil, errors.New(string(errMsg))
		}

		var resp *DirectUserResponse
		if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
			return nil, nil, err
		}

		return resp.User, resp.Token, nil
	}
}
