package card

// ErrorMessage is an API error message owned by the card handlers.
type ErrorMessage string

const (
	ErrMsgUnauthorized        ErrorMessage = "unauthorized"
	ErrMsgCardIDRequired      ErrorMessage = "card id is required"
	ErrMsgCardNotFound        ErrorMessage = "card not found"
	ErrMsgNotAuthorizedUpdate ErrorMessage = "not authorized to update this card"
	ErrMsgNotAuthorizedDelete ErrorMessage = "not authorized to delete this card"
)
