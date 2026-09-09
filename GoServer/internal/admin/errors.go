package admin

// ErrorMessage is an API error message owned by the admin handlers.
type ErrorMessage string

const (
	ErrMsgFetchUsers            ErrorMessage = "Failed to fetch users"
	ErrMsgFetchUserDetails      ErrorMessage = "Failed to fetch user details"
	ErrMsgAdminIdentityMissing  ErrorMessage = "Admin identity missing"
	ErrMsgInvalidAdminIdentity  ErrorMessage = "Invalid admin identity"
	ErrMsgUserIDRequired        ErrorMessage = "User ID required"
	ErrMsgUserNotFound          ErrorMessage = "User not found"
	ErrMsgApproveUser           ErrorMessage = "Failed to approve user"
	ErrMsgRejectUser            ErrorMessage = "Failed to reject user"
	ErrMsgSuspendUser           ErrorMessage = "Failed to suspend user"
	ErrMsgReactivateUser        ErrorMessage = "Failed to re-enable user"
	ErrMsgDeleteUser            ErrorMessage = "Failed to delete user"
	ErrMsgCardNotFound          ErrorMessage = "Card not found"
	ErrMsgDeleteCard            ErrorMessage = "Failed to delete card"
	ErrMsgLastAuthenticator     ErrorMessage = "Cannot delete the last authenticator"
	ErrMsgAuthenticatorNotFound ErrorMessage = "Authenticator or user not found"
	ErrMsgDeleteAuthenticator   ErrorMessage = "Failed to delete authenticator"
)
