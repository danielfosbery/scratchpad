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
	Subtotal int64
	Discount int64
	Tax      int64
	Total    int64
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

func caclulateBreakdown(unitPrice int64, currencyCode string, quantity int, discount *Discount, taxRate float64, taxInclusive, payTax bool) (Breakdown, error) {
	var tax currency.Amount

	up, err := currency.NewAmount(fmt.Sprintf("%d", unitPrice), currencyCode)

	subtotal, err := up.Mul(fmt.Sprintf("%d", quantity))
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
			tax, err = currency.NewAmount("0", currencyCode)

			if err != nil {
				return Breakdown{}, err
			}

			total = preTaxAmount
		}

		return Breakdown{
			Subtotal: subtotal.Round().BigInt().Int64(),
			Discount: discountAmount.Round().BigInt().Int64(),
			Tax:      tax.Round().BigInt().Int64(),
			Total:    total.Round().BigInt().Int64(),
		}, nil
	}

	if payTax {
		tax, err = postDiscountAmount.Mul(fmt.Sprintf("%f", taxRate))
		if err != nil {
			return Breakdown{}, err
		}
	} else {
		tax, err = currency.NewAmount("0", currencyCode)

		if err != nil {
			return Breakdown{}, err
		}
	}

	total, err := postDiscountAmount.Add(tax)
	if err != nil {
		return Breakdown{}, err
	}

	return Breakdown{
		Subtotal: subtotal.Round().BigInt().Int64(),
		Discount: discountAmount.Round().BigInt().Int64(),
		Tax:      tax.Round().BigInt().Int64(),
		Total:    total.Round().BigInt().Int64(),
	}, nil
}

func applyDiscount(amount currency.Amount, discount *Discount) (currency.Amount, error) {
	if discount == nil {
		a, _ := currency.NewAmount("0", amount.CurrencyCode())

		return a, nil
	}

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
