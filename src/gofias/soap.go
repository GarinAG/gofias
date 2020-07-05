package main

import (
	"github.com/tiaguinho/gosoap"
	"log"
	"net/http"
	"time"
)

// Not correct WSDL in FIAS service.
// WSDL return http scheme, but need https
// Fix it with custom transport
type CustomTransport struct {
	transport http.Transport
}

func (ct *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "fias.nalog.ru" {
		req.URL.Scheme = "https"
	}
	return ct.transport.RoundTrip(req)
}

func executeSoap(soapAction string, params gosoap.Params) *gosoap.Response {
	tr := &CustomTransport{}
	httpClient := &http.Client{
		Timeout:   1500 * time.Millisecond,
		Transport: tr,
	}
	soap, err := gosoap.SoapClient(fiasServiceUrl, httpClient)
	if err != nil {
		log.Fatalf("SoapClient error: %s", err)
	}

	res, err := soap.Call(soapAction, params)
	if err != nil {
		log.Fatalf("Call error: %s\n%s", err, res)
	}

	return res
}
