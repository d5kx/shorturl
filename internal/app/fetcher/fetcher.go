package fetcher

type Fetcher interface {
	Fetch() error
}
