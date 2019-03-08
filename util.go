package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
)

type APIResponse struct {
	Success bool `json:"success"`
}


type APIRequest struct {
	file     []byte
	fileName string
	password string
}


type File struct{
	Name string
	Content []byte
	Header textproto.MIMEHeader
}

type Error struct {
	StatusCode int  `json:"status_code"`
	Success bool `json:"success"`
	ErrorType    string `json:"error"`
	Message string `json:"message"`
	Errors map[string][]string `json:"errors,omitempty"`
}

func NewUploadError()*Error {
	er := Error{}
	er.Success = false
	er.ErrorType = "upload_error"
	er.Message = "Erro upload"
	er.StatusCode = http.StatusBadRequest
	return &er
}

// WriteStructWithHeader - Generate APIGatewayProxyResponse to be return
func WriteStructWithHeader(data interface{}, statusCode int, headers map[string]string) (*events.APIGatewayProxyResponse, error) {

	bytes, err := json.Marshal(data)
	w := core.NewProxyResponseWriter()
	w.Header().Set("Content-Type", "application/json")

	for key, header := range headers {
		w.Header().Set(key, header)
	}

	w.WriteHeader(statusCode)
	w.Write(bytes)
	response, err := write(w)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func WriteStruct(data interface{}, statusCode int) (*events.APIGatewayProxyResponse, error) {

	bytes, err := json.Marshal(data)
	w := core.NewProxyResponseWriter()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(bytes)
	resp, err := write(w)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func WriteBinary(data []byte, headers map[string]string, statusCode int) (*events.APIGatewayProxyResponse, error) {

	w := core.NewProxyResponseWriter()
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(statusCode)
	w.Write(data)

	resp, err := write(w)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func WriteString(data string, statusCode int) (*events.APIGatewayProxyResponse, error) {

	w := core.NewProxyResponseWriter()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(data))

	resp, err := write(w)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func write(w *core.ProxyResponseWriter) (*events.APIGatewayProxyResponse, error) {

	resp, err := w.GetProxyResponse()
	if err != nil {
		return nil, err
	}
	return &resp, nil
}


// BasicAuth - Generate Base64 according with 'username' and 'password'.
func GetBasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " +base64.StdEncoding.EncodeToString([]byte(auth))
}

func ParseHttpRequest(r events.APIGatewayProxyRequest)(*http.Request,error){

	decodedBody := []byte(r.Body)

	params := url.Values{}
	for k, v := range r.QueryStringParameters {
		params.Add(k, v)
	}

	if r.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(r.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	req,err := http.NewRequest(r.HTTPMethod,r.Path+"?"+params.Encode(),bytes.NewReader(decodedBody))

	if err !=nil {
		return nil, err
	}

	req.Header = make(map[string][]string)

	for k, v := range r.Headers {
		if k == "content-type" || k == "Content-Type" {
			req.Header.Set(k, v)
		}
	}

	return req,nil
}


func GetFile(field string,req *http.Request)(*File,error){

	file, fh,err := req.FormFile(field)
	if err!=nil{
		return nil,err
	}
	buf := bytes.Buffer{}

	if _, err := io.Copy(&buf, file); err != nil {
		return nil, err
	}

	return &File{Name:fh.Filename,Content:buf.Bytes(),Header:fh.Header},nil
}

