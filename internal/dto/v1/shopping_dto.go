package v1dto

import (
	"time"

	"github.com/google/uuid"

	"gocommerce/internal/db/sqlc"
)

type CheckoutInput struct {
	RecipientName string `json:"recipient_name" binding:"required,max=100"`
	Phone         string `json:"phone" binding:"required,min=8,max=20"`
	Address       string `json:"address" binding:"required,min=5,max=500"`
}
type OrderDTO struct {
	ID            uuid.UUID        `json:"id"`
	Status        string           `json:"status"`
	TotalVND      int64            `json:"total_vnd"`
	RecipientName string           `json:"recipient_name"`
	Phone         string           `json:"phone"`
	Address       string           `json:"address"`
	PaymentMethod string           `json:"payment_method"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	Items         []sqlc.OrderItem `json:"items,omitempty"`
}

func MapOrder(o sqlc.Order, items []sqlc.OrderItem) OrderDTO {
	return OrderDTO{ID: o.ID, Status: o.Status, TotalVND: o.TotalVnd, RecipientName: o.RecipientName, Phone: o.Phone, Address: o.Address, PaymentMethod: o.PaymentMethod, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt, Items: items}
}

type CartDTO struct {
	Items    []sqlc.ListCartItemsRow `json:"items"`
	TotalVND int64                   `json:"total_vnd"`
}
