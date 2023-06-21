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
	} else {
		rBody = nil
	}

	req, err := http.NewRequest(method, url, rBody)
	if err != nil {
		return "", err
	}
	req.Header = reqHeader

	if rBody != nil && len(contentType) != 0 {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{}
	if sslSkip {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bytesBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ret respResult
	outputOptions := strings.Split(outputOption, "|")
	outputLen := len(outputOptions)
	if outputLen == 0 {
		return "", errors.New("Invalid Output Option, Zero Option")
	} else {
		invalidOption := true
		if outputLen == 1 && outputOptions[0] == "FULL" {
			invalidOption = false
			outputOptions = []string{"PROTO", "STATUS", "HEADER", "BODY"}
			outputLen = 4
		}

		if contains(outputOptions, "PROTO") {
			invalidOption = false
			ret.Proto = resp.Proto
		}
		if contains(outputOptions, "STATUS") {
			invalidOption = false
			ret.Status = resp.Status
		} else if contains(outputOptions, "STATUS_CODE") {
			invalidOption = false
			ret.StatusCode = resp.StatusCode
		}
		if contains(outputOptions, "HEADER") {
			invalidOption = false
			ret.Header = resp.Header
		}
		if contains(outputOptions, "BODY") {
			invalidOption = false
			switch bodyOption {
			case "txt":
				ret.Body = string(bytesBody)
			case "b64":
				ret.Body = base64.StdEncoding.EncodeToString(bytesBody)
			case "hex":
				ret.Body = hex.EncodeToString(bytesBody)
			default:
				return "", errors.New("Invalid Body Option")
			}
		}

		if invalidOption {
			return "", errors.New("Invalid Output Option, " + fmt.Sprintf("(%v)", outputOptions))
		}
	}

	jBuffer := &bytes.Buffer{}
	encoder := json.NewEncoder(jBuffer)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(ret)
	if err != nil {
		return "", err
	}

	return string(jBuffer.Bytes()), nil
}

//export http_raw_init
func http_raw_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.int {
	if args.arg_count < 3 {
		msg := `
		http_raw(method string, url string, body string, option ...string) requires method, url, body argment
		` + optionDescription
		C.strcpy(message, C.CString(msg))
		return 1
	}
	return 0
}

//export http_raw
func http_raw(initid *C.UDF_INIT, args *C.UDF_ARGS, result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {
	gArg_count := uint(args.arg_count)

	var ret string
	var err error
	gArgs := ((*[arrLength]*C.char)(unsafe.Pointer(args.args)))[:gArg_count:gArg_count]
	method := C.GoString(*args.args)
	switch method {
	case "GET":
		if gArg_count == 3 {
			ret, err = httpRaw(method, C.GoString(gArgs[1]), "", "", nil)
		} else {
			ret, err = httpRaw(method, C.GoString(gArgs[1]), "", "", gArgs[3:])
		}
	default:
		if gArg_count == 3 {
			ret, err = httpRaw(method, C.GoString(gArgs[1]), "", C.GoString(gArgs[2]), nil)
		} else {
			ret, err = httpRaw(method, C.GoString(gArgs[1]), "", C.GoString(gArgs[2]), gArgs[3:])
		}
	}

	if err != nil {
		ret = err.Error()
	}

	result = C.CString(ret)
	*length = uint64(utf8.RuneCountInString(ret))
	return result
}

//export http_get_init
func http_get_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.int {
	if args.arg_count == 0 {
		msg := `
		http_get(url string, option ...string) requires url argment
		` + optionDescription
		C.strcpy(message, C.CString(msg))
		return 1
	}

	return 0
}

//export http_get
func http_get(initid *C.UDF_INIT, args *C.UDF_ARGS, result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {
	gArg_count := uint(args.arg_count)

	var ret string
	var err error
	if gArg_count == 1 {
		ret, err = httpRaw("GET", C.GoString(*args.args), "", "", nil)
	} else {
		gArgs := ((*[arrLength]*C.char)(unsafe.Pointer(args.args)))[:gArg_count:gArg_count]
		ret, err = httpRaw("GET", C.GoString(*args.args), "", "", gArgs[1:])
	}

	if err != nil {
		ret = err.Error()
	}

	result = C.CString(ret)
	*length = uint64(utf8.RuneCountInString(ret))
	return result
}

//export http_post_init
func http_post_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.int {
	if args.arg_count < 3 {
		msg := `
		http_post(url string, contentType string, body string, option ...string) requires url, contentType, body argment
		` + optionDescription
		C.strcpy(message, C.CString(msg))
		return 1
	}
	return 0
}

//export http_post
func http_post(initid *C.UDF_INIT, args *C.UDF_ARGS, result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {
	gArg_count := uint(args.arg_count)

	var ret string
	var err error
	gArgs := ((*[arrLength]*C.char)(unsafe.Pointer(args.args)))[:gArg_count:gArg_count]
	if gArg_count == 3 {
		ret, err = httpRaw("POST", C.GoString(*args.args), C.GoString(gArgs[1]), C.GoString(gArgs[2]), nil)
	} else {
		ret, err = httpRaw("POST", C.GoString(*args.args), C.GoString(gArgs[1]), C.GoString(gArgs[2]), gArgs[3:])
	}

	if err != nil {
		ret = err.Error()
	}

	result = C.CString(ret)
	*length = uint64(utf8.RuneCountInString(ret))
	return result
}

//export http_help_init
func http_help_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.int {
	return 0
}

//export http_help
func http_help(initid *C.UDF_INIT, args *C.UDF_ARGS, result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	msg := `
	Method List.
	http_raw(method string, url string, body string, option ...string) requires method, url, body argment
	http_get(url string, option ...string) requires url argment
	http_post(url string, contentType string, body string, option ...string) requires url, contentType, body argment

	` + optionDescription

	result = C.CString(msg)
	*length = uint64(utf8.RuneCountInString(msg))
	return result
}

func main() {
}
