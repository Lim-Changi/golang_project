package readrequests

import (
	"golang_project/libs/domain/requests"
	config "golang_project/services/userSvc/internal"
	"log"
)

// Requests : 읽기 요청 맵
var Requests = requests.NewReadRequests()

func init() {
	log.Printf("Init %s readrequests package.\n", config.SVCName)
}

// Init :
func Init() {
	var dummy interface{}
	_ = dummy
	// Dummy Init -  꼭 필요함
}
