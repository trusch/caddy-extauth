package main

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		val := r.Header.Get("Authorization")
		if val == "secret" {
			log.Info().Str("ip", r.RemoteAddr).Msg("success")
			w.Header().Add("X-Token", "token")
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Info().Str("ip", r.RemoteAddr).Msg("fail")
		w.WriteHeader(http.StatusUnauthorized)
	})
	http.ListenAndServe(":8000", nil)
}
