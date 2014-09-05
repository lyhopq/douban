package oauth

import (
"testing"
"fmt"
"bufio"
"os"
)

//for douban
const site = "http://www.douban.com"
const requestTokenUri = "/service/auth/request_token"
const accessTokenUri = "/service/auth/access_token"
const authorizationUri = "/service/auth/authorize";
const miniblogUri = "http://api.douban.com/miniblog/saying"



const apiKey = "0cb76f8d92f0abc826264ea8b8d20358"
const apiKeySecret = "66f9e7a00c51eb46"


func TestRequestToken(t *testing.T){
    c := NewConsumer(apiKey,apiKeySecret,site,requestTokenUri,accessTokenUri,authorizationUri)
    out := c.GetAccessURL()
    if len(out) <= 0 {
        t.Errorf("err\n")
    }
    fmt.Printf("parse this url to your browser: %s\n",out)
    fmt.Printf("after your authorization\n")
    reader := bufio.NewReader(os.Stdin)
    reader.ReadString('\n')
    if (! c.GetAccessToken() ){
        t.Errorf("err\n")
    }


    resp := c.Request("POST","http://api.douban.com/miniblog/saying",map[string] string {
        "Content-Type":"application/atom+xml",
        }, `<entry xmlns:ns0="http://www.w3.org/2005/Atom" xmlns:db="http://www.douban.com/xmlns/"><content>this is just a test</content></entry>`,)
    if resp.StatusCode  != 201{
        fmt.Printf("err code %d\n",resp.StatusCode)
        t.Errorf("err\n")
    }
    f,_ := os.Open("access.xml",os.O_WRONLY | os.O_CREAT,0666)
    c.Save(f)
}

func TestReadToken(t *testing.T){
    c := NewConsumer("","",site,requestTokenUri,accessTokenUri,authorizationUri)
    f,_ := os.Open("access.xml",os.O_RDONLY,0)
    c.Load(f)
    resp := c.Request("POST","http://api.douban.com/miniblog/saying",map[string] string {
        "Content-Type":"application/atom+xml",
        }, `<entry xmlns:ns0="http://www.w3.org/2005/Atom" xmlns:db="http://www.douban.com/xmlns/"><content>this is just a test</content></entry>`,)
    if resp.StatusCode  != 201{
        fmt.Printf("err code %d\n",resp.StatusCode)
        t.Errorf("err\n")
    }

}
