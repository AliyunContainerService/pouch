package httpstream

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsUpgradeRequest(t *testing.T) {
	request, err := http.NewRequest("GET", "www.123.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(http.CanonicalHeaderKey(HeaderConnection), HeaderUpgrade)
	if !IsUpgradeRequest(request) {
		t.Fatalf("expect return True from IsUpgradeRequest but return False")
	}

	request1, err := http.NewRequest("GET", "www.123.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	request1.Header.Set(http.CanonicalHeaderKey(HeaderConnection), "non-up")
	if IsUpgradeRequest(request1) {
		t.Fatalf("expect return False from IsUpgradeRequest but return True")
	}
}

func TestNegotiateProtocol(t *testing.T) {
	clientProtocols1 := []string{
		"1", "2", "3", "4"}
	serverProtocols1 := []string{
		"2", "4"}
	if code := negotiateProtocol(clientProtocols1, serverProtocols1); code != "2" {
		t.Fatalf("expect return 2, for 2 is the first matching protocol, but get %s", code)
	}

	clientProtocols2 := []string{
		"1", "2", "3", "4"}
	serverProtocols2 := []string{}

	if code := negotiateProtocol(clientProtocols2, serverProtocols2); code != "" {
		t.Fatalf("expect return nil, for server protocols are empty, but get %s", code)
	}

	clientProtocols3 := []string{}
	serverProtocols3 := []string{
		"1", "2", "3", "4"}

	if code := negotiateProtocol(clientProtocols3, serverProtocols3); code != "" {
		t.Fatalf("expect return nil, for client protocols are empty, but get %s", code)
	}
}

type fakeResponse struct {
	header http.Header
	code   int
}

func TestHandshake(t *testing.T) {
	// Normal case
	clientProtocols1 := []string{
		"1", "2"}
	serverProtocols1 := []string{
		"2", "4"}
	request1, err := http.NewRequest("GET", "www.123.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range clientProtocols1 {
		request1.Header.Add(http.CanonicalHeaderKey(HeaderProtocolVersion), v)
	}
	response1 := httptest.NewRecorder()

	Handshake(response1, request1, serverProtocols1)
	version1 := response1.Header().Get(HeaderProtocolVersion)
	if version1 != "2" {
		t.Fatalf("expect return 2, for 2 is the first matching protocol, but get %s", version1)
	}
	//404 Not Found
	clientProtocols2 := []string{
		"1", "2", "3", "4"}
	serverProtocols2 := []string{
		"5", "6", "7", "8"}
	request2, err := http.NewRequest("GET", "www.123.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range clientProtocols2 {
		request2.Header.Add(http.CanonicalHeaderKey(HeaderProtocolVersion), v)
	}

	response2 := httptest.NewRecorder()

	Handshake(response2, request2, serverProtocols2)
	code2 := response2.Code
	if code2 != http.StatusForbidden {
		t.Fatalf("expect get 404 Not Found, but get %d", code2)
	}
	versions2 := response2.HeaderMap[HeaderAcceptedProtocolVersions]
	for i, v := range versions2 {
		if v != serverProtocols2[i] {
			t.Fatalf("server protocol is not consist with the protocol in response header")
		}
	}
	//client list is nil
	serverProtocols3 := []string{
		"2", "4"}
	request3, err := http.NewRequest("GET", "www.123.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	response3 := httptest.NewRecorder()

	Handshake(response3, request3, serverProtocols3)
	version3 := response3.Header().Get(HeaderProtocolVersion)

	if version3 != "" {
		t.Fatalf("expect return nil but %s received", version3)
	}
	//server list is nil
	clientProtocols4 := []string{
		"2", "4"}
	serverProtocols4 := []string{}
	request4, err := http.NewRequest("GET", "www.123.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range clientProtocols4 {
		request4.Header.Add(http.CanonicalHeaderKey(HeaderProtocolVersion), v)
	}

	response4 := httptest.NewRecorder()

	Handshake(response4, request4, serverProtocols4)
	version4 := response4.Header().Get(HeaderProtocolVersion)

	if version4 != "" {
		t.Fatalf("expect return nil but %s received", version4)
	}
}
