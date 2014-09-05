package douban

import (
	"testing"
	//"os"
	"fmt"
	"strconv"
	"time"
)

const kid = 45723922
const isbn = "9787115230270"

func TestUser(t *testing.T) {
	time.Sleep(1e6)
	fmt.Printf("test user ....\n")
	api := NewApi()
	user := api.GetUserById(kid)
	fmt.Println(user)
	if user == nil {
		t.Errorf("find no user\n")
	}
	uid, err := strconv.ParseInt(user.Id, 10, 0)
	if err != nil || uid != kid {
		t.Errorf("get wrong id\n")
	}
}

func TestContact(t *testing.T) {
	time.Sleep(1e6)
	fmt.Printf("test friends ....\n")
	api := NewApi()
	contact := api.GetContactById(kid)
	if contact == nil {
		t.Errorf("find no user\n")
	}
	fmt.Printf("you have %d contacts\n", len(contact.Entry))
}

func TestBook(t *testing.T) {
	time.Sleep(1e6)
	fmt.Printf("test book ....\n")
	api := NewApi()
	book := api.GetBookByIsbn(isbn)
	fmt.Println(book)
	if book == nil {
		t.Errorf("find no book\n")
	}
	if book.Title != "Python基础教程" {
		t.Errorf("get wrong book title\n")
	}
	if book.Pages != "471" {
		t.Errorf("get wrong book pages\n")
	}
	if book.Url != "http://book.douban.com/subject/4866934/" {
		t.Errorf("get wrong book url\n")
		fmt.Println(book.Url)
	}
	if book.Cover.Small != "http://img3.douban.com/spic/s4387251.jpg" {
		t.Errorf("get wrong book cover\n")
	}
}
