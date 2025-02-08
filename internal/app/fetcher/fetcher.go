package fetcher

type Fetcher interface {
	Fetch() error
	//AddHandler(string, processor.Processor)
}
