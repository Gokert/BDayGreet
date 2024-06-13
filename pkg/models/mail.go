package models

type Mail struct {
	To      string
	Subject string
	Body    string
}

type MailConfigServer struct {
	AddrEmail string
	Password  string
	Port      int
	AddrHost  string
}
