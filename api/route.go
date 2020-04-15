package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/vaulty/proxy/api/request"
	"github.com/vaulty/proxy/model"
)

func (s *Server) HandleRouteCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vault := request.VaultFrom(r.Context())
		route := &model.Route{}

		err := json.NewDecoder(r.Body).Decode(route)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		route.VaultID = vault.ID

		err = s.storage.CreateRoute(route)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(route)
	}
}

func (s *Server) HandleRouteList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vault := request.VaultFrom(r.Context())
		routes, err := s.storage.ListRoutes(vault.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(routes)
	}
}

func (s *Server) RouteCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vault := request.VaultFrom(r.Context())
		routeID := chi.URLParam(r, "routeID")
		route, err := s.storage.FindRouteByID(vault.ID, routeID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := request.WithRoute(r.Context(), route)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) HandleRouteFind() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := request.RouteFrom(r.Context())

		json.NewEncoder(w).Encode(route)
	}
}
