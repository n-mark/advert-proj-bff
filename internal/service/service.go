package service

import (
	"context"
	"fmt"
	"log/slog"

	"bff-finalproj/internal/clients"
	"bff-finalproj/internal/config"
)

// BFF aggregates downstream service calls.
type BFF struct {
	cfg      *config.Config
	advertCmd *clients.Client
	advertQ  *clients.Client
	profile  *clients.Client
	order    *clients.Client
	delivery *clients.Client
	billing  *clients.Client
	dialog   *clients.Client
}

// NewBFF creates a BFF service with all downstream clients.
func NewBFF(cfg *config.Config) *BFF {
	return &BFF{
		cfg:       cfg,
		advertCmd: clients.New(cfg.AdvertCmdURL, cfg.InternalToken),
		advertQ:   clients.New(cfg.AdvertQueryURL, cfg.InternalToken),
		profile:   clients.New(cfg.ProfileURL, cfg.InternalToken),
		order:     clients.New(cfg.OrderURL, cfg.InternalToken),
		delivery:  clients.New(cfg.DeliveryURL, cfg.InternalToken),
		billing:   clients.New(cfg.BillingURL, cfg.InternalToken),
		dialog:    clients.New(cfg.DialogURL, cfg.InternalToken),
	}
}

// GetAdvertFull aggregates an advert card with seller profile.
func (b *BFF) GetAdvertFull(ctx context.Context, advertID string) (map[string]any, error) {
	var advert map[string]any
	if _, err := b.advertCmd.Get(ctx, "/api/v1/adverts/"+advertID, &advert); err != nil {
		return nil, fmt.Errorf("fetch advert: %w", err)
	}

	createdBy, _ := advert["created_by"].(string)
	if createdBy == "" {
		// try nested field names used by advert-cmd-svc
		if cb, ok := advert["createdBy"].(string); ok {
			createdBy = cb
		}
	}

	if createdBy != "" {
		var profile map[string]any
		if _, err := b.profile.Get(ctx, "/internal/v1/users/"+createdBy, &profile); err != nil {
			slog.Warn("failed to enrich seller profile", "user_id", createdBy, "error", err)
		} else {
			advert["seller"] = profile
		}
	}

	return advert, nil
}

// GetOrderFull aggregates an order with advert, delivery and billing details.
func (b *BFF) GetOrderFull(ctx context.Context, orderID string) (map[string]any, error) {
	var order map[string]any
	if _, err := b.order.Get(ctx, "/api/v1/orders/"+orderID, &order); err != nil {
		return nil, fmt.Errorf("fetch order: %w", err)
	}

	advertID := extractString(order, "advert_id", "advertID")
	if advertID != "" {
		var advert map[string]any
		if _, err := b.advertCmd.Get(ctx, "/internal/v1/adverts/"+advertID, &advert); err != nil {
			slog.Warn("failed to enrich advert", "advert_id", advertID, "error", err)
		} else {
			order["advert"] = advert
		}
	}

	deliveryID := extractString(order, "delivery_id", "deliveryID")
	if deliveryID != "" {
		var delivery map[string]any
		if _, err := b.delivery.Get(ctx, "/internal/v1/deliveries/"+deliveryID, &delivery); err != nil {
			slog.Warn("failed to enrich delivery", "delivery_id", deliveryID, "error", err)
		} else {
			order["delivery"] = delivery
		}
	}

	// billing transactions by order id
	var txs []map[string]any
	if _, err := b.billing.Get(ctx, "/internal/v1/orders/"+orderID+"/transactions", &txs); err != nil {
		slog.Warn("failed to fetch transactions", "order_id", orderID, "error", err)
	} else {
		order["transactions"] = txs
	}

	return order, nil
}

// GetUserCabinet aggregates profile, orders and dialogs for a user.
func (b *BFF) GetUserCabinet(ctx context.Context, userID string) (map[string]any, error) {
	cabinet := map[string]any{
		"user_id": userID,
		"profile": nil,
		"orders":  nil,
		"dialogs": nil,
	}

	var profile map[string]any
	if _, err := b.profile.Get(ctx, "/internal/v1/users/"+userID, &profile); err != nil {
		return nil, fmt.Errorf("fetch profile: %w", err)
	}
	cabinet["profile"] = profile

	var orders []map[string]any
	if _, err := b.order.Get(ctx, "/internal/v1/users/"+userID+"/orders", &orders); err != nil {
		slog.Warn("failed to fetch orders", "user_id", userID, "error", err)
	}
	cabinet["orders"] = orders

	var dialogs []map[string]any
	if _, err := b.dialog.Get(ctx, "/internal/v1/users/"+userID+"/dialogs", &dialogs); err != nil {
		slog.Warn("failed to fetch dialogs", "user_id", userID, "error", err)
	}
	cabinet["dialogs"] = dialogs

	return cabinet, nil
}

func extractString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok {
			return v
		}
	}
	return ""
}
