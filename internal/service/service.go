package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"bff-finalproj/internal/clients"
	"bff-finalproj/internal/config"
)

// BFF aggregates downstream service calls.
type BFF struct {
	cfg       *config.Config
	auth      *clients.Client
	advertCmd *clients.Client
	advertQ   *clients.Client
	profile   *clients.Client
	order     *clients.Client
	delivery  *clients.Client
	billing   *clients.Client
	dialog    *clients.Client
}

// NewBFF creates a BFF service with all downstream clients.
func NewBFF(cfg *config.Config) *BFF {
	return &BFF{
		cfg:       cfg,
		auth:      clients.New(cfg.AuthURL, cfg.InternalToken),
		advertCmd: clients.New(cfg.AdvertCmdURL, cfg.InternalToken),
		advertQ:   clients.New(cfg.AdvertQueryURL, cfg.InternalToken),
		profile:   clients.New(cfg.ProfileURL, cfg.InternalToken),
		order:     clients.New(cfg.OrderURL, cfg.InternalToken),
		delivery:  clients.New(cfg.DeliveryURL, cfg.InternalToken),
		billing:   clients.New(cfg.BillingURL, cfg.InternalToken),
		dialog:    clients.New(cfg.DialogURL, cfg.InternalToken),
	}
}

// GetAdvertFull aggregates an advert card with seller profile (name, surname)
// and seller phone from auth-service.
func (b *BFF) GetAdvertFull(ctx context.Context, advertID string) (map[string]any, error) {
	var advert map[string]any
	if _, err := b.advertCmd.Get(ctx, "/api/v1/adverts/"+advertID, &advert); err != nil {
		return nil, fmt.Errorf("fetch advert: %w", err)
	}

	b.enrichAdvert(ctx, advert)
	return advert, nil
}

// GetOrderFull aggregates an order with advert, delivery, billing details and
// buyer/seller phones from auth-service.
func (b *BFF) GetOrderFull(ctx context.Context, orderID string) (map[string]any, error) {
	var order map[string]any
	if _, err := b.order.Get(ctx, "/api/v1/order/"+orderID, &order); err != nil {
		return nil, fmt.Errorf("fetch order: %w", err)
	}

	// order-svc keeps advert id inside items[0].advert_id
	if advertID := extractAdvertID(order); advertID != "" {
		var advert map[string]any
		if _, err := b.advertCmd.Get(ctx, "/internal/v1/adverts/"+advertID, &advert); err != nil {
			slog.Warn("failed to enrich advert", "advert_id", advertID, "error", err)
		} else {
			b.enrichAdvert(ctx, advert)
			order["advert"] = advert
		}
	}

	deliveryID := extractID(order, "delivery_id", "deliveryID")
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

	// buyer = user_id, seller = seller_id
	if buyerID := extractID(order, "user_id", "userID", "buyer_id", "buyerID"); buyerID != "" {
		if phone := b.fetchPhone(ctx, buyerID); phone != "" {
			order["buyer_phone"] = phone
		}
	}
	if sellerID := extractID(order, "seller_id", "sellerID"); sellerID != "" {
		if phone := b.fetchPhone(ctx, sellerID); phone != "" {
			order["seller_phone"] = phone
		}
	}

	return order, nil
}

// enrichAdvert attaches a `seller` object (name from profile-service, phone
// from auth-service) to an advert using its created_by field.
func (b *BFF) enrichAdvert(ctx context.Context, advert map[string]any) {
	if advert == nil {
		return
	}

	createdBy := extractID(advert, "created_by", "createdBy")
	if createdBy == "" {
		return
	}

	seller := map[string]any{}

	var profile map[string]any
	if _, err := b.profile.Get(ctx, "/internal/v1/users/"+createdBy, &profile); err != nil {
		slog.Warn("failed to enrich seller profile", "user_id", createdBy, "error", err)
	} else {
		seller = profile
	}

	if phone := b.fetchPhone(ctx, createdBy); phone != "" {
		seller["phone"] = phone
	}

	if len(seller) > 0 {
		advert["seller"] = seller
	}
}

// GetUserCabinet aggregates profile, orders and dialogs for a user, plus phone.
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

	// phone from auth-service
	if phone := b.fetchPhone(ctx, userID); phone != "" {
		cabinet["phone"] = phone
	}

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

// fetchPhone returns the user phone from auth-service, or "" on failure.
func (b *BFF) fetchPhone(ctx context.Context, userID string) string {
	var user map[string]any
	if _, err := b.auth.Get(ctx, "/internal/v1/users/"+userID, &user); err != nil {
		slog.Warn("failed to fetch user from auth", "user_id", userID, "error", err)
		return ""
	}

	phone, _ := user["phone"].(string)
	return phone
}

// extractID pulls an identifier that may be a string or a JSON number and
// returns it as a string. JSON numbers are unmarshalled as float64.
func extractID(m map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}

		switch t := v.(type) {
		case string:
			if t != "" {
				return t
			}
		case float64:
			return strconv.FormatInt(int64(t), 10)
		case int64:
			return strconv.FormatInt(t, 10)
		case int:
			return strconv.Itoa(t)
		}
	}
	return ""
}

// extractAdvertID returns the advert id referenced by an order. order-svc
// stores it as a top-level advert_id (rare) or inside items[0].advert_id.
func extractAdvertID(order map[string]any) string {
	if id := extractID(order, "advert_id", "advertID"); id != "" {
		return id
	}

	items, ok := order["items"].([]any)
	if !ok || len(items) == 0 {
		return ""
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		return ""
	}
	return extractID(first, "advert_id", "advertID")
}