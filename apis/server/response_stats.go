package server

import "sync"

// RequestStats is the request dealing result,
// including all handled request number, 2xx3xx number, 4xx number and 5xx number
type RequestStats struct {
	sync.Mutex
	// req5xxCount is the count number of failed request which returns a status code of 5xx.
	req5xxCount uint64
	// req4xxCount is the count number of failed request which returns a status code of 5xx.
	req4xxCount uint64
	// req2xxCount is the count number of requests which returns a status code of 2xx and 3xx.
	req2xx3xxCount uint64
	// reqCount is the count number of all request.
	reqCount uint64
}

func (rs *RequestStats) getAllCount() uint64 {
	rs.Lock()
	defer rs.Unlock()
	return rs.reqCount
}

func (rs *RequestStats) get2xx3xxCount() uint64 {
	rs.Lock()
	defer rs.Unlock()
	return rs.req2xx3xxCount
}

func (rs *RequestStats) get4xxCount() uint64 {
	rs.Lock()
	defer rs.Unlock()
	return rs.req4xxCount
}

func (rs *RequestStats) get5xxCount() uint64 {
	rs.Lock()
	defer rs.Unlock()
	return rs.req5xxCount
}

func (rs *RequestStats) increaseReqCount() {
	rs.Lock()
	rs.reqCount++
	rs.Unlock()
}

func (rs *RequestStats) increase2xx3xxCount() {
	rs.Lock()
	rs.req2xx3xxCount++
	rs.Unlock()
}

func (rs *RequestStats) increase4xxCount() {
	rs.Lock()
	rs.req4xxCount++
	rs.Unlock()
}

func (rs *RequestStats) increase5xxCount() {
	rs.Lock()
	rs.req5xxCount++
	rs.Unlock()
}
