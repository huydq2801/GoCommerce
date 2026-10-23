package v1service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"huydq2801.com/golearn/internal/db/sqlc"
	v1dto "huydq2801.com/golearn/internal/dto/v1"
	"huydq2801.com/golearn/internal/repository"
	"huydq2801.com/golearn/internal/utils"
)

type ShoppingService struct{ store *repository.ShoppingStore }

func NewShoppingService(store *repository.ShoppingStore) *ShoppingService {
	return &ShoppingService{store: store}
}
func (s *ShoppingService) GetCart(ctx context.Context, userID int32) (v1dto.CartDTO, error) {
	var view v1dto.CartDTO
	err := s.store.Transact(ctx, func(q *sqlc.Queries) error {
		cartID, err := q.EnsureCart(ctx, userID)
		if err != nil {
			return err
		}
		items, err := q.ListCartItems(ctx, cartID)
		if err != nil {
			return err
		}
		view.Items = items
		for _, item := range items {
			view.TotalVND += int64(item.Quantity) * item.PriceVnd
		}
		return nil
	})
	return view, utils.DatabaseError(err)
}
func (s *ShoppingService) SetCartItem(ctx context.Context, userID int32, productID uuid.UUID, quantity int32) error {
	if quantity < 1 || quantity > 1000 {
		return utils.NewError("Quantity must be between 1 and 1000", utils.ErrCodeBadRequest)
	}
	return utils.DatabaseError(s.store.Transact(ctx, func(q *sqlc.Queries) error {
		cartID, err := q.EnsureCart(ctx, userID)
		if err != nil {
			return err
		}
		if _, err = q.GetProduct(ctx, productID); err != nil {
			return err
		}
		items, err := q.ListCartItems(ctx, cartID)
		if err != nil {
			return err
		}
		exists := false
		for _, item := range items {
			if item.ProductID == productID {
				exists = true
			}
		}
		if !exists && len(items) >= 50 {
			return utils.NewError("Cart is limited to 50 distinct products", utils.ErrCodeConflict)
		}
		return q.UpsertCartItem(ctx, sqlc.UpsertCartItemParams{CartID: cartID, ProductID: productID, Quantity: quantity})
	}))
}
func (s *ShoppingService) RemoveCartItem(ctx context.Context, userID int32, productID uuid.UUID) error {
	return utils.DatabaseError(s.store.Transact(ctx, func(q *sqlc.Queries) error {
		cartID, err := q.EnsureCart(ctx, userID)
		if err != nil {
			return err
		}
		if productID == uuid.Nil {
			return q.ClearCart(ctx, cartID)
		}
		return q.DeleteCartItem(ctx, sqlc.DeleteCartItemParams{CartID: cartID, ProductID: productID})
	}))
}

var phonePattern = regexp.MustCompile(`^\+?[0-9]{8,15}$`)

