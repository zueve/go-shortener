package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/zueve/go-shortener/pkg/logging"
)

const (
	tokenHeaderName = "X-Token"
	tokenHeaderAge  = 3000
	secret          = "somesecretstring"
)

func setCookieHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(tokenHeaderName)
		logger := log(r.Context())
		var token string
		if err == http.ErrNoCookie {
			token = ""
		} else if err != nil {
			logger.Error().Err(err).Msg("Can't read cookie")
			cancel(w)
		} else {
			token = tokenCookie.Value
		}

		if !validateToken(token) {
			token, err = generateToken()
			if err != nil {
				logger.Error().Err(err).Msg("problen with token generation")
				cancel(w)
			}
			cookie := &http.Cookie{
				Name:   tokenHeaderName,
				Value:  token,
				MaxAge: tokenHeaderAge,
				Path:   "/",
			}
			logger.Info().Msgf("Set token %s:%s", cookie.Name, cookie.Value)
			http.SetCookie(w, cookie)
			r.AddCookie(cookie)
		}

		next.ServeHTTP(w, r)
	})
}

func validateToken(token string) bool {
	data, err := hex.DecodeString(token)
	if err != nil || len(data) < 32 {
		return false
	}
	id := data[:16]
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(id)
	sign := h.Sum(nil)

	return hmac.Equal(sign, data[16:])
}

func generateToken() (string, error) {
	id, err := uuid.New().MarshalBinary()
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(id)
	sign := h.Sum(nil)
	token := append(id, sign...)

	return hex.EncodeToString(token), nil
}

func getUserID(r *http.Request) (string, error) {
	tokenCookie, err := r.Cookie(tokenHeaderName)
	if err != nil {
		log(r.Context()).Error().Err(err)
		return "", err
	}
	token := tokenCookie.Value
	data, err := hex.DecodeString(token)
	if err != nil {
		log(r.Context()).Error().Err(err)
		return "", err
	}
	id := data[:16]
	return hex.EncodeToString(id), nil
}

func cancel(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("content-type", "plain/text")
	w.Write([]byte("internal server error"))
}

func log(ctx context.Context) *zerolog.Logger {
	_, logger := logging.GetCtxLogger(ctx)
	logger = logger.With().
		Str(logging.Source, "setCookieHandler").
		Str(logging.Layer, "api").
		Logger()

	return &logger
}
