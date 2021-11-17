package webserver

type WebserverProvider interface {
	Run() (err error)
}
