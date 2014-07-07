package testutils

import (

	"atlantis/router/logger"
	"net/http"
	"net/http/httptest"
)



func NewTestHAProxyLogRecord(getUrl string) (*logger.HAProxyLogRecord, *httptest.ResponseRecorder) {	
	r, _ := http.NewRequest("GET", getUrl, nil)
	w   := httptest.NewRecorder()
	return &logger.HAProxyLogRecord{
		ResponseWriter:		w,
		Request:		r,
	} , w 
}

