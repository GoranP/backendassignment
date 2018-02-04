package usage

import (
	"github.com/docopt/docopt-go"
)

var usage = `workerlayer

Usage:
  workerlayer [--port=port] [--redis=ip] 
  workerlayer -h | --help
  workerlayer --version

Options:
  -h --help             Show this screen.
  --version             Show version.
  --port=port           Listening port of service
  --redis=ip            Redis server 
  `

func Redis() string {
	arguments, _ := docopt.Parse(usage, nil, true, "workerlayer 2.0", false)
	path := arguments["--redis"]
	if path == nil {
		return "127.0.0.1"
	}

	return path.(string)
}
