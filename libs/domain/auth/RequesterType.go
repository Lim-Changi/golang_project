package auth

type RequesterType string

const (
	RequesterAdmin RequesterType = "admin"
	RequesterUser  RequesterType = "user"
)

var RequesterTypeList = []RequesterType{
	RequesterAdmin,
	RequesterUser,
}

func IsValidRequesterType(requesterType RequesterType) bool {
	switch requesterType {
	case RequesterAdmin:
		fallthrough
	case RequesterUser:
		return true
	default:
		return false
	}
}
