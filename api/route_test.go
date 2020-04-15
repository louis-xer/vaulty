package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
	"github.com/vaulty/proxy/api/request"
	"github.com/vaulty/proxy/model"
	"github.com/vaulty/proxy/storage/test_storage"
)

func TestHandleRouteCreate(t *testing.T) {
	st := test_storage.NewTestStorage()
	server := NewServer(test_storage.NewTestStorage())
	defer test_storage.Reset()

	vault := &model.Vault{Upstream: "https://example.com"}
	err := st.CreateVault(vault)
	require.NoError(t, err)

	in := new(bytes.Buffer)
	json.NewEncoder(in).Encode(&model.Route{
		Type:     model.RouteInbound,
		Method:   http.MethodPost,
		Path:     "/tokenize",
		Upstream: "https://example.com",
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", in)
	r = r.WithContext(request.WithVault(r.Context(), vault))

	server.HandleRouteCreate()(w, r)

	require.Equal(t, 200, w.Code)

	out := &model.Route{}
	json.NewDecoder(w.Body).Decode(out)

	require.NotEmpty(t, out.ID)
	require.Equal(t, model.RouteInbound, out.Type)
	require.Equal(t, "POST", out.Method)
	require.Equal(t, "/tokenize", out.Path)
	require.Equal(t, "https://example.com", out.Upstream)
	require.Equal(t, vault.ID, out.VaultID)
}

func TestRouteCtx(t *testing.T) {
	st := test_storage.NewTestStorage()
	server := NewServer(st)
	defer test_storage.Reset()

	vault := &model.Vault{Upstream: "https://example.com"}
	err := st.CreateVault(vault)
	require.NoError(t, err)

	route := &model.Route{
		Type:     model.RouteInbound,
		Method:   http.MethodPost,
		Path:     "/tokenize",
		Upstream: "https://example.com",
		VaultID:  vault.ID,
	}
	err = st.CreateRoute(route)
	require.NoError(t, err)

	t.Run("Test route is set into the request context", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		routeCtx := new(chi.Context)
		routeCtx.URLParams.Add("routeID", route.ID)

		ctx := r.Context()
		ctx = request.WithVault(ctx, vault)
		ctx = context.WithValue(ctx, chi.RouteCtxKey, routeCtx)

		r = r.WithContext(ctx)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, route, request.RouteFrom(r.Context()))
		})

		server.RouteCtx(testHandler).ServeHTTP(w, r)
	})

	t.Run("Test route not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		routeCtx := new(chi.Context)
		routeCtx.URLParams.Add("routeID", "xxx")

		ctx := r.Context()
		ctx = request.WithVault(ctx, vault)
		ctx = context.WithValue(ctx, chi.RouteCtxKey, routeCtx)

		r = r.WithContext(ctx)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should never be called")
		})

		server.RouteCtx(testHandler).ServeHTTP(w, r)

		require.Equal(t, 404, w.Code)
	})
}

func TestWithRoute(t *testing.T) {
	st := test_storage.NewTestStorage()
	server := NewServer(test_storage.NewTestStorage())
	defer test_storage.Reset()

	vault := &model.Vault{Upstream: "https://example.com"}
	err := st.CreateVault(vault)
	require.NoError(t, err)

	route := &model.Route{
		Type:     model.RouteInbound,
		Method:   http.MethodPost,
		Path:     "/tokenize",
		Upstream: "https://example.com",
		VaultID:  vault.ID,
	}
	err = st.CreateRoute(route)
	require.NoError(t, err)

	t.Run("Find route", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		ctx := r.Context()
		ctx = request.WithVault(ctx, vault)
		ctx = request.WithRoute(ctx, route)

		r = r.WithContext(ctx)

		server.HandleRouteFind()(w, r)

		require.Equal(t, 200, w.Code)

		out := &model.Route{}
		json.NewDecoder(w.Body).Decode(out)
		require.Equal(t, route.ID, out.ID)
	})

	t.Run("List routes", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		ctx := r.Context()
		ctx = request.WithVault(ctx, vault)

		r = r.WithContext(ctx)

		server.HandleRouteList()(w, r)

		require.Equal(t, 200, w.Code)

		want := []*model.Route{
			route,
		}

		got := []*model.Route{}
		json.NewDecoder(w.Body).Decode(&got)

		require.Equal(t, want, got)

	})
	t.Run("Update route", func(t *testing.T) {})
	t.Run("Delete route", func(t *testing.T) {})
}
