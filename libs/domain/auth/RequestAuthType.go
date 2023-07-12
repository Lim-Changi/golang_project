package auth

// RequestAuthType :
type RequestAuthType string

// RequestAuthTypes
const (
	RequestAuthIAM     RequestAuthType = "IAM"
	RequestAuthLambda  RequestAuthType = "User"
	RequestAuthUnknown RequestAuthType = "Unknown"
)

// IsValidRequestAuthType :
func IsValidRequestAuthType(authType RequestAuthType) bool {
	switch authType {
	case RequestAuthLambda:
		fallthrough
	case RequestAuthIAM:
		return true
	default:
		return false
	}
}
