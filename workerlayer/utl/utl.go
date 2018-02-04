package utl

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"runtime/debug"
	"strings"
	"time"
)

var (
	_level int
)

func init() {
	//TODO: read this flag from config
	_level = 2
}

////////////////////
// generic helpers
////////////////////

//serialize object to JSON
func JSON(p interface{}) []byte {
	r, e := json.Marshal(p)
	if e != nil {
		log.Println(string(r))
		INFO("127.0.0.1", "json", "utl.JSON", e.Error())
		return []byte{}
	}
	return r
}

//measure time elaped for a function or methods
//called at begining of function in defer
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	INFO("[time] "+name, elapsed)
}

//////////////////////////////////////////////////
// recover && stacktrace in case of panic helpers
//////////////////////////////////////////////////
type recoverFunc func()

//generic goroutine panic handler && logger with recover function as parametar
func HandleDefer(goname string, start time.Time, rec recoverFunc) {
	//case of panic - log panic
	if r := recover(); r != nil {
		LogRecover(goname, r)
		if rec != nil { //case recover func
			go rec()
		}
	}
}

func LogRecover(goname string, r interface{}) {
	// find out exactly what the error was and set err
	var err error
	switch x := r.(type) {
	case string:
		err = errors.New(x)
	case error:
		err = x
	default:
		err = errors.New(fmt.Sprintf("Unknown panic! (%v)", r))
	}
	// invalidate rep
	ERR("127.0.0.1", goname, "error", err.Error())

	//dump stack in log
	stack := string(debug.Stack())
	arr := strings.Split(stack, "\n")
	for _, l := range arr {
		ERR(l)
	}

}

///////////////////////////
//tracing helper functions
///////////////////////////

//defalut level 2 - includes error, warn and info
//level 1 - includes error, warn, info and notice
//level 0 - DEBUG - includes error, warn, info, notice and debug
func SetTraceLevel(level int) {
	_level = level
}

func ERR(v ...interface{}) {

	if _level < 5 {
		log.Printf("\033[1;4;31m[ERROR] %v \033[0m\n", strings.TrimRight(fmt.Sprintln(v...), "\n"))
	}

}

func WARN(v ...interface{}) {
	if _level < 4 {
		log.Printf("\033[1;33m[WARN] %v \033[0m\n", strings.TrimRight(fmt.Sprintln(v...), "\n"))
	}
}

func INFO(v ...interface{}) {
	if _level < 3 {
		log.Printf("\033[32m[INFO] %v \033[0m\n", strings.TrimRight(fmt.Sprintln(v...), "\n"))
	}
}

func NOTICE(v ...interface{}) {
	if _level < 2 {
		log.Printf("[NOTICE] %v\n", strings.TrimRight(fmt.Sprintln(v...), "\n"))
	}
}

func DEBUG(v ...interface{}) {
	if _level < 1 {
		log.Printf("\033[1;35m[DEBUG] %v \033[0m\n", strings.TrimRight(fmt.Sprintln(v...), "\n"))
	}
}

//////////////////////////////////////
// compression/unmcopmression heplers
///////////////////////////////////////
func GZIP(data []byte) []byte {
	defer TimeTrack(time.Now(), "gzip data")

	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)

	_, err := zw.Write(data)
	if err != nil {
		ERR("gzip", err)
	}
	if err := zw.Close(); err != nil {
		ERR("gzip close", err)
	}
	return buf.Bytes()
}

func GUNZIP(data []byte) []byte {
	defer TimeTrack(time.Now(), "gunzip data")

	var buf bytes.Buffer
	buf.Write(data)

	zr, err := gzip.NewReader(&buf)
	if err != nil {
		ERR("gunzip", err)
		return nil
	}
	if err := zr.Close(); err != nil {
		ERR("gunzip close", err)
		return nil
	}

	gunzipped, err := ioutil.ReadAll(zr)
	if err != nil {
		ERR("gunzip readall", err)
		return nil
	}
	return gunzipped
}
