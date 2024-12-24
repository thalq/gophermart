package models

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	jwt.StandardClaims
	UserID int64 `json:"user_id"`
}

type Order struct {
	Number     string    `db:"order_id" json:"number"`
	Status     string    `db:"status" json:"status"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
	Accrual    float32   `db:"accrual" json:"accrual,omitempty"`
}

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

type WithdrawResponse struct {
	OrderID     string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type AccrualInfo struct {
	OrderID string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func (a *AccrualInfo) SetDefaults(orderID string) {
	if a.Status == "" {
		a.Status = "NEW"
	}
	if a.Accrual == 0 {
		a.Accrual = 0.0
	}
}
