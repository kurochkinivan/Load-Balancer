package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/kurochkinivan/load_balancer/internal/usecase"
)

type ClientsUseCase interface {
	Clients(ctx context.Context) ([]*entity.Client, error)
	CreateClient(ctx context.Context, client *entity.Client) error
	DeleteClient(ctx context.Context, ipAdress string) error
}

type ClientsHandler struct {
	clientsUseCase ClientsUseCase
	bytesLimit     int64
}

func NewClientsHandler(clientsUseCase ClientsUseCase, bytesLimit int64) *ClientsHandler {
	return &ClientsHandler{
		clientsUseCase: clientsUseCase,
		bytesLimit:     bytesLimit,
	}
}

func (h *ClientsHandler) Register(router *httprouter.Router) {
	router.GET("/v1/api/clients/", h.clients)
	router.POST("/v1/api/clients/", h.createClient)
	router.DELETE("/v1/api/clients/:ip_address", h.deleteClient)
}

func (h *ClientsHandler) clients(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clients, err := h.clientsUseCase.Clients(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(clients)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type createClientRequest struct {
	IPAddress     string `json:"ip_address"`
	Capacity      int32  `json:"capacity"`
	RatePerSecond int32  `json:"rate_per_second"`
}

func (h *ClientsHandler) createClient(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var req createClientRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.clientsUseCase.CreateClient(r.Context(), &entity.Client{
		IPAddress:     req.IPAddress,
		Capacity:      req.Capacity,
		RatePerSecond: req.RatePerSecond,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrClientExists) {
			http.Error(w, "client already exists", http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *ClientsHandler) deleteClient(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	ipAdress := params.ByName("ip_address")
	if ipAdress == "" {
		http.Error(w, "ipAdress is required", http.StatusBadRequest)
		return
	}

	err := h.clientsUseCase.DeleteClient(r.Context(), ipAdress)
	if err != nil {
		if errors.Is(err, usecase.ErrClientNotFound) {
			http.Error(w, "client not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
