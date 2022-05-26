package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type CreatePaymentRequestRequest struct {
	CustomerID       uuid.UUID                  `json:"customer_id"`
	BillingAddressID uuid.UUID                  `json:"billing_address_id"`
	BusinessID       *uuid.UUID                 `json:"business_id,omitempty"`
	Discount         *DiscountDynamic           `json:"discount,omitempty"`
	DiscountID       *uuid.UUID                 `json:"discount_id,omitempty"`
	Items            []CreatePaymentRequestItem `json:"items"`
}

type DiscountDynamic struct {
	Name      string `json:"name"`
	Value     int    `json:"value"`
	Type      string `json:"type"`
	Recurring bool   `json:"recurring"`
}

type ProductPriceDynamic struct {
	ProductID uuid.UUID `json:"product_id"`
	UnitPrice int64     `json:"unit_price"`
}

type ProductPriceConfig struct {
	ProductID           uuid.UUID `json:"product_id"`
	DefaultCurrencyCode string    `json:"default_currency_code"`
	Prices              []Price   `json:"prices"`
}

type Price struct {
	CurrencyCode string `json:"currency_code"`
	UnitPrice    int64  `json:"unit_price"`
}

type CreatePaymentRequestItem struct {
	ProductPriceID *uuid.UUID           `json:"product_price_id,omitempty"`
	ProductPrice   *ProductPriceDynamic `json:"product_price,omitempty"`
	Quantity       int                  `json:"quantity"`
	Discount       *DiscountDynamic     `json:"discount"`
	Description    string               `json:"description,omitempty"`
}

type PaymentRequestResponse struct {
	ID               uuid.UUID            `json:"id"`
	CustomerID       uuid.UUID            `json:"customer_id"`
	BillingAddressID uuid.UUID            `json:"billing_address_id"`
	BusinessID       *uuid.UUID           `json:"business_id,omitempty"`
	Discount         *DiscountDynamic     `json:"discount,omitempty"`
	Items            []PaymentRequestItem `json:"items"`
}

type PaymentRequestItem struct {
	Totals      Breakdown        `json:"totals"`
	UnitPrice   int64            `json:"unit_price"`
	Quantity    int              `json:"quantity"`
	Discount    *DiscountDynamic `json:"discount"`
	ProductName string           `json:"product_name"`
	Description string           `json:"description,omitempty"`
}

type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	TaxCategory string    `json:"tax_category"`
}

var paymentRequestCache map[string]PaymentRequestResponse

func CreatePaymentRequest(ctx context.Context, r CreatePaymentRequestRequest) (PaymentRequestResponse, error) {
	taxRate := 0.2                // TODO get from API
	taxInclusive := true          // TODO get from API
	payTax := r.BusinessID != nil // TODO get business via API and check if tax code set
	currencyCode := "USD"         // TODO get based on address

	pr := PaymentRequestResponse{
		ID:               uuid.New(),
		CustomerID:       r.CustomerID,
		BillingAddressID: r.BillingAddressID,
		BusinessID:       r.BusinessID,
	}

	for _, i := range r.Items {
		if i.ProductPrice != nil && i.ProductPriceID != nil {
			return PaymentRequestResponse{}, errors.New("can't have both product_price_id and product_price set on an item")
		}

		if i.ProductPriceID != nil {
			// get from API
			pp, err := getProductPrice(*i.ProductPriceID)
			if err != nil {
				return PaymentRequestResponse{}, fmt.Errorf("cannot get product price: %w", err)
			}

			i.ProductPrice = &pp
		}

		product, err := getProduct(i.ProductPrice.ProductID)
		if err != nil {
			return PaymentRequestResponse{}, fmt.Errorf("cannot get product: %w", err)
		}

		bd, err := caclulateBreakdown(i.ProductPrice.UnitPrice, currencyCode, i.Quantity, nil, taxRate, taxInclusive, payTax) // TODO pass in discount
		if err != nil {
			return PaymentRequestResponse{}, fmt.Errorf("cannot calculate price breakdown: %w", err)
		}

		pr.Items = append(pr.Items, PaymentRequestItem{
			UnitPrice:   i.ProductPrice.UnitPrice,
			Quantity:    i.Quantity,
			Discount:    i.Discount,
			ProductName: product.Name,
			Description: i.Description,
			Totals:      bd,
		})
	}

	paymentRequestCache[pr.ID.String()] = pr

	return pr, nil
}

func getPaymentRequest(id uuid.UUID) (PaymentRequestResponse, error) {
	pr, ok := paymentRequestCache[id.String()]
	if !ok {
		return PaymentRequestResponse{}, fmt.Errorf("payment request not found: id %s", id)
	}

	return pr, nil
}

func getProductPrice(id uuid.UUID) (ProductPriceDynamic, error) {
	m := map[string]ProductPriceDynamic{
		"D26C7F6A-75F2-4745-8DC3-0540FD9F2DDE": ProductPriceConfig{
			ProductID: uuid.MustParse("FE508038-65CD-4AC4-A064-EF4BC0AC8FFB"),
			UnitPrice: 2000,
		},
	}

	pp, ok := m[id.String()]
	if !ok {
		return ProductPriceDynamic{}, fmt.Errorf("product price not found: id %s", id)
	}

	return pp, nil
}

func getProduct(id uuid.UUID) (Product, error) {
	m := map[string]Product{
		"FE508038-65CD-4AC4-A064-EF4BC0AC8FFB": Product{
			ID:          uuid.MustParse("FE508038-65CD-4AC4-A064-EF4BC0AC8FFB"),
			Name:        "Test Product",
			TaxCategory: "ebook",
		},
	}

	p, ok := m[id.String()]
	if !ok {
		return Product{}, fmt.Errorf("product not found: id %s", id)
	}

	return p, nil
}
