package ctrd

import "testing"
import "net/url"

func TestHasPort(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		//测试用例
		{name: "test1", s: "fdafd", want: false},
		{name: "test2", s: "fdafd:3333", want: true},
		{name: "test3", s: "[12:21:34:13]:1000", want: true},
		{name: "test4", s: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasPort(tt.s)
			if err != tt.want {
				t.Errorf("HasPort() error = %v", err) //若不符合预期则报错
			}
		})
	}
}

func TestCanonicalAddr(t *testing.T) {
	type Test struct {
		name string
		url  *url.URL
		want string
	}

	test1 := Test{name: "test1", want: "www.example.com:8080"}
	test1.url, _ = url.Parse("http://www.example.com:8080")

	test2 := Test{name: "test2", want: ":"}
	test2.url, _ = url.Parse("www.example.com:8080")

	test3 := Test{name: "test3", want: "www.example.com:80"}
	test3.url, _ = url.Parse("http://www.example.com")

	test4 := Test{name: "test4", want: ":"}
	test4.url, _ = url.Parse("errorCase")

	test5 := Test{name: "test5", want: ":"}
	test5.url, _ = url.Parse("")

	var tests = [5]Test{test1, test2, test3, test4, test5}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canonicalAddr(tt.url)
			if got != tt.want {
				t.Errorf("canonicalAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUseProxy(t *testing.T) {
	tests := []struct {
		name     string
		hostPort string
		want     bool
	}{
		{name: "test1", hostPort: "", want: true},
		{name: "test2", hostPort: "errorCase", want: false},
		{name: "test3", hostPort: "google.com:80", want: true},
		{name: "test4", hostPort: "localhost:80", want: false},
		{name: "test5", hostPort: "127.0.0.1:80", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := useProxy(tt.hostPort)
			if got != tt.want {
				t.Errorf("useProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}
