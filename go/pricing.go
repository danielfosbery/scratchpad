package main

import (
	"fmt"

	"github.com/bojanz/currency"
)

type Item struct {
	ProductID string
	Name      string
	UnitPrice currency.Amount
}

type Breakdown struct {
	Subtotal currency.Amount
	Discount currency.Amount
	Tax      currency.Amount
	Total    currency.Amount
}

type Discount struct {
	Type         string
	Amount       string
	CurrencyCode string
}

func main() {
	up, _ := currency.NewAmount("34.95", "EUR")

	fmt.Printf("%+v\n", up.BigInt())

	up, _ = currency.NewAmount("3495", "JPY")

	fmt.Printf("%+v\n", up.BigInt())
}

func caclulateBreakdown(unitPrice currency.Amount, quantity int, discount Discount, taxRate float32, taxInclusive, payTax bool) (Breakdown, error) {
	var tax currency.Amount

	subtotal, err := unitPrice.Mul(fmt.Sprintf("%d", quantity))
	if err != nil {
		return Breakdown{}, err
	}

	discountAmount, err := applyDiscount(subtotal, discount)
	if err != nil {
		return Breakdown{}, err
	}

	postDiscountAmount, err := subtotal.Sub(discountAmount.Round())
	if err != nil {
		return Breakdown{}, err
	}

	if taxInclusive {
		preTaxAmount, err := postDiscountAmount.Div(fmt.Sprintf("%f", 1+taxRate))
		if err != nil {
			return Breakdown{}, err
		}

		var total currency.Amount
		if payTax {
			tax, err = postDiscountAmount.Sub(preTaxAmount.Round())
			if err != nil {
				return Breakdown{}, err
			}

			total = postDiscountAmount
		} else {
			tax, err = currency.NewAmount("0", unitPrice.CurrencyCode())

			if err != nil {
				return Breakdown{}, err
			}

			total = preTaxAmount
		}

		return Breakdown{
			Subtotal: subtotal.Round(),
			Discount: discountAmount.Round(),
			Tax:      tax.Round(),
			Total:    total.Round(),
		}, nil
	}

	if payTax {
		tax, err = postDiscountAmount.Mul(fmt.Sprintf("%f", taxRate))
		if err != nil {
			return Breakdown{}, err
		}
	} else {
		tax, err = currency.NewAmount("0", unitPrice.CurrencyCode())

		if err != nil {
			return Breakdown{}, err
		}
	}

	total, err := postDiscountAmount.Add(tax)
	if err != nil {
		return Breakdown{}, err
	}

	return Breakdown{
		Subtotal: subtotal.Round(),
		Discount: discountAmount.Round(),
		Tax:      tax.Round(),
		Total:    total.Round(),
	}, nil
}

func applyDiscount(amount currency.Amount, discount Discount) (currency.Amount, error) {
	if discount.Type == "flat_amount" {
		discountAmount, err := currency.NewAmount(discount.Amount, discount.CurrencyCode)
		if err != nil {
			return currency.Amount{}, err
		}

		if discountAmount.CurrencyCode() == amount.CurrencyCode() {
			return discountAmount, nil
		}

		// TODO get exchange rate
		discountAmountConverted, err := discountAmount.Convert(amount.CurrencyCode(), "1.2")
		if err != nil {
			return currency.Amount{}, err
		}

		return discountAmountConverted, nil
	}

	if discount.Type == "percentage" {
		discountAmount, err := amount.Mul(discount.Amount)
		if err != nil {
			return currency.Amount{}, err
		}

		return discountAmount, nil
	}

	return currency.Amount{}, fmt.Errorf("discount.Type %s is unknown", discount.Type)
}
