package oauth

import (
"encoding/base64"
"bytes"
"http"
"fmt"
"crypto/hmac"
"strings"
"container/vector"
"time"
"rand"
"once"
"sort"
"io"
"os"
"net"
"bufio"
"xml"
)
type SignatureTypes int
const (
    HMACSHA1 = iota
    PLAINTEXT
    RSASHA1
)

type nopCloser struct {
        io.Reader
}

func (nopCloser) Close() os.Error { return nil }



type Consumer struct{
    ApiKey,ApiKeySecret string
    AccessToken,AccessTokenSecret string 
    site string
    request,access,authorize string
    requestToken,requestTokenSecret string
}


func NewConsumer(consumer_key string,consumerSecret string,site string,request string,access string ,authorize string) (c Consumer) {
    c.ApiKey = consumer_key
    c.ApiKeySecret = consumerSecret
    c.site = site
    c.request = request
    c.access = access
    c.authorize = authorize
    return
}

const oAuthVersion = "1.0"
const oAuthParameterPrefix = "oauth_"
const oAuthConsumerKeyKey = "oauth_consumer_key";
const oAuthCallbackKey = "oauth_callback";
const oAuthVersionKey = "oauth_version";
const oAuthSignatureMethodKey = "oauth_signature_method";
const oAuthSignatureKey = "oauth_signature";
const oAuthTimestampKey = "oauth_timestamp";
const oAuthNonceKey = "oauth_nonce";
const oAuthTokenKey = "oauth_token";
const oAuthTokenSecretKey = "oauth_token_secret";

const hMACSHA1SignatureType = "HMAC-SHA1";
const plainTextSignatureType = "PLAINTEXT";
const rSASHA1SignatureType = "RSA-SHA1";

const unreservedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.~";

type param struct {
    key string
    val *string
}

type paramVector struct {
    vector.Vector
}

func (h *paramVector) Less(i, j int) bool {
    p1,p2 := h.At(i).(param),h.At(j).(param)
    if p1.key == p2.key{
        return *p1.val < *p2.val
    }
    return p1.key < p2.key
}

func newParam(key string,val string) param{
    return param{key,&val}
}



func newParamNil(key string) param{
    return param{key,nil}
}


func getTimeStamp() string{
    return fmt.Sprintf("%d",time.Seconds())
}

func getNonce() string{
    once.Do(func (){
        rand.Seed(time.Seconds()+1)
    })
    var end,begin int64
    end =  9999999
    begin = 123400
    return fmt.Sprintf("%d",(rand.Int63()%(end-begin))+begin)
}

func urlencode(val string ) (ret string){
    for i:=0;i<len(val);i++{
        c := val[i]
        s := string(c)
        if strings.Index(unreservedChars,s) != -1{
            ret += s
        }else{
            ret += fmt.Sprintf("%%%s%s",string("0123456789ADCDEF"[c>>4]),string("0123456789ABCDEF"[c&15]))
        }
    }
    return
}


func genarateSigBase(url *http.URL,method string,consumerKey string,consumerSecret string,token string ,tokenSecret string,signatureType SignatureTypes) (timeStamp string,nonce string ,sig string) {

    timeStamp = getTimeStamp()
    nonce = getNonce()
    query := url.RawQuery
    if len(query) > 0 &&  query[0] == '?' {
        query = query[1:]
    }
    strs := strings.Split(query,"&",-1)
    var vec paramVector
    for _,str := range strs {
        if len(str) == 0 || strings.Index(str,oAuthParameterPrefix) == 0 {
            continue
        }
        if strings.Index(str,"=") > -1{
            tmp := strings.Split(str,"=",2)
            vec.Push(newParam(tmp[0],tmp[1]))
        }else{
            vec.Push(newParamNil(str))
        }
    }
    vec.Push(newParam(oAuthVersionKey,oAuthVersion))
    vec.Push(newParam(oAuthNonceKey,nonce))
    vec.Push(newParam(oAuthTimestampKey,timeStamp))
    vec.Push(newParam(oAuthSignatureMethodKey,hMACSHA1SignatureType))
    vec.Push(newParam(oAuthConsumerKeyKey,consumerKey))
    if len(token) > 0{
        vec.Push(newParam(oAuthTokenKey,token))
    }
    sort.Sort(&vec)
    var np,nurl string
    id := 0
    vec.Do(
        func (v interface{}) {
            p := v.(param)
            if p.val != nil {
                np+= fmt.Sprintf("%s=%s",p.key,*p.val)
            }else{
                np+= p.key
            }
            id++
            if  id < vec.Len() {
                np+= "&"
            }
        })
    //scheme://[userinfo@]host/path[?query][#fragment]
    nurl = fmt.Sprintf("%s://%s%s",url.Scheme,url.Host,url.Path)
    sig = fmt.Sprintf("%s&%s&%s", strings.ToUpper(method),
    urlencode(nurl),
    urlencode(np))
    return
}


func genarateSig(url string,method string,consumerKey string,consumerSecret string,token string ,tokenSecret string,signatureType SignatureTypes) (timeStamp string,nonce string,sig string ){
    u,err := http.ParseURL(url)
    if err != nil {
        return
    }
    key := fmt.Sprintf("%s&%s",consumerSecret,tokenSecret)
    h := hmac.NewSHA1([]byte(key))
    timeStamp,nonce,sbase := genarateSigBase(u,method,consumerKey,consumerSecret,token,tokenSecret,signatureType)
    h.Write([]byte(sbase))
    bb := &bytes.Buffer{}
    encoder := base64.NewEncoder(base64.StdEncoding,bb)
    encoder.Write(h.Sum())
    encoder.Close()
    sig = urlencode(bb.String())
    return
}

func (c *Consumer) genarateRequestURL() string{
    url := c.site + c.request
    timeStamp,nonce,sig := genarateSig(url,"GET",c.ApiKey,c.ApiKeySecret,"", "",HMACSHA1)
    return fmt.Sprintf("%s?oauth_consumer_key=%s&oauth_nonce=%s&oauth_timestamp=%s&oauth_signature_method=%s&oauth_version=%s&oauth_signature=%s",url,c.ApiKey,nonce,timeStamp,hMACSHA1SignatureType,oAuthVersion,sig)
}

func (c *Consumer) genarateAccessURL() string{
    url := c.site + c.access
    timeStamp,nonce,sig := genarateSig(url,"GET",c.ApiKey,c.ApiKeySecret,c.requestToken,c.requestTokenSecret,HMACSHA1)
    return fmt.Sprintf("%s?oauth_consumer_key=%s&oauth_nonce=%s&oauth_timestamp=%s&oauth_signature_method=%s&oauth_version=%s&oauth_signature=%s&oauth_token=%s",url,c.ApiKey,nonce,timeStamp,hMACSHA1SignatureType,oAuthVersion,sig,c.requestToken)
}


func parseResponse(out string)(ret map[string]string) {
    ret = make(map[string]string)
    if len(out) > 0{
        tmp := strings.Split(out,"&",-1)
        for _,item := range(tmp) {
            if strings.Index(item,"=") > -1 {
                tmp2 := strings.Split(item,"=",-1)
                ret[tmp2[0]] = tmp2[1]
            }else{
                ret[item] = ""
            }
        }
    }
    return
}


/** 
* @brief get request token and return a url for user authorization
* 
* @param Consumer 
* @param GetAccessURL( 
*/
func (c *Consumer)GetAccessURL()  string{
    url := c.genarateRequestURL()
    r,_, err := http.Get(url)
	if err != nil {
        return ""
	}
	if r.StatusCode != 200{
        return ""
	}
    out := &bytes.Buffer{}
    io.Copy(out,r.Body)
    r.Body.Close()
    ret := parseResponse(out.String())
    c.requestToken = ret[oAuthTokenKey]
    c.requestTokenSecret = ret[oAuthTokenSecretKey]
    return fmt.Sprintf("%s%s?%s=%s",c.site,c.authorize,oAuthTokenKey,c.requestToken)
}

