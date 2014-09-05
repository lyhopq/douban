package douban

type User struct {
	Id      string
	Name    string
	Created string
	Banned  bool `json:"is_banned"`
	Suicide bool `json:"is_suicide"`
	Avatar  string
}

type Contacts struct {
	Entry []User
}
