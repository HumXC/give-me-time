package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"strconv"
)

// 这是运行在安卓系统里的程序
func main() {
	quality, isIgnoreErr, err := ParseArg(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}
	buf, err := Cmd("screencap", "-p")
	if err != nil && !isIgnoreErr {
		panic(err)
	} else {
		if buf.Len() == 0 {
			return
		}
	}
	p, err := png.Decode(buf)
	if err != nil {
		panic(err)
	}
	err = jpeg.Encode(os.Stdout, p, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(err)
	}
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

func Cmd(name string, arg ...string) (*bytes.Buffer, error) {
	var err error
	c := exec.Command(name, arg...)
	stdout := bytes.NewBuffer(make([]byte, 0, 4096))
	stderr := &bytes.Buffer{}
	c.Stdout = stdout
	c.Stderr = stderr
	err = c.Run()
	if err != nil {
		return nil, errors.New(err.Error() + " :" + stderr.String())
	}
	return stdout, nil
}