/** 
* @brief after user authorzied the access url,use this function to get access token to do sth....evil
* 
* @param Consumer 
* @param GetAccessToken( 
*/
func (c *Consumer)GetAccessToken() bool {
    url := c.genarateAccessURL()
    r,_, err := http.Get(url)
	if err != nil {
        return false
	}
	if r.StatusCode != 200{
        return false
	}
    out := &bytes.Buffer{}
    io.Copy(out,r.Body)
    r.Body.Close()
    ret := parseResponse(out.String())
    c.AccessToken = ret[oAuthTokenKey]
    c.AccessTokenSecret = ret[oAuthTokenSecretKey]
    return true
}


//========================================this is some copy of golang:http package======================================
// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

// Used in Send to implement io.ReadCloser by bundling together the
// io.BufReader through which we read the response, and the underlying
// network connection.
type readClose struct {
	io.Reader
	io.Closer
}

// Send issues an HTTP request.  Caller should close resp.Body when done reading it.
//
// TODO: support persistent connections (multiple requests on a single connection).
// send() method is nonpublic because, when we refactor the code for persistent
// connections, it may no longer make sense to have a method with this signature.
func send(req *http.Request) (resp *http.Response) {
	if req.URL.Scheme != "http" {
		return nil
	}

	addr := req.URL.Host
	if !hasPort(addr) {
		addr += ":http"
	}
	conn, err := net.Dial("tcp", "", addr)
	if err != nil {
		return nil
	}

	err = req.Write(conn)
	if err != nil {
		conn.Close()
		return nil
	}

	reader := bufio.NewReader(conn)
	resp, err = http.ReadResponse(reader, req.Method)
	if err != nil {
		conn.Close()
		return nil
	}

	resp.Body = readClose{resp.Body, conn}

	return
}
//========================================end copy of golang:http package======================================


/** 
* @brief do some sth with the access key you get
* 
*/
func (c *Consumer)Request(method string,url string,header map[string] string,content string) (*http.Response) {
    timeStamp,nonce,sig := genarateSig(url,"POST",c.ApiKey,c.ApiKeySecret,c.AccessToken, c.AccessTokenSecret,HMACSHA1)
    head := fmt.Sprintf("OAuth realm=\"\", oauth_consumer_key=%s, oauth_nonce=%s, oauth_timestamp=%s, oauth_signature_method=%s, oauth_version=%s ,oauth_signature=%s, oauth_token=%s",c.ApiKey,nonce,timeStamp,hMACSHA1SignatureType,oAuthVersion,sig,c.AccessToken);
    req := new(http.Request)
    req.URL, _ = http.ParseURL(url)
    req.Header = make(map[string] string)
    req.Header["Authorization"] = head
    for key,val := range(header){
        req.Header[key] = val
    }
    req.Method = "POST"
    data := "<?xml version='1.0' encoding='UTF-8'?>" + content
    req.ContentLength = int64(len(data))
    req.Body = nopCloser{bytes.NewBufferString(data)}
    resp := send(req)
    return resp
}


func (c *Consumer)Save(f io.Writer){
    out := fmt.Sprintf(`<xml>
<apikey>%s</apikey>
<apikeysecret>%s</apikeysecret>
<accesstoken>%s</accesstoken>
<accesstokensecret>%s</accesstokensecret>
</xml>`,c.ApiKey,c.ApiKeySecret,c.AccessToken,c.AccessTokenSecret)
    f.Write([]byte(out))

}

func (c *Consumer)Load(f io.Reader){
    xml.Unmarshal(f,c)
}








