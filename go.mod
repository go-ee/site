module github.com/go-ee/site

go 1.15

require github.com/urfave/cli/v2 v2.3.0

require (
	github.com/go-ee/utils v0.0.0-20201104184309-5b62a7627986
	github.com/sirupsen/logrus v1.7.0
)

replace github.com/go-ee/utils => ../utils/
