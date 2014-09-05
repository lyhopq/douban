package douban

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	kDefaultClient        = "go-douban"
	kDefaultClientURL     = ""
	kDefaultClientVersion = "0.1"
	kDefaultUserAgent     = "go-douban"
	kErr                  = "GoDouban Error: "
	kWarn                 = "GoDouban Warning: "
	kDefaultTimelineAlloc = 20
	apiKey                = "0d0a08e70f9be1eb271e1ab52ed3a5c1"
	_QUERY_PERSON         = "https://api.douban.com/v2/user/%d?apikey=%s"
	_QUERY_BOOK           = "https://api.douban.com/v2/book/isbn/%s?apikey=%s"
	_QUERY_FRIEND         = "http://api.douban.com/people/%d/contacts?apikey=%s&max-results=100"
	_QUERY_OAUTH          = "http://www.douban.com/service/auth/request_token?oauth_consumer_key=%s"
)

const (
	_STATUS = iota
	_SLICESTATUS
	_SLICESEARCH
	_USER
	_SLICEUSER
	_BOOL
	_ERROR
	_RATELIMIT
)

func parseResponse(response *http.Response) (string, error) {
	var b []byte
	b, _ = ioutil.ReadAll(response.Body)
	response.Body.Close()
	bStr := string(b)

	return bStr, nil
}

type Api struct {
	user           string
	pass           string
	proxy          string
	errors         chan error
	lastError      error
	client         string
	clientURL      string
	clientVersion  string
	userAgent      string
	receiveChannel interface{}
}

// Creates and initializes new Api objec
func NewApi() *Api {
	api := new(Api)
	api.init()
	return api
}

// Initializes a new Api object, called by NewApi()
func (self *Api) init() {
	self.errors = make(chan error, 16)
	self.receiveChannel = nil
	self.client = kDefaultClient
	self.clientURL = kDefaultClientURL
	self.clientVersion = kDefaultClientVersion
	self.userAgent = kDefaultUserAgent
}

func (self *Api) SetProxy(proxy string) {
	self.proxy = proxy
}

func (self *Api) getJsonFromUrl(url string) string {
	r, err := httpGet(url, self.user, self.pass, self.proxy)
	if err != nil {
		fmt.Printf(kErr + err.Error())
		return ""
	}

	data, err := parseResponse(r)
	if err != nil {
		fmt.Printf(kErr + err.Error())
		return ""
	}

	return data
}

func (self Api) GetUserById(id uint64) *User {
	var user User
	jsonString := self.getJsonFromUrl(fmt.Sprintf(_QUERY_PERSON, id, apiKey))
	buf := bytes.NewBufferString(jsonString)
	json.Unmarshal(buf.Bytes(), &user)
	return &user
}

func (self Api) GetContactById(id uint64) *Contacts {
	var con Contacts
	jsonString := self.getJsonFromUrl(fmt.Sprintf(_QUERY_FRIEND, id))
	buf := bytes.NewBufferString(jsonString)
	json.Unmarshal(buf.Bytes(), &con)
	return &con
}

func (self Api) GetBookByIsbn(isbn string) *Book {
	var book Book
	jsonString := self.getJsonFromUrl(fmt.Sprintf(_QUERY_BOOK, isbn, apiKey))
	buf := bytes.NewBufferString(jsonString)
	json.Unmarshal(buf.Bytes(), &book)
	return &book
}
