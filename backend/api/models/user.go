package models

type MediaContainer struct {
	Accounts []Account `xml:"Account"`
}

type Account struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type User struct {
	Username string `xml:"username,attr"`
	Email    string `xml:"email,attr"`
	ID       string `xml:"id,attr"`
}
