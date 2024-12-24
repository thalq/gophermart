package orders

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/thalq/gopher_mart/internal/constants"
	"github.com/thalq/gopher_mart/internal/errors"
	logger "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/internal/models"
)

type OrderHandler struct {
	service              *OrderService
	AccrualSystemAddress string
}

func fetchAccrualInfo(orderNumber string, AccrualSystemAddress string) (models.AccrualInfo, error) {
	url := AccrualSystemAddress + "/api/orders/" + orderNumber
	resp, err := http.Get(url)
	if err != nil {
		logger.Sugar.Errorf("Failed to send request to accrual system: %v", err)
		return models.AccrualInfo{}, err
	}
	logger.Sugar.Infof("Got response from accrual system: %v", resp)

	defer resp.Body.Close()

	var accrualInfo models.AccrualInfo
	if resp.StatusCode == http.StatusOK {
		logger.Sugar.Infof("Order %s was successfully accrued", orderNumber)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Sugar.Errorf("Failed to read response body: %v", err)
			return models.AccrualInfo{}, err
		}
		if err := json.Unmarshal(body, &accrualInfo); err != nil {
			logger.Sugar.Errorf("Failed to unmarshal response: %v", err)
			return models.AccrualInfo{}, err
		}
		logger.Sugar.Infof("Got accrual info: %v", accrualInfo)
	} else if resp.StatusCode == http.StatusNoContent {
		accrualInfo.SetDefaults(orderNumber)
		logger.Sugar.Infof("Order %s not found", orderNumber)
	} else if resp.StatusCode == http.StatusTooManyRequests {
		logger.Sugar.Infof("Too many requests to accrual system")
		return models.AccrualInfo{}, errors.ErrTooManyRequests
	} else if resp.StatusCode == http.StatusInternalServerError {
		logger.Sugar.Infof("Internal server error in accrual system")
		return models.AccrualInfo{}, errors.ErrInternalServer
	}
	logger.Sugar.Infof("Order %s was successfully accrued", orderNumber)
	return accrualInfo, nil
}

func NewOrderHandler(service *OrderService, AccrualSystemAddress string) *OrderHandler {
	return &OrderHandler{service: service, AccrualSystemAddress: AccrualSystemAddress}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	orderNumber := strings.TrimSpace(string(body))

	if !ValidateOrderNumber(orderNumber) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}
	userHasOrder, err := h.service.CheckUserHasOrders(userID, orderNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if userHasOrder {
		w.WriteHeader(http.StatusOK)
		logger.Sugar.Infof("User %d has order %s", userID, orderNumber)
	} else {
		otherUserHasOrder, err := h.service.CheckOtherUserHasOrders(orderNumber)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if otherUserHasOrder {
			http.Error(w, "Order already exists", http.StatusConflict)
			logger.Sugar.Infof("Order %s already exists for another user", orderNumber)
		} else {
			accrualInfoChan := make(chan models.AccrualInfo)
			go func(orderNumber string) {
				accrualInfo, err := fetchAccrualInfo(orderNumber, h.AccrualSystemAddress)
				if err != nil {
					close(accrualInfoChan)
					return
				}
				accrualInfoChan <- accrualInfo
				close(accrualInfoChan)
			}(orderNumber)
			accrualInfo, ok := <-accrualInfoChan
			if !ok {
				http.Error(w, "Failed to get accrual info", http.StatusInternalServerError)
				return
			}

			if err := h.service.CreateOrder(userID, orderNumber, accrualInfo); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusAccepted)
			logger.Sugar.Infof("User %d created order %s", userID, orderNumber)

		}
	}
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.service.GetOrders(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		http.Error(w, "No orders for user", http.StatusNoContent)
		return
	}
	logger.Sugar.Infof("Got %d orders for user", len(orders))
	response, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (h *OrderHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.service.GetBalance(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Sugar.Infof("Got balance for user %d", userID)
	response, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (h *OrderHandler) WithdrawRequest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}
	logger.Sugar.Infof("User %d requested withdraw", userID)

	var request models.WithdrawRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Sugar.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, &request); err != nil {
		logger.Sugar.Errorf("Failed to unmarshal request: %v", err)
		http.Error(w, "Failed to unmarshal request", http.StatusBadRequest)
		return
	}
	logger.Sugar.Infof("Got withdraw request: %v", request)

	if !ValidateOrderNumber(request.Order) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}
	accrualInfoChan := make(chan models.AccrualInfo)
	go func(orderNumber string) {
		accrualInfo, err := fetchAccrualInfo(orderNumber, h.AccrualSystemAddress)
		if err != nil {
			close(accrualInfoChan)
			return
		}
		accrualInfoChan <- accrualInfo
		close(accrualInfoChan)
	}(request.Order)
	accrualInfo, ok := <-accrualInfoChan
	if !ok {
		http.Error(w, "Failed to get accrual info", http.StatusInternalServerError)
		return
	}
	response := h.service.WithdrawRequest(userID, request.Order, request.Sum, accrualInfo)
	w.WriteHeader(response)
}

func (h *OrderHandler) UserWithdrawls(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawls, err := h.service.GetUserWithdrawls(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(withdrawls) == 0 {
		http.Error(w, "No withdrawls for user", http.StatusNoContent)
		return
	}
	logger.Sugar.Infof("Got %d withdrawls for user", len(withdrawls))
	response, err := json.Marshal(withdrawls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}
