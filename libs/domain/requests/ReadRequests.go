package requests

import "log"

var commonReadRequestMap = RequestMap{}

func registerReadRequest(emptyInstance ReadRequest, description string) {
	reqName := commonReadRequestMap.register(emptyInstance, description)
	log.Printf("commonReadRequestMap.registerReadRequest: %s 요청 등록 완료\n", reqName)
}

// ReadRequests ...
type ReadRequests struct {
	requestMap RequestMap
}

// Register ...
func (t *ReadRequests) Register(emptyInstance ReadRequest, description string) {
	t.requestMap.register(emptyInstance, description)
}

// GetRequestMap ...
func (t *ReadRequests) GetRequestMap() RequestMap {
	return (*t).requestMap
}

// GetCommonRequestMap ...
func (t *ReadRequests) GetCommonRequestMap() RequestMap {
	return t.requestMap.retrieveCommonRequestMap()
}

// GetLocalRequestMap ...
func (t *ReadRequests) GetLocalRequestMap() RequestMap {
	return t.requestMap.retrieveLocalRequestMap()
}

// NewReadRequests ...
func NewReadRequests() *ReadRequests {
	obj := &ReadRequests{
		requestMap: RequestMap{},
	}
	for k, v := range commonReadRequestMap {
		obj.requestMap[k] = v
	}
	return obj
}
