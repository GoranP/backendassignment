package usage

import (
	"github.com/docopt/docopt-go"
)

var usage = `weblayer

Usage:
  weblayer_api [--port=port] [--redis=ip] 
  weblayer_api -h | --help
  weblayer_api --version

Options:
  -h --help             Show this screen.
  --version             Show version.
  --port=port           Listening port of service
  --redis=ip            Redis server 
  `

func Port() string {
	arguments, _ := docopt.Parse(usage, nil, true, "weblayer 2.0", false)
	path := arguments["--port"]
	if path == nil {
		return "8888"
	}
	return path.(string)
}

func Redis() string {
	arguments, _ := docopt.Parse(usage, nil, true, "weblayer 2.0", false)
	path := arguments["--redis"]
	if path == nil {
		return "127.0.0.1"
	}

	return path.(string)
}
