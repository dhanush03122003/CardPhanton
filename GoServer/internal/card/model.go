package card

import "time"

// PaymentMethodType represents the type of payment method
type PaymentMethodType string

const (
	PaymentMethodCredit PaymentMethodType = "Credit"
	PaymentMethodDebit  PaymentMethodType = "Debit"
)

// IsValid checks if the payment method type is valid
func (p PaymentMethodType) IsValid() bool {
	switch p {
	case PaymentMethodCredit, PaymentMethodDebit:
		return true
	}
	return false
}

type Card struct {
	ID                string           `json:"id" db:"id"`
	UserID            string           `json:"user_id" db:"user_id"`
	PAN               string           `json:"pan" db:"pan"`
	CardholderName    string           `json:"cardholder_name" db:"cardholder_name"`
	BankName          string           `json:"bank_name" db:"bank_name"`
	PaymentMethodType PaymentMethodType `json:"payment_method_type" db:"payment_method_type"`
	CardBrand         string           `json:"card_brand" db:"card_brand"`
	ExpMonth          int              `json:"exp_month" db:"exp_month"`
	ExpYear           int              `json:"exp_year" db:"exp_year"`
	Cvv               int              `json:"cvv" db:"cvv"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
}