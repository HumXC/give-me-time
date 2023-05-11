package main

import (
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/sunshineplan/imgconv"
)

// 这是运行在安卓系统里的程序
func main() {
	quality, isIgnoreErr, err := ParseArg(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}
	begain := time.Now()
	fmt.Sprintln(begain)
	c := exec.Command("screencap", "-p")
	buf, _ := c.StdoutPipe()
	c.Start()
	if err != nil && !isIgnoreErr {
		panic(err)
	}
	p, err := png.Decode(buf)
	if err != nil {
		panic(err)
	}

	begain = time.Now()
	err = jpeg.Encode(null, p, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(err)
	}

	// -----
	begain = time.Now()
	imgconv.Write(null, p, &imgconv.FormatOption{Format: imgconv.JPEG, EncodeOption: []imgconv.EncodeOption{
		imgconv.Quality(10),
	}})
}

// 第一个返回值是 [JPEG 的质量]，第二个返回值是 [是否忽略错误]
// 当第一个参数是 1-100 的整数，第二个参数是 bool 值
func ParseArg(args []string) (int, bool, error) {
	ignoreErr := false
	e := errors.New(
		"parameter error, usage: screencap [JPEG quality] [is ignore error]\n" +
			"- [JPEG quality]: ranges from 1 to 100 inclusive\n" +
			"- [is ignore error] true or false",
	)
	if len(args) < 2 {
		return 0, false, e
	}
	quality, err := strconv.Atoi(args[1])
	if err != nil || quality < 1 || quality > 100 {
		return 0, false, e
	}

	if len(args) > 2 {
		ignoreErr, err = strconv.ParseBool(args[2])
		if err != nil {
			return 0, false, e
		}
	}

	return quality, ignoreErr, nil
}

func Cmd(name string, arg ...string) (io.ReadCloser, error) {
	var err error
	c := exec.Command(name, arg...)
	stdout, _ := c.StdoutPipe()
	stderr, _ := c.StderrPipe()
	err = c.Run()
	if err != nil {
		errMsg, _ := io.ReadAll(stderr)
		return nil, errors.New(err.Error() + " :" + string(errMsg))
	}
	return stdout, nil
}
