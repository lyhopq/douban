package douban

type rat struct {
	Average string
	Max     int
	Min     int
	Num     int `json:"numRaters"`
}

type tag struct {
	Count int
	Name  string
}

type cover struct {
	Small  string
	Medium string
	Large  string
}
type Book struct {
	Id          string
	Title       string
	SubTitle    string
	AltTitle    string `json:"alt_title"`
	Author      []string
	AuthorIntro string `json:"author_intro"`
	Translator  []string
	Publisher   string
	PubDate     string
	Isbn10      string
	Isbn13      string
	Cover       cover `json:"images"`
	Price       string
	Pages       string
	Url         string `json:"alt"`
	Summary     string
	Catalog     string
	Tags        []tag
	Rating      rat `json:"rating"`
}
