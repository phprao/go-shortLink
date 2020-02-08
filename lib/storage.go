package lib

type Storage interface {
	Shorten(url string, exp int64) (string, error)
	ShortLinkInfo(link string) (interface{}, error)
	UnShorten(link string) (string, error)
}