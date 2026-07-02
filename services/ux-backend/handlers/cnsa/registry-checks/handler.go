package registrychecks

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	"github.com/red-hat-storage/odf-operator/services/ux-backend/handlers"
	"github.com/red-hat-storage/odf-operator/services/ux-backend/util"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RegistryCheckRequest struct {
	RegistryURL            string `json:"registryURL"`
	RegistryRepositoryName string `json:"registryRepositoryName"`
	SecretKey              string `json:"secretKey"`
	SecretNamespace        string `json:"secretNamespace"`
}

type RegistryCheckResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func HandleMessage(w http.ResponseWriter, r *http.Request, client client.Client) {
	switch r.Method {
	case "POST":
		handleInput(w, r, client)
	default:
		handleUnsupportedMethod(w, r)
	}
}

func handleInput(w http.ResponseWriter, r *http.Request, client client.Client) {
	var req RegistryCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorf("failed to decode request body: %v", err)
		response := RegistryCheckResponse{
			Success: false,
			Message: "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			klog.Errorf("failed to encode response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}
	err := util.TestRegistryConnection(req.RegistryURL, req.RegistryRepositoryName, req.SecretKey, req.SecretNamespace, client)
	if err != nil {
		klog.Errorf("failed to test registry connection: %v", err)
		response := RegistryCheckResponse{
			Success: false,
			Message: fmt.Sprintf("Registry connection failed: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			klog.Errorf("failed to encode response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}
	response := RegistryCheckResponse{
		Success: true,
		Message: "Registry connection successful",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		klog.Errorf("failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func handleUnsupportedMethod(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Only POST method should be used to send data to this endpoint %s", r.URL.Path)
	w.Header().Set("Content-Type", handlers.ContentTypeTextPlain)
	w.Header().Set("Allow", "POST")
	w.WriteHeader(http.StatusMethodNotAllowed)

	if _, err := fmt.Fprintf(w, "Unsupported method : %s", html.EscapeString(r.Method)); err != nil {
		klog.Errorf("failed write data to response writer: %v", err)
	}
}
