//
// Copyright 2010 cloud <cloud@douban>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// From gotweet - a command line twitter client by Dmitry Chestnykh
// modified by Bill Casarin
//

package douban

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"net/url"
	"strconv"
	"strings"
)

type readClose struct {
	io.Reader
	io.Closer
}

type badStringError struct {
	what string
	str  string
}

func (e *badStringError) Error() string { return fmt.Sprintf("%s %q", e.what, e.str) }

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func send(req *http.Request) (resp *http.Response, err error) {
	addr := req.URL.Host
	if !hasPort(addr) {
		addr += ":http"
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	err = req.Write(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	reader := bufio.NewReader(conn)
	resp, err = http.ReadResponse(reader, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}

	r := io.Reader(reader)
	if v := resp.Header.Get("Content-Length"); v != "" {
		n, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return nil, &badStringError{"invalid Content-Length", v}
		}
		r = io.LimitReader(r, n)
	}
	resp.Body = readClose{r, conn}

	return
}

func encodedUsernameAndPassword(user, pwd string) string {
	bb := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, bb)
	encoder.Write([]byte(user + ":" + pwd))
	encoder.Close()
	return bb.String()
}

func authGet(urlStr, user, pwd string) (r *http.Response, err error) {
	var req http.Request

	req.Header.Set("Authorization", "Basic "+encodedUsernameAndPassword(user, pwd))
	if req.URL, err = url.Parse(urlStr); err != nil {
		return
	}
	if r, err = send(&req); err != nil {
		return
	}
	return
}

// Post issues a POST to the specified URL.
//
// Caller should close r.Body when done reading it.
func authPost(urlStr, user, pwd, client, clientURL, version, agent, bodyType string,
	body io.Reader) (r *http.Response, err error) {
	var req http.Request
	req.Method = "POST"
	req.Body = body.(io.ReadCloser)
	req.Header = map[string][]string{
		"Content-Type":         {bodyType},
		"Transfer-Encoding":    {"chunked"},
		"User-Agent":           {agent},
		"X-Twitter-Client":     {client},
		"X-Twitter-Client-URL": {clientURL},
		"X-Twitter-Version":    {version},
		"Authorization":        {"Basic " + encodedUsernameAndPassword(user, pwd)},
	}

	req.URL, err = url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return send(&req)
}

//指定代理ip
func getTransportFieldURL(proxy_addr string) (transport *http.Transport) {
	url_i := url.URL{}
	url_proxy, _ := url_i.Parse(proxy_addr)
	transport = &http.Transport{Proxy: http.ProxyURL(url_proxy)}
	return
}

// Do an authenticated Get if we've called Authenticated, otherwise
// just Get it without authentication
func httpGet(urlStr, user, pass, proxy string) (*http.Response, error) {
	var r *http.Response
	var err error

	if user != "" && pass != "" {
		r, err = authGet(urlStr, user, pass)
	} else {
		transport := getTransportFieldURL(proxy)
		client := &http.Client{Transport: transport}
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			log.Fatal(err.Error())
		}
		r, err = client.Do(req)
	}

	return r, err
}

// Do an authenticated Post if we've called Authenticated, otherwise
// just Post it without authentication
func httpPost(urlStr, user, pass, client, clientURL, version, agent,
	data string) (*http.Response, error) {
	var r *http.Response
	var err error

	body := bytes.NewBufferString(data)
	bodyType := "application/x-www-form-urlencoded"

	if user != "" && pass != "" {
		r, err = authPost(urlStr, user, pass, client, clientURL,
			version, agent, bodyType, body)
	} else {
		r, err = http.Post(urlStr, bodyType, body)
	}

	return r, err
}
