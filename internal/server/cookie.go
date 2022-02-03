package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

const (
	tokenHeaderName = "X-Token"
	tokenHeaderAge  = 3000
	secret          = "somesecretstring"
)

func setCookieHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(tokenHeaderName)

		var token string
		if err == http.ErrNoCookie {
			token = ""
		} else if err != nil {
			cancel(w)
		} else {
			token = tokenCookie.Value
		}

		if !validateToken(token) {
			token, err = generateToken()
			if err != nil {
				fmt.Println("problen with token generation", err)
				cancel(w)
			}
			cookie := &http.Cookie{
				Name:   tokenHeaderName,
				Value:  token,
				MaxAge: tokenHeaderAge,
				Path:   "/",
			}
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
		fmt.Println("Cookie don't setup un middelware:", err)
		return "", err
	}
	token := tokenCookie.Value
	data, err := hex.DecodeString(token)
	if err != nil {
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