func (s *ShoppingService) Checkout(ctx context.Context, userID int32, key uuid.UUID, input v1dto.CheckoutInput) (v1dto.OrderDTO, bool, error) {
	input.RecipientName = strings.TrimSpace(input.RecipientName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Address = strings.TrimSpace(input.Address)
	if input.RecipientName == "" || len(input.Address) < 5 || !phonePattern.MatchString(input.Phone) {
		return v1dto.OrderDTO{}, false, utils.NewError("Invalid recipient information", utils.ErrCodeBadRequest)
	}
	encoded, _ := json.Marshal(input)
	digest := sha256.Sum256(encoded)
	requestHash := hex.EncodeToString(digest[:])
	var result v1dto.OrderDTO
	replay := false
	err := s.store.Transact(ctx, func(q *sqlc.Queries) error {
		// All cart edits and checkouts acquire this row first, serializing operations for one user.
		cartID, err := q.EnsureCart(ctx, userID)
		if err != nil {
			return err
		}
		existing, err := q.GetOrderByKey(ctx, sqlc.GetOrderByKeyParams{UserID: userID, IdempotencyKey: key})
		if err == nil {
			if existing.RequestHash != requestHash {
				return utils.NewError("Idempotency key was already used with different input", utils.ErrCodeConflict)
			}
			items, e := q.ListOrderItems(ctx, existing.ID)
			if e != nil {
				return e
			}
			result = v1dto.MapOrder(existing, items)
			replay = true
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		cart, err := q.ListCartItems(ctx, cartID)
		if err != nil {
			return err
		}
		if len(cart) == 0 {
			return utils.NewError("Cart is empty", utils.ErrCodeConflict)
		}
		total := int64(0)
		snapshots := make([]sqlc.CreateOrderItemParams, 0, len(cart))
		// ListCartItems orders by product UUID. Every inventory path uses this lock order.
		for _, line := range cart {
			p, e := q.LockProduct(ctx, line.ProductID)
			if e != nil {
				return e
			}
			if !p.IsActive || p.StockQuantity < line.Quantity {
				return utils.NewError("A product is unavailable or has insufficient stock", utils.ErrCodeConflict)
			}
			rows, e := q.DecreaseStock(ctx, sqlc.DecreaseStockParams{ID: p.ID, StockQuantity: line.Quantity})
			if e != nil {
				return e
			}
			if rows != 1 {
				return utils.NewError("Insufficient stock", utils.ErrCodeConflict)
			}
			total += p.PriceVnd * int64(line.Quantity)
			snapshots = append(snapshots, sqlc.CreateOrderItemParams{ProductID: p.ID, Sku: p.Sku, ProductName: p.Name, UnitPriceVnd: p.PriceVnd, Quantity: line.Quantity})
		}
		order, e := q.CreateOrder(ctx, sqlc.CreateOrderParams{UserID: userID, TotalVnd: total, RecipientName: input.RecipientName, Phone: input.Phone, Address: input.Address, IdempotencyKey: key, RequestHash: requestHash})
		if e != nil {
			return e
		}
		for _, line := range snapshots {
			line.OrderID = order.ID
			if e = q.CreateOrderItem(ctx, line); e != nil {
				return e
			}
		}
		if e = q.ClearCart(ctx, cartID); e != nil {
			return e
		}
		items, e := q.ListOrderItems(ctx, order.ID)
		if e != nil {
			return e
		}
		result = v1dto.MapOrder(order, items)
		return nil
	})
	return result, replay, utils.DatabaseError(err)
}
func (s *ShoppingService) GetOrder(ctx context.Context, userID int32, id uuid.UUID, admin bool) (v1dto.OrderDTO, error) {
	q := s.store.Queries()
	order, err := q.GetOrder(ctx, id)
	if err != nil {
		return v1dto.OrderDTO{}, utils.DatabaseError(err)
	}
	if !admin && order.UserID != userID {
		return v1dto.OrderDTO{}, utils.NewError("Order not found", utils.ErrCodeNotFound)
	}
	items, err := q.ListOrderItems(ctx, id)
	return v1dto.MapOrder(order, items), utils.DatabaseError(err)
}
func (s *ShoppingService) ListOrders(ctx context.Context, userID int32, p utils.Page) ([]v1dto.OrderDTO, int64, error) {
	q := s.store.Queries()
	p.Defaults()
	orders, err := q.ListOrders(ctx, sqlc.ListOrdersParams{UserID: userID, PageLimit: int32(p.Limit), PageOffset: p.Offset()})
	if err != nil {
		return nil, 0, utils.DatabaseError(err)
	}
	out := make([]v1dto.OrderDTO, 0, len(orders))
	for _, order := range orders {
		out = append(out, v1dto.MapOrder(order, nil))
	}
	total, err := q.CountOrders(ctx, userID)
	return out, total, utils.DatabaseError(err)
}
func (s *ShoppingService) CancelOrder(ctx context.Context, userID int32, id uuid.UUID) (v1dto.OrderDTO, error) {
	var result v1dto.OrderDTO
	err := s.store.Transact(ctx, func(q *sqlc.Queries) error {
		order, err := q.LockOrder(ctx, id)
		if err != nil {
			return err
		}
		if order.UserID != userID {
			return utils.NewError("Order not found", utils.ErrCodeNotFound)
		}
		items, err := q.ListOrderItems(ctx, id)
		if err != nil {
			return err
		}
		if order.Status == "cancelled" {
			result = v1dto.MapOrder(order, items)
			return nil
		}
		if order.Status != "placed" {
			return utils.NewError("Only placed orders may be cancelled", utils.ErrCodeConflict)
		}
		for _, line := range items {
			if _, err = q.LockProduct(ctx, line.ProductID); err != nil {
				return err
			}
			if err = q.RestoreStock(ctx, sqlc.RestoreStockParams{ID: line.ProductID, StockQuantity: line.Quantity}); err != nil {
				return err
			}
		}
		order, err = q.SetOrderStatus(ctx, sqlc.SetOrderStatusParams{ID: id, Status: "cancelled"})
		if err != nil {
			return err
		}
		result = v1dto.MapOrder(order, items)
		return nil
	})
	return result, utils.DatabaseError(err)
}
func (s *ShoppingService) AdvanceOrder(ctx context.Context, id uuid.UUID, status string) (v1dto.OrderDTO, error) {
	var result v1dto.OrderDTO
	err := s.store.Transact(ctx, func(q *sqlc.Queries) error {
		order, err := q.LockOrder(ctx, id)
		if err != nil {
			return err
		}
		next := map[string]string{"placed": "processing", "processing": "shipped", "shipped": "completed"}
		if order.Status != status {
			if next[order.Status] != status {
				return utils.NewError("Invalid order status transition", utils.ErrCodeConflict)
			}
			order, err = q.SetOrderStatus(ctx, sqlc.SetOrderStatusParams{ID: id, Status: status})
			if err != nil {
				return err
			}
		}
		items, err := q.ListOrderItems(ctx, id)
		result = v1dto.MapOrder(order, items)
		return err
	})
	return result, utils.DatabaseError(err)
}
