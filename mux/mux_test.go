package mux

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func testHTTP(method, url string, status int, result string, t *testing.T) {
	var req *http.Request
	req, _ = http.NewRequest(method, url, nil)
	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:   1,
			DisableKeepAlives: true,
		},
	}
	if resp, err := client.Do(req); err != nil {
		t.Error(err)
	} else if resp.StatusCode != status {
		t.Error(resp.StatusCode)
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Error(err)
	} else if string(body) != result {
		t.Error(string(body))
	}
}

func testHTTPError(method, url string, t *testing.T) {
	var req *http.Request
	req, _ = http.NewRequest(method, url, nil)
	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:   1,
			DisableKeepAlives: true,
		},
	}
	if _, err := client.Do(req); err == nil {
		t.Error()
	}
}

func testHTTPTLS(method, url string, status int, result string, t *testing.T) {
	var req *http.Request
	req, _ = http.NewRequest(method, url, nil)
	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:   1,
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	if resp, err := client.Do(req); err != nil {
		t.Error(err)
	} else if resp.StatusCode != status {
		t.Error(resp.StatusCode)
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Error(err)
	} else if string(body) != result {
		t.Error(string(body))
	}
}

func TestRum(t *testing.T) {
	addr := ":8080"
	m := NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	done := make(chan struct{})
	go func() {
		m.Run(addr)
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	testHTTP("GET", "http://"+addr+"/", http.StatusOK, "Hello World", t)
	m.Close()
	<-done
}

func TestRumTLS(t *testing.T) {
	certFile := "server.crt"
	keyFile := "server.key"
	defer os.Remove(certFile)
	defer os.Remove(keyFile)
	cf, err := os.Create(certFile)
	if err != nil {
		t.Error()
	}
	cf.Write(testCertPEM)
	cf.Close()
	kf, err := os.Create(keyFile)
	if err != nil {
		t.Error()
	}
	kf.Write(testKeyPEM)
	kf.Close()
	addr := ":8080"
	m := NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	done := make(chan struct{})
	go func() {
		m.RunTLS(addr, certFile, keyFile)
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	testHTTPTLS("GET", "https://"+addr+"/", http.StatusOK, "Hello World", t)
	m.Close()
	<-done
}

func TestFastRouter(t *testing.T) {
	addr := ":8080"
	m := NewRouter()
	m.SetFast(true)
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	done := make(chan struct{})
	go func() {
		m.Run(addr)
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	testHTTP("GET", "http://"+addr+"/", http.StatusOK, "Hello World", t)
	m.Close()
	<-done
}

func TestRumPoll(t *testing.T) {
	addr := ":8080"
	m := NewRouter()
	m.SetPoll(true)
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	done := make(chan struct{})
	go func() {
		m.Run(addr)
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	testHTTP("GET", "http://"+addr+"/", http.StatusOK, "Hello World", t)
	m.Close()
	<-done
}

func TestFastRumPollTLS(t *testing.T) {
	certFile := "server.crt"
	keyFile := "server.key"
	defer os.Remove(certFile)
	defer os.Remove(keyFile)
	cf, err := os.Create(certFile)
	if err != nil {
		t.Error()
	}
	cf.Write(testCertPEM)
	cf.Close()
	kf, err := os.Create(keyFile)
	if err != nil {
		t.Error()
	}
	kf.Write(testKeyPEM)
	kf.Close()
	addr := ":8080"
	m := NewRouter()
	m.SetFast(true)
	m.SetPoll(true)
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	done := make(chan struct{})
	go func() {
		m.RunTLS(addr, certFile, keyFile)
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	testHTTPTLS("GET", "https://"+addr+"/", http.StatusOK, "Hello World", t)
	testHTTPError("GET", "http://"+addr+"/", t)
	m.Close()
	<-done
}

func TestListenAndServe(t *testing.T) {
	addr := ":8080"
	m := NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	done := make(chan struct{})
	go func() {
		ListenAndServe(addr, m)
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	testHTTP("GET", "http://"+addr+"/", http.StatusOK, "Hello World", t)
	DefaultServer.Close()
	<-done
}

func TestListenAndServeTLS(t *testing.T) {
	certFile := "server.crt"
	keyFile := "server.key"
	defer os.Remove(certFile)
	defer os.Remove(keyFile)
	cf, err := os.Create(certFile)
	if err != nil {
		t.Error()
	}
	cf.Write(testCertPEM)
	cf.Close()
	kf, err := os.Create(keyFile)
	if err != nil {
		t.Error()
	}
	kf.Write(testKeyPEM)
	kf.Close()
	addr := ":8080"
	m := NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	done := make(chan struct{})
	go func() {
		ListenAndServeTLS(addr, certFile, keyFile, m)
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	testHTTPTLS("GET", "https://"+addr+"/", http.StatusOK, "Hello World", t)
	DefaultServer.Close()
	<-done
}

func TestRun(t *testing.T) {
	addr := ":8080"
	m := NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	done := make(chan struct{})
	go func() {
		m.Run(addr)
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	if err := m.Run(addr); err == nil {
		t.Error(err)
	}
	m.Close()
	<-done
}

func TestRunTLS(t *testing.T) {
	certFile := "server.crt"
	keyFile := "server.key"
	defer os.Remove(certFile)
	defer os.Remove(keyFile)
	cf, err := os.Create(certFile)
	if err != nil {
		t.Error()
	}
	cf.Write(testCertPEM)
	cf.Close()
	kf, err := os.Create(keyFile)
	if err != nil {
		t.Error()
	}
	kf.Write(testKeyPEM)
	kf.Close()
	addr := ":8080"
	m := NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	done := make(chan struct{})
	go func() {
		m.RunTLS(addr, certFile, keyFile)
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	if err := m.RunTLS(addr, certFile, keyFile); err == nil {
		t.Error(err)
	}
	m.Close()
	<-done
}

func TestRunTLSMore(t *testing.T) {
	addr := ":8080"
	m := NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	if err := m.RunTLS(addr, "", ""); err == nil {
		t.Error(err)
	}
	m.Close()
}

func TestStrSliceContains(t *testing.T) {
	if !strSliceContains([]string{"http/1.1"}, "http/1.1") {
		t.Error()
	}
	if strSliceContains([]string{"http/1.0"}, "http/1.1") {
		t.Error()
	}
}

var testKeyPEM = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQClne2gm0eXs/fV
Ffk/pPncBqlVvSy4XXu+n7VC7uAVjWd5Qv2aID/ECJSAzoeHHrzKhwE35yVnUZYl
i0M+KFen6cZ1l/bJLZ2stTKjnVRLIWTEU0ySBWdbq5noKCqgu95/UEWYYspfuBVU
aPoFi9QJhGMy1Bqay5jK85O+k+6/ZDa80uJBU4x1Ii5um72Cno+nLoAHqZ21hW58
pHQSdjLwcitclB0I11vk5ukVaQ0Ex+khyTWOQUPVmlBlHt/MxppqnPw0VCEm/8ID
WbN0GknBJ6BMm72OI9rv7CXVKCB3p9bwTHWs+ZAc1i0FIbd/QRAUZ3R18xRy5vhp
5x+eCxkFAgMBAAECggEAcVTsGn6H9FDsJD4ALURI0v8f+5MTuioGlwCVOnNaloUI
LNvhzQmKeSxh6u7poVYwXOsbs1SSrOwQsiplXnCe40G/VBMkhrCXeGsadwofd7Yn
TY2vFTFrdDiufkaWPR2yeS7GISxhTjUbK4MWrJruA4E7nLm2Zv0x1pfuH3j95Kul
xzn+x5f39IlTtUMcPtlxmyIQJhP8IZCNMH1czMviEsVhse1HjlP6RPtcyG3vSUTl
sT+SFV+qVEQM8i/PsopEXmSp2c57wLZuPAfyZE2v52ZLayrvozWEwCzkPhXwBh8Z
YL9mWpFWMi21xm/6yGAsemt0wQSamVMud1dCJkt3lQKBgQDbSVuUfT0Y9yEIbsYa
xykm2b9h9EYRTrW1LTmFV7olSc2y4ODX749M2aVViS0sptSFH+16RrBRqx4znTWE
Zw4Ie00NYdudP+e50KQttbWlCWLLPW//6A0bh5osMtwVH/DeWh9D8Dv5p8z/sM04
ni9WnBEgZv9nCDBcHZr/kdGPVwKBgQDBWEjPRb9VbHkChk5P16F/Z28zlXOfLznQ
YJFnoViMClLoZ24lOFwwQsKRbmMGivz4Kk/ANMUsQQTdaWgHzwoNAfSR+Vka4V4O
VmtM/Ysn993AASHTnQ0+XJ1SgDkjRvXOHLAovzPe5fJN3x98s8kauAxDH/jSoTBY
zNafb8kNAwKBgC34e/1uH9/a9w7Ygilc9TCS139GgWXhdVBtEbD1ii2FAl1srVY3
f3gh7Ho5DsJcmxLpdZR7uzzzvZoea/i8zgFTL/0NNcgKfJ7EsZZvkIvvoslusMsa
L51E2TpI00COeKV0iTdp5njZB63Wt78V0mC6yRBc51AHNHQrNLI9NIxzAoGAYJXc
0YaTPXZRwTwctPAQYlfmiktXlib+Z6VGHhHQ7SxfHHBG6tsL2qQPFJ2jZGpcUxC4
BFmYQjI83urAKF56M+NNLPr2U2JQHTN4qN5j/NRq93bZUiA7tHKJuOQLJx9UveFA
8yoJHImckOAsl5wsk2dr4MCBCZ+ySjoR4VXzhscCgYEAhLL+pGWDpbKrXlzX3Rd7
HD8hQgS2ubWleYW8/36pDXYqJGEdI+LrJNeEBQsb1obn/EAwEImGVCqzliMZuWo0
B4LnxJwEFkFRbWtPGqi/ncMNwol5mBWKOUT1o3X+8Hnz/xdzlvkkI4y9XDGnw/7d
8Ma5+ITkoPjiBl57eyDgbNc=
-----END PRIVATE KEY-----
`)

var testCertPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIDjDCCAnQCCQC5nnhzVhCHoTANBgkqhkiG9w0BAQsFADCBhjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgMB0JlaWppbmcxEDAOBgNVBAcMB0JlaWppbmcxDjAMBgNVBAoM
BWhzbGFtMQ4wDAYDVQQLDAVoc2xhbTESMBAGA1UEAwwJaHNsYW0uY29tMR8wHQYJ
KoZIhvcNAQkBFhA3OTE4NzQxNThAcXEuY29tMCAXDTIxMDExMDE1MjUzOVoYDzMw
MjAwNTEzMTUyNTM5WjCBhjELMAkGA1UEBhMCQ04xEDAOBgNVBAgMB0JlaWppbmcx
EDAOBgNVBAcMB0JlaWppbmcxDjAMBgNVBAoMBWhzbGFtMQ4wDAYDVQQLDAVoc2xh
bTESMBAGA1UEAwwJaHNsYW0uY29tMR8wHQYJKoZIhvcNAQkBFhA3OTE4NzQxNThA
cXEuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApZ3toJtHl7P3
1RX5P6T53AapVb0suF17vp+1Qu7gFY1neUL9miA/xAiUgM6Hhx68yocBN+clZ1GW
JYtDPihXp+nGdZf2yS2drLUyo51USyFkxFNMkgVnW6uZ6CgqoLvef1BFmGLKX7gV
VGj6BYvUCYRjMtQamsuYyvOTvpPuv2Q2vNLiQVOMdSIubpu9gp6Ppy6AB6mdtYVu
fKR0EnYy8HIrXJQdCNdb5ObpFWkNBMfpIck1jkFD1ZpQZR7fzMaaapz8NFQhJv/C
A1mzdBpJwSegTJu9jiPa7+wl1Sggd6fW8Ex1rPmQHNYtBSG3f0EQFGd0dfMUcub4
aecfngsZBQIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQBjnAnDA6X0LrLEbe5ZYRVF
5FhOZdV7TIGcEALag1MeAbGNFqsbXsQ6TF7aabGrX9kBesD+CZLyiynIS5cN62Vk
WMTQ1yXT9mcjBdVmRNWp7ty6jsr/iy7Ry08qlCjPuxbd2iDsOvwuBlq48+HQwKAs
jMRKday/IBXJ4uniyZnubWHvo6Ao/JsTHxlRWGzTXrRuyVONHDTTaivb5nSErHOP
tvdPWDGkbeztGj0hKJ6K7/sApsTyyzmJcZIE7MlEltWOnz5B1mZPJ4WkUz4//W+g
z1FVhxowNtorq+Ir3kQFpknqtBYqzapl4FdyGEP4JC41pYs1iz3r48AthhLm0TXu
-----END CERTIFICATE-----
`)
