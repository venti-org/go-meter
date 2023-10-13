package gmeter

import "net/http"

type Request struct {
	ID  int
	Req *http.Request
}
