package main

// #include <stdio.h>
// #include <sys/types.h>
// #include <sys/stat.h>
// #include <stdlib.h>
// #include <string.h>
// #include <mysql.h>
import "C"
import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode/utf8"
	"unsafe"
	"time"
)

type respResult struct {
	Proto      string              `json:",omitempty"`
	Status     string              `json:",omitempty"`
	StatusCode int                 `json:",omitempty"`
	Header     map[string][]string `json:",omitempty"`
	Body       string              `json:",omitempty"`
}

const optionDescription = `option:
-b	Define body input type.(hex: hexdecimal output. ex/[ascii]"Hello" -> 48656c6c6f, b64: base64 encoded, txt(default): text)
-B	Define body output type.(hex: hexdecimal output. ex/[ascii]"Hello" -> 48656c6c6f, b64: base64 encoded, txt(default): text)
-H	Pass custom headers to server (H)
-O	Define kind of result.(PROTO, STATUS or STATUS_CODE, HEADER, BODY(default), FULL) ex/-O PROTO|STATUS|HEADER|BODY equal -O FULL
-s	Define tls/ssl skip verified true / false
`
const arrLength = 1 << 30

func contains(slice []string, str string) bool {
	for _, n := range slice {
		if n == str {
			return true
		}
	}
	// index := sort.SearchStrings(slice, str)
	// if index < len(slice) {
	// 	return slice[index] == str
	// }

	return false
}

func httpRaw(method string, url string, contentType string, body string, options []*C.char) (string, error) {
	reqHeader := http.Header{}
	bodyOption := "txt"
	iBodyOption := "txt"
	outputOption := "BODY"
	sslSkip := false
	if options != nil {
		for _, opt := range options {
			option := strings.Split(C.GoString(opt), " ")

			switch option[0] {
			case "-H":
				header := strings.Split(strings.Join(option[1:], " "), ":")
				if len(header) != 2 {
					return "", errors.New("Invalid Header Option")
				}
				reqHeader.Add(header[0], header[1])
			case "-B":
				bodyOption = option[1]
			case "-b":
				iBodyOption = option[1]
			case "-O":
				outputOption = option[1]
			case "-s":
				sslSkip = option[1] == "true"
			}
		}
	}

	var rBody io.Reader
	if len(body) > 0 {
		switch iBodyOption {
		case "txt":
			rBody = strings.NewReader(body)
		case "b64":
			b64Datas, err := base64.StdEncoding.DecodeString(body)
			if err != nil {
				return "", err
			}
			rBody = bytes.NewReader(b64Datas)
		case "hex":
			hexDatas, err := hex.DecodeString(body)
			if err != nil {
				return "", err
			}
			rBody = bytes.NewReader(hexDatas)
		}
	}

	req, err := http.NewRequest(method, url, rBody)
	if err != nil {
		return "", err
	}

	if len(contentType) > 0 {
		req.Header.Add("Content-Type", contentType)
	}

	for k, v := range reqHeader {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: sslSkip},
	}
	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respResult := respResult{
		Proto:      resp.Proto,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
	}
	switch outputOption {
	case "PROTO":
		return respResult.Proto, nil
	case "STATUS":
		return respResult.Status, nil
	case "STATUS_CODE":
		return fmt.Sprintf("%d", respResult.StatusCode), nil
	case "HEADER":
		headers, err := json.Marshal(respResult.Header)
		if err != nil {
			return "", err
		}
		return string(headers), nil
	}

	// outputOption == "BODY" || outputOption == "FULL"
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	respResult.Body = string(b)
	switch bodyOption {
	case "txt":
		break
	case "b64":
		b64Body := base64.StdEncoding.EncodeToString([]byte(respResult.Body))
		respResult.Body = b64Body
	case "hex":
		hexBody := hex.EncodeToString([]byte(respResult.Body))
		respResult.Body = hexBody
	}

	if outputOption == "FULL" {
		fullResult, err := json.Marshal(respResult)
		if err != nil {
			return "", err
		}
		return string(fullResult), nil
	}

	return respResult.Body, nil
}

func httpGet(url string, options []*C.char) (string, error) {
	return httpRaw("GET", url, "", "", options)
}

func httpPost(url string, contentType string, body string, options []*C.char) (string, error) {
	return httpRaw("POST", url, contentType, body, options)
}

func httpPut(url string, contentType string, body string, options []*C.char) (string, error) {
	return httpRaw("PUT", url, contentType, body, options)
}

func httpDelete(url string, options []*C.char) (string, error) {
	return httpRaw("DELETE", url, "", "", options)
}

//export HttpGet
func HttpGet(url *C.char, options []*C.char) *C.char {
	ret, err := httpGet(C.GoString(url), options)
	if err != nil {
		return C.CString(fmt.Sprintf("error:%v", err))
	}
	return C.CString(ret)
}

//export HttpPost
func HttpPost(url *C.char, contentType *C.char, body *C.char, options []*C.char) *C.char {
	ret, err := httpPost(C.GoString(url), C.GoString(contentType), C.GoString(body), options)
	if err != nil {
		return C.CString(fmt.Sprintf("error:%v", err))
	}
	return C.CString(ret)
}

//export HttpPut
func HttpPut(url *C.char, contentType *C.char, body *C.char, options []*C.char) *C.char {
	ret, err := httpPut(C.GoString(url), C.GoString(contentType), C.GoString(body), options)
	if err != nil {
		return C.CString(fmt.Sprintf("error:%v", err))
	}
	return C.CString(ret)
}

//export HttpDelete
func HttpDelete(url *C.char, options []*C.char) *C.char {
	ret, err := httpDelete(C.GoString(url), options)
	if err != nil {
		return C.CString(fmt.Sprintf("error:%v", err))
	}
	return C.CString(ret)
}

//export OptionDescription
func OptionDescription() *C.char {
	return C.CString(optionDescription)
}

func main() {
}


