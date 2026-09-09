package card

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"webauthn-server/internal/timeutil"
)

// Repository defines DB operations for cards.
type Repository interface {
	GetCardsLinkedToUserId(ctx context.Context, userID string) ([]*Card, error)
	GetAllCards(ctx context.Context) ([]*Card, error)
	SaveCardUnderUserID(ctx context.Context, card *Card) error
	UpdateCardDetails(ctx context.Context, card *Card) error
	DeleteCard(ctx context.Context, cardID, userID string) error
	DeleteCardsByUserID(ctx context.Context, userID string) error
	GetCardByID(ctx context.Context, cardID string) (*Card, error)
	GetCardByPAN(ctx context.Context, pan string) (*Card, error)
}

// GetAllCards retrieves cards across all users for the admin visualizer.
func (r *SQLiteCardRepository) GetAllCards(ctx context.Context) ([]*Card, error) {
	rows, err := r.pool.QueryContext(ctx,
		"SELECT id, user_id, PAN, cardholder_name, bank_name, payment_method_type, card_brand, product_name, linked_phone_number, exp_month, exp_year, cvv, created_at, updated_at FROM cards ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		var card Card
		if err := rows.Scan(&card.ID, &card.UserID, &card.PAN, &card.CardholderName, &card.BankName, &card.PaymentMethodType, &card.CardBrand, &card.ProductName, &card.LinkedPhoneNumber, &card.ExpMonth, &card.ExpYear, &card.Cvv, &card.CreatedAt, &card.UpdatedAt); err != nil {
			return nil, err
		}
		cards = append(cards, &card)
	}
	return cards, rows.Err()
}

// SQLiteCardRepository is a concrete implementation of Repository using database/sql.
type SQLiteCardRepository struct {
	pool *sql.DB
}

// NewSQLiteCardRepository creates a new repository with the provided pool.
func NewSQLiteCardRepository(pool *sql.DB) *SQLiteCardRepository {
	return &SQLiteCardRepository{pool: pool}
}

// GetCardsLinkedToUserId retrieves all cards associated with a user ID.
func (r *SQLiteCardRepository) GetCardsLinkedToUserId(ctx context.Context, userID string) ([]*Card, error) {
	rows, err := r.pool.QueryContext(ctx,
		"SELECT id, user_id, PAN, cardholder_name, bank_name, payment_method_type, card_brand, product_name, linked_phone_number, exp_month, exp_year, cvv, created_at, updated_at FROM cards WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		var card Card
		err := rows.Scan(
			&card.ID,
			&card.UserID,
			&card.PAN,
			&card.CardholderName,
			&card.BankName,
			&card.PaymentMethodType,
			&card.CardBrand,
			&card.ProductName,
			&card.LinkedPhoneNumber,
			&card.ExpMonth,
			&card.ExpYear,
			&card.Cvv,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, &card)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

// SaveCardUnderUserID creates a new card for a user.
func (r *SQLiteCardRepository) SaveCardUnderUserID(ctx context.Context, card *Card) error {
	card.ID = uuid.NewString()
	card.CreatedAt = timeutil.Now()
	card.UpdatedAt = timeutil.Now()

	_, err := r.pool.ExecContext(ctx,
		"INSERT INTO cards (id, user_id, PAN, cardholder_name, bank_name, payment_method_type, card_brand, product_name, linked_phone_number, exp_month, exp_year, cvv, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		card.ID,
		card.UserID,
		card.PAN,
		card.CardholderName,
		card.BankName,
		card.PaymentMethodType,
		card.CardBrand,
		card.ProductName,
		card.LinkedPhoneNumber,
		card.ExpMonth,
		card.ExpYear,
		card.Cvv,
		card.CreatedAt,
		card.UpdatedAt,
	)
	return err
}

// UpdateCardDetails updates an existing card.
func (r *SQLiteCardRepository) UpdateCardDetails(ctx context.Context, card *Card) error {
	card.UpdatedAt = timeutil.Now()

	_, err := r.pool.ExecContext(ctx,
		"UPDATE cards SET PAN = ?, cardholder_name = ?, bank_name = ?, payment_method_type = ?, card_brand = ?, product_name = ?, linked_phone_number = ?, exp_month = ?, exp_year = ?, cvv = ?, updated_at = ? WHERE id = ? AND user_id = ?",
		card.PAN,
		card.CardholderName,
		card.BankName,
		card.PaymentMethodType,
		card.CardBrand,
		card.ProductName,
		card.LinkedPhoneNumber,
		card.ExpMonth,
		card.ExpYear,
		card.Cvv,
		card.UpdatedAt,
		card.ID,
		card.UserID,
	)
	return err
}

// DeleteCardsByUserID deletes all cards owned by a user.
func (r *SQLiteCardRepository) DeleteCardsByUserID(ctx context.Context, userID string) error {
	_, err := r.pool.ExecContext(ctx, "DELETE FROM cards WHERE user_id = ?", userID)
	return err
}

// DeleteCard deletes a card by ID (ensures it belongs to the user).
func (r *SQLiteCardRepository) DeleteCard(ctx context.Context, cardID, userID string) error {
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM cards WHERE id = ? AND user_id = ?",
		cardID,
		userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetCardByID retrieves a single card by ID.
func (r *SQLiteCardRepository) GetCardByID(ctx context.Context, cardID string) (*Card, error) {
	var card Card
	err := r.pool.QueryRowContext(ctx,
		"SELECT id, user_id, PAN, cardholder_name, bank_name, payment_method_type, card_brand, product_name, linked_phone_number, exp_month, exp_year, cvv, created_at, updated_at FROM cards WHERE id = ?",
		cardID,
	).Scan(
		&card.ID,
		&card.UserID,
		&card.PAN,
		&card.CardholderName,
		&card.BankName,
		&card.PaymentMethodType,
		&card.CardBrand,
		&card.ProductName,
		&card.ExpMonth,
		&card.LinkedPhoneNumber,
		&card.ExpYear,
		&card.Cvv,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &card, nil
}

// GetCardByPAN retrieves a card by PAN number.
func (r *SQLiteCardRepository) GetCardByPAN(ctx context.Context, pan string) (*Card, error) {
	var card Card
	err := r.pool.QueryRowContext(ctx,
		"SELECT id, user_id, PAN, cardholder_name, bank_name, payment_method_type, card_brand, product_name, linked_phone_number, exp_month, exp_year, cvv, created_at, updated_at FROM cards WHERE PAN = ?",
		pan,
	).Scan(
		&card.ID,
		&card.UserID,
		&card.PAN,
		&card.CardholderName,
		&card.BankName,
		&card.PaymentMethodType,
		&card.CardBrand,
		&card.ProductName,
		&card.ExpMonth,
		&card.LinkedPhoneNumber,
		&card.ExpYear,
		&card.Cvv,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &card, nil
}
