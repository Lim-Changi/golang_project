package readrequests

import (
	"context"
	"golang_project/libs/domain/auth"
	"golang_project/libs/domain/requests"
	"reflect"

	"github.com/pkg/errors"
)

// GetUserEntityRequest : 특정 User 엔티티 조회
type GetUserEntityRequest struct {
	_      struct{}
	UserID string `jsonschema:"description=유저 아이디"`
}

// RequestMethodType :
func (*GetUserEntityRequest) RequestMethodType() requests.RequestMethodType {
	return requests.RequestMethodRead
}

// Validate :
func (t *GetUserEntityRequest) Validate() error {
	if reflect.ValueOf(t.UserID).IsZero() {
		return errors.New("UserID 가 입력되지 않았습니다")
	}
	return nil
}

func (t *GetUserEntityRequest) IsAllowedRequesterType(requesterType auth.RequesterType) bool {
	switch requesterType {
	case auth.RequesterAdmin:
		fallthrough
	case auth.RequesterUser:
		return true
	default:
		return false
	}
}

// GetUserEntityResponse :
type GetUserEntityResponse struct {
	Item *string `jsonschema:"description=User 엔티티"`
}

// GetCreateReadResponseInstanceFunc ...
func (t *GetUserEntityRequest) GetCreateReadResponseInstanceFunc() requests.CreateReadResponseInstanceFunc {
	return func() interface{} { return &GetUserEntityResponse{} }
}

// ReadRequest Map 등록
func init() {
	Requests.Register(&GetUserEntityRequest{}, "특정 User 엔티티 조회")
}

// ProcessReadRequest :
// TODO. Repo 준비되면 추가해주어야함
func (t *GetUserEntityRequest) ProcessReadRequest(ctx context.Context, svcName string) (response interface{}, err error) {
	//repo, ok := er.(*pssRepo.EntityRepository)
	//if !ok || nil == repo {
	//	return nil, errors.New("엔티티 저장소를 추출하지 못했습니다")
	//}
	//item, err := repo.GetUserEntity(ctx, t.ProdID)
	//if nil != err {
	//	return nil, errors.Wrapf(err, "%s User 엔티티를 쿼리하지 못했습니다", t.ProdID)
	//}
	item := "나는 유저입니다"
	return &GetUserEntityResponse{Item: &item}, nil
}
