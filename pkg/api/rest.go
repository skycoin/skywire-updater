package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/skycoin/skywire-updater/pkg/update"
)

func handleREST(g Gateway) http.Handler {
	r := chi.NewRouter()
	r.Get("/services", services(g))
	r.Get("/services/{srv}/check", checkService(g))
	r.Post("/services/{srv}/update/{ver}", updateService(g))
	return r
}

func services(g Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, g.Services())
	}
}

func checkService(g Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			pSrv = chi.URLParam(r, "srv")
		)
		release, err := g.Check(r.Context(), pSrv)
		if err != nil {
			if err == update.ErrServiceNotFound {
				writeJSON(w, http.StatusNotFound, err)
				return
			}
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, release)
	}
}

func updateService(g Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			pSrv = chi.URLParam(r, "srv")
			pVer = chi.URLParam(r, "ver")
		)
		ok, err := g.Update(r.Context(), pSrv, pVer)
		if err != nil {
			if err == update.ErrServiceNotFound {
				writeJSON(w, http.StatusNotFound, err)
				return
			}
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}
		if !ok {
			writeJSON(w, http.StatusInternalServerError, errors.New("update failed"))
			return
		}
		writeJSON(w, http.StatusOK, ok)
	}
}

// writes a json object on a http.ResponseWriter with the given code,
// panics on marshaling error.
func writeJSON(w http.ResponseWriter, code int, v interface{}) {

	// HTTPError is included in an HTTPResponse
	type HTTPError struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	// HTTPResponse represents the http response struct
	type HTTPResponse struct {
		Error *HTTPError  `json:"error,omitempty"`
		Data  interface{} `json:"data,omitempty"`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err, ok := v.(error); ok {
		v = HTTPResponse{
			Error: &HTTPError{
				Message: err.Error(),
				Code:    code,
			},
		}
	} else {
		v = HTTPResponse{
			Data: v,
		}
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.WithError(err).Fatal()
	}
}
