package auth

// GetAccessibleResourcesInput ...
type GetAccessibleResourcesInput struct {
	_           struct{}
	UserID      string `jsonschema:"description=유저 아이디"`
	ServiceName string `jsonschema:"description=서비스명,example=product,example=payment"`
}

// GetAccessibleResourcesOutput ...
type GetAccessibleResourcesOutput struct {
	_         struct{}
	Resources []string `jsonschema:"description=자원 목록"`
}
