package main

import (
	"testing"

	"github.com/bojanz/currency"
	"github.com/stretchr/testify/assert"
)

func TestBreadownIncTax(t *testing.T) {
	up, _ := currency.NewAmount("34.95", "EUR")
	b, err := caclulateBreakdown(up, 1, Discount{Type: "percentage", Amount: "0.5"}, 0.2, true, false)
	if err != nil {
		panic(err)
	}

	assertAmountEqual(t, "34.95", "EUR", b.Subtotal)
	assertAmountEqual(t, "17.48", "EUR", b.Discount)
	assertAmountEqual(t, "0", "EUR", b.Tax)
	assertAmountEqual(t, "17.47", "EUR", b.Total)
	assertBreakdownAddsUpTaxInternal(t, b)
}

func assertAmountEqual(t *testing.T, expectedAmount, expectedCurrency string, actual currency.Amount) {
	expected, err := currency.NewAmount(expectedAmount, expectedCurrency)
	if err != nil {
		t.Error(err)
	}

	assert.True(t, expected.Equal(actual), "actual amount %s does not equal expected %s", actual.String(), expected.String())
}

// subtotal - discount + tax = total
func assertBreakdownAddsUpTaxExternal(t *testing.T, breakdown Breakdown) {
	preTaxAmount, err := breakdown.Subtotal.Sub(breakdown.Discount)
	if err != nil {
		t.Error(err)
	}

	postTaxAmount, err := preTaxAmount.Add(breakdown.Tax)
	if err != nil {
		t.Error(err)
	}

	assert.True(t, postTaxAmount.Equal(breakdown.Total), "added amount %s does not equal total %s", postTaxAmount.String(), breakdown.Total.String())
}

// subtotal - discount = total
func assertBreakdownAddsUpTaxInternal(t *testing.T, breakdown Breakdown) {
	postDiscountAmount, err := breakdown.Subtotal.Sub(breakdown.Discount)
	if err != nil {
		t.Error(err)
	}

	assert.True(t, postDiscountAmount.Equal(breakdown.Total), "added amount %s does not equal total %s", postDiscountAmount.String(), breakdown.Total.String())
}
