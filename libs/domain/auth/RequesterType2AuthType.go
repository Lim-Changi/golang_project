package auth

func RequesterType2AuthType(requesterType RequesterType) RequestAuthType {
	switch requesterType {
	case RequesterAdmin:
		return RequestAuthIAM
	case RequesterUser:
		return RequestAuthLambda
	default:
		return RequestAuthUnknown
	}
}
