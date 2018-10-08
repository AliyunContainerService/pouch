package httpstream

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/cri/stream/constant"
)

func TestIsUpgradeRequest(t *testing.T) {
	req, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("create new http request failed: %v", err)
	}

	if IsUpgradeRequest(req) {
		t.Fatalf("IsUpgradeRequest return true, but the request dose not include upgrade header")
	}

	req.Header.Set(http.CanonicalHeaderKey(HeaderConnection), HeaderUpgrade)

	if !IsUpgradeRequest(req) {
		t.Fatalf("IsUpgradeRequest return false, but the request dose include upgrade header")
	}
}

func TestNegotiateProtocol(t *testing.T) {
	clientProtocols := []string{constant.StreamProtocolV1Name, constant.StreamProtocolV2Name}
	serverProtocls := []string{constant.StreamProtocolV2Name, constant.StreamProtocolV3Name}

	protocol := negotiateProtocol(clientProtocols, serverProtocls)
	if protocol != constant.StreamProtocolV2Name {
		t.Fatalf("negotiateProtocol could not find the common protocol")
	}

	clientProtocols = []string{}
	protocol = negotiateProtocol(clientProtocols, serverProtocls)
	if protocol != "" {
		t.Fatalf("negotiateProtocol should return empty string when there is no common protocol")
	}
}

func TestHandshake(t *testing.T) {
	req, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("create new http request failed: %v", err)
	}

	w := httptest.NewRecorder()

	serverProtocls := []string{}

	// clientProtocols and serverProtocls both are empty.
	protocol, _ := Handshake(w, req, serverProtocls)
	if protocol != "" {
		t.Fatalf("clientProtocols and serverProtocls both are empty, returned protocol should be empty")
	}

	// clientProtocols is empty.
	serverProtocls = []string{constant.StreamProtocolV3Name}
	protocol, _ = Handshake(w, req, serverProtocls)
	if protocol != "" {
		t.Fatalf("clientProtocols is empty, returned protocol should be empty")
	}

	// serverProtocls is empty.
	req.Header.Add(http.CanonicalHeaderKey(HeaderProtocolVersion), constant.StreamProtocolV1Name)
	protocol, _ = Handshake(w, req, []string{})
	if protocol != "" {
		t.Fatalf("serverProtocls is empty, returned protocol should be empty")
	}

	// clientProtocols and serverProtocls have no common protocol.
	protocol, _ = Handshake(w, req, serverProtocls)
	if protocol != "" {
		t.Fatalf("no common protocol, returned protocol should be empty")
	}
	if w.Code != http.StatusForbidden {
		t.Fatalf("no common protocol, the status of response should be StatusForbidden")
	}
	if !reflect.DeepEqual(w.HeaderMap[http.CanonicalHeaderKey(HeaderAcceptedProtocolVersions)], serverProtocls) {
		t.Fatalf("no common protocol, the response should include the protocols which server support")
	}

	// successful handshake.
	w = httptest.NewRecorder()
	serverProtocls = []string{constant.StreamProtocolV2Name, constant.StreamProtocolV3Name}
	req.Header.Add(http.CanonicalHeaderKey(HeaderProtocolVersion), constant.StreamProtocolV2Name)

	protocol, _ = Handshake(w, req, serverProtocls)
	if protocol != constant.StreamProtocolV2Name {
		t.Fatalf("Handshake should return the common protocol")
	}
	if w.Header().Get(HeaderProtocolVersion) != constant.StreamProtocolV2Name {
		t.Fatalf("Handshake should write the negotiated protocol into the header of response")
	}

}
