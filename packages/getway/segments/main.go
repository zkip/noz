package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"noz.zkip.cc/utils"
)

type ValidationAuth struct {
	ok           bool
	which        uint64
	needTransfer bool
	resp         *http.Response
}

func getValidation(r *http.Request, authService *url.URL) *ValidationAuth {
	va := &ValidationAuth{}

	var err error
	client := &http.Client{}

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(content)
	r.Body = io.NopCloser(bytes.NewBuffer(content))

	req, _ := http.NewRequest("GET", authService.String(), body)
	req.URL.Path = "/"
	req.Header = r.Header.Clone()

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode == 901 {
		whichStr := resp.Header.Get("X-Authenticate-Which")
		which, _ := strconv.Atoi(whichStr)
		va.which = uint64(which)
		va.ok = true
		return va
	}

	va.needTransfer = true
	va.resp = resp
	return va
}

var necessary_transfer_headers = map[string]bool{
	"Content-Type": true,
}

func transferResponse(target *http.Response, dest http.ResponseWriter) {
	// copy headers
	for name, values := range target.Header {
		if !necessary_transfer_headers[name] {
			continue
		}
		dest.Header()[name] = values
	}

	_, err := io.Copy(dest, target.Body)
	if err != nil {
		panic(err)
	}
}

func main() {

	defer utils.Setup()()

	serveHost := utils.GetServeHost(9000)
	fmt.Println("Server active on: ", serveHost)

	err := http.ListenAndServe(serveHost, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		serviceURL, err := utils.ResolveService(r)
		if err != nil {
			rw.WriteHeader(http.StatusBadGateway)
			panic(err)
		}
		fmt.Println("Service URL: ", serviceURL, " Origin: ", r.URL)
		if utils.IsURLAuthNeednt(r.URL) {
			proxy := httputil.NewSingleHostReverseProxy(serviceURL)
			proxy.ServeHTTP(rw, r)

		} else {
			va := getValidation(r, utils.ResolveAuthServiceHost())

			if va.needTransfer {
				transferResponse(va.resp, rw)
			} else if va.ok {
				proxy := httputil.NewSingleHostReverseProxy(serviceURL)
				r.Header.Add("X-Authenticate-Which", utils.ToString(va.which))
				proxy.ServeHTTP(rw, r)
			} else {
				rw.WriteHeader(http.StatusUnprocessableEntity)
			}
		}
	}))
	fmt.Println(err)
}
