package api

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"pypibot/store"
)

type userResp struct {
	*store.User
	Key string `json:"pub-key"`
}

func writeJson(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Panic(err)
	}
}

func Install(r *http.ServeMux, s *store.Store) {
	r.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		var users []*userResp

		if err := s.ForEachUser(func(key []byte, user *store.User) error {
			u := *user
			users = append(users, &userResp{
				User: &u,
				Key:  hex.EncodeToString(key),
			})
			return nil
		}); err != nil {
			log.Panic(err)
		}

		writeJson(w, users, http.StatusOK)
	})
}
