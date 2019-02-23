/*
	package mhttp is used to connect HPFS server and unmarshal result if error.
*/
package mhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"bs-2018/hpfs-client/errors"
)

var host = ""

func SetHost(h string) {
	host = h
}

type Rsp struct {
	Success bool        `json:"success"`
	Msg     interface{} `json:"msg"`
}

func req(urlPath, method string, token string, reqBody io.Reader) (interface{}, error) {
	client := &http.Client{}
	log.Printf("[MHTTP] %s:%s\n", method, urlPath)
	request, err := http.NewRequest(method, urlPath, reqBody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == 200 {
		result := &Rsp{}
		err := json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		if result.Success {
			return result.Msg, nil
		}
		e := result.Msg.(map[string]interface{})
		return nil, errors.ErrorS{Code: int(e["code"].(float64)), Msg: e["msg"]}
	}
	return nil, errors.ErrorS{Code: 0, Msg: "request filed by unknown error"}
}

func Get(urlPath string, query map[string]string, token string) (interface{}, error) {
	urlPath = host + urlPath
	qs := ""
	if query != nil {
		p := url.Values{}
		for k, v := range query {
			p.Add(k, v)
		}
		qs = p.Encode()
	}
	u, err := url.Parse(urlPath)
	if err != nil {
		return nil, err
	}
	u.RawQuery = qs
	return req(u.String(), "GET", token, bytes.NewBuffer([]byte("")))
}

// eg /v1/meta {username:"tab.zhang"} uid=tab.zhang
func Post(urlPath string, postForm []byte, query map[string]string, token string) (interface{}, error) {
	urlPath = host + urlPath
	qs := ""
	if query != nil {
		p := url.Values{}
		for k, v := range query {
			p.Add(k, v)
		}
		qs = p.Encode()
	}
	u, _ := url.Parse(urlPath)
	u.RawQuery = qs
	return req(u.String(), "POST", token, bytes.NewBuffer(postForm))
}

func Delete(urlPath string, postForm []byte, query map[string]string, token string) (interface{}, error) {
	urlPath = host + urlPath
	qs := ""
	if query != nil {
		p := url.Values{}
		for k, v := range query {
			p.Add(k, v)
		}
		qs = p.Encode()
	}
	u, _ := url.Parse(urlPath)
	u.RawQuery = qs
	return req(u.String(), "DELETE", token, bytes.NewBuffer(postForm))
}
