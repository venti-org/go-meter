package gmeter

type DriverConfig struct {
	Concurrency   int
	Count         int
	Path          string
	Api           string
	ExtraJsonPath string
}

type ClientConfig struct {
	Api string
	Count int
}