package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	httperror "github.com/kurochkinivan/load_balancer/internal/conroller/http/v1/errors"
	"github.com/kurochkinivan/load_balancer/internal/conroller/http/v1/middleware"
	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/kurochkinivan/load_balancer/internal/usecase"
)

type ClientsUseCase interface {
	Clients(ctx context.Context) ([]*entity.Client, error)
	CreateClient(ctx context.Context, client *entity.Client) error
	UpdateClient(ctx context.Context, client *entity.Client) error
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
	router.GET("/v1/api/clients/", middleware.ErrorMiddlewareParams(h.clients))
	router.POST("/v1/api/clients/", middleware.ErrorMiddlewareParams(h.createClient))
	router.PUT("/v1/api/clients/:ip_address", middleware.ErrorMiddlewareParams(h.updateClient))
	router.DELETE("/v1/api/clients/:ip_address", middleware.ErrorMiddlewareParams(h.deleteClient))
}

func (h *ClientsHandler) clients(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {
	clients, err := h.clientsUseCase.Clients(r.Context())
	if err != nil {
		return httperror.InternalServerError(err, "failed to get all clients")
	}

	err = json.NewEncoder(w).Encode(clients)
	if err != nil {
		return httperror.ErrSerialize(err)
	}

	return nil
}

type createClientRequest struct {
	IPAddress     string `json:"ip_address"`
	Capacity      int32  `json:"capacity"`
	RatePerSecond int32  `json:"rate_per_second"`
}

func (h *ClientsHandler) createClient(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {
	var req createClientRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return httperror.ErrDeserialize(err)
	}

	err = h.clientsUseCase.CreateClient(r.Context(), &entity.Client{
		IPAddress:     req.IPAddress,
		Capacity:      req.Capacity,
		RatePerSecond: req.RatePerSecond,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrClientExists) {
			return httperror.Conflict(err, "client already exists")
		}

		return httperror.InternalServerError(err, "failed to create client")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type updateClientRequest struct {
	Capacity      int32 `json:"capacity"`
	RatePerSecond int32 `json:"rate_per_second"`
}

func (h *ClientsHandler) updateClient(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {
	ipAdress := params.ByName("ip_address")
	if ipAdress == "" {
		return httperror.BadRequest(nil, "ipAdress is required")
	}

	var req updateClientRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return httperror.ErrDeserialize(err)
	}

	err = h.clientsUseCase.UpdateClient(r.Context(), &entity.Client{
		IPAddress:     ipAdress,
		Capacity:      req.Capacity,
		RatePerSecond: req.RatePerSecond,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrClientNotFound) {
			return httperror.NotFound(err, "client not found")
		}

		return httperror.InternalServerError(err, "failed to update client")
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func (h *ClientsHandler) deleteClient(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {
	ipAdress := params.ByName("ip_address")
	if ipAdress == "" {
		return httperror.BadRequest(nil, "ipAdress is required")
	}

	err := h.clientsUseCase.DeleteClient(r.Context(), ipAdress)
	if err != nil {
		if errors.Is(err, usecase.ErrClientNotFound) {
			return httperror.NotFound(err, "client not found")
		}

		return httperror.InternalServerError(err, "failed to delete client")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
