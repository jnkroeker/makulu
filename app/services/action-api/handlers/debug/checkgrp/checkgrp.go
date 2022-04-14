// liveness and readiness endpoints for our own debugging
package checkgrp

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/jnkroeker/makulu/business/data"
	"go.uber.org/zap"
)

type Handlers struct {
	Build     string
	GqlConfig data.GraphQLConfig
	Log       *zap.SugaredLogger
}

func (h Handlers) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status := "ok"
	statusCode := http.StatusOK
	err := data.Validate(ctx, h.GqlConfig.URL, 5*time.Second)
	if err != nil {
		status = "db not ready"
		statusCode = http.StatusInternalServerError
	}

	data := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	if err := response(w, statusCode, data); err != nil {
		h.Log.Errorw("readiness", "ERROR", err)
	}

	h.Log.Infow("readiness", "statusCode", statusCode, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)

}

func (h Handlers) Liveness(w http.ResponseWriter, r *http.Request) {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	// Pod, PodIP, Node, Namespace are env variables that are set up in base-action.yaml
	data := struct {
		Status    string `json:"status,omitempty"`
		Build     string `json:"build,omitempty"`
		Host      string `json:"host,omitempty"`
		Pod       string `json:"pod,omitempty"`
		PodIP     string `json:"podIP,omitempty"`
		Node      string `json:"node,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Status:    "up",
		Build:     h.Build,
		Host:      host,
		Pod:       os.Getenv("KUBERNETES_PODNAME"),
		PodIP:     os.Getenv("KUBERNETES_NAMESPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	statusCode := http.StatusOK
	if err := response(w, statusCode, data); err != nil {
		h.Log.Errorw("liveness", "ERROR", err)
	}

	h.Log.Infow("liveness", "statusCode", statusCode, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)
}

func response(w http.ResponseWriter, statusCode int, data interface{}) error {
	// Convert the response value to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
