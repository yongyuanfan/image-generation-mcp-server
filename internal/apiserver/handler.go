package apiserver

import (
	"encoding/json"
	"net/http"

	"github.com/yong/image-generation-mcp-server/internal/common"
	"github.com/yong/image-generation-mcp-server/internal/config"
	"github.com/yong/image-generation-mcp-server/internal/model"
	imagesvc "github.com/yong/image-generation-mcp-server/internal/service/image"
)

type Handler struct {
	service *imagesvc.Service
	config  config.Config
}

func NewHandler(cfg config.Config, service *imagesvc.Service) http.Handler {
	h := &Handler{service: service, config: cfg}
	mux := http.NewServeMux()
	mux.HandleFunc("/images/generations", h.handleTextToImage)
	mux.HandleFunc("/images/edits", h.handleImageToImage)
	mux.HandleFunc("/models", h.handleModels)
	return mux
}

func Healthz(w http.ResponseWriter, _ *http.Request) {
	common.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleTextToImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.GenerateImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	result, err := h.service.TextToImage(r.Context(), req)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	common.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) handleImageToImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.GenerateImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	result, err := h.service.ImageToImage(r.Context(), req)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	common.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	common.WriteJSON(w, http.StatusOK, h.service.Models())
}
