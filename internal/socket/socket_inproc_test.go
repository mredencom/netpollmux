package socket

import "testing"

func TestINPROC(t *testing.T) {
	address := ":9999"
	serverSock := NewINPROCSocket(nil)
	l, err := serverSock.Listen(address)
	defer l.Close()
	if err != nil {
		t.Error(err)
	}
	if _, err := serverSock.Listen(address); err == nil {
		t.Error(err)
	}
	l.Serve(nil)
	l.ServeConn(nil, nil)
	l.ServeData(nil, nil)
	l.ServeMessages(nil, nil)
	l.Close()
}
