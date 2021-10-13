package utils

import (
	"fmt"
	"net/http"
	"regexp"
)

type WrongPatternErr struct {
	pattern string
}

func (wpr *WrongPatternErr) Error() string {
	return fmt.Sprintf("Wrong pattern. given %s", wpr.pattern)
}

type NOZHTTPServer struct {
	handlerPatterns []http.HandlerFunc
	methodPatterns  []uint16
	pathPatterns    []string
}

func NewNOZHTTPServer() *NOZHTTPServer {
	return &NOZHTTPServer{
		[]http.HandlerFunc{},
		[]uint16{},
		[]string{},
	}
}

func (hc *NOZHTTPServer) HandleFunc(handler http.HandlerFunc, pathPattern string, methods ...uint16) {

	if pathPattern == "" {
		panic(&WrongPatternErr{pathPattern})
	}

	var method uint16 = Http_method_any
	isAnyMethod := len(methods) == 0

	if !isAnyMethod {
		method = methods[0]
	}

	hc.handlerPatterns = append(hc.handlerPatterns, handler)
	hc.methodPatterns = append(hc.methodPatterns, method)
	hc.pathPatterns = append(hc.pathPatterns, pathPattern)
}

func (hc *NOZHTTPServer) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	method := ParseMethodTypeByTypeString(r.Method)

	var hitedHandler http.HandlerFunc

	for i := len(hc.pathPatterns); i > 0; i-- {
		pathPattern := hc.pathPatterns[i-1]
		methodPattern := hc.methodPatterns[i-1]

		isMethodMatched := method == methodPattern || methodPattern == Http_method_any

		if !isMethodMatched {
			continue
		}

		pathMatched, err := regexp.MatchString(pathPattern, urlPath)
		if err != nil {
			panic(err)
		}

		if pathMatched {
			hitedHandler = hc.handlerPatterns[i-1]
			break
		}
	}

	defer hc.handleError(wr, r)

	if hitedHandler != nil {
		hitedHandler(wr, r)
	}
}
func (hc *NOZHTTPServer) handleError(wr http.ResponseWriter, r *http.Request) {
	hu := GenHandlerUtils(wr)

	RunIfOK(recover() != nil, hu.SendInternalServerErr)
}

const (
	Http_method_any = iota
	Http_method_get
	Http_method_set
	Http_method_patch
	Http_method_delete
	Http_method_new
	Http_method_action
	Http_method_option

	Http_method_post

	Http_method_auth
)

func ParseMethodTypeByTypeString(method string) uint16 {
	switch method {
	case "GET":
		return Http_method_get
	case "SET":
		return Http_method_set
	case "PATCH":
		return Http_method_patch
	case "DEL":
		return Http_method_delete
	case "NEW":
		return Http_method_new
	case "ACTION":
		return Http_method_action
	case "OPTION":
		return Http_method_option
	case "POST":
		return Http_method_post
	case "AUTH":
		return Http_method_auth
	default:
		return Http_method_any
	}
}
