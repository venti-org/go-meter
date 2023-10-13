package gmeter

type DriverConfig struct {
	Concurrency            int
	Skip                   int
	SkipError              bool
	ClientConfig           ClientConfig
	RequestGeneratorConfig RequestGeneratorConfig
}

type ClientConfig struct {
	Count int
	Proxy string
}

type RequestGeneratorConfig struct {
	Headers       []string
	Method        string
	Url           string
	UrlsPath      string
	Body          string
	BodyPath      string
	BodiesPath    string
	ExtraJsonPath string
}
