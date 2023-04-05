package cv

import (
	"errors"
	"image"

	"gocv.io/x/gocv"
)

var ErrIMEmpty = errors.New("empty image")
var ErrVTooLow = errors.New("value too low")

// 使用模板匹配在 img 中匹配 tmpl
// 第一个返回值是 maxVal，第二个返回值是 maxLoc
func Find(img, tmpl gocv.Mat) (float32, image.Point, error) {
	grayImg := gocv.NewMat()
	grayTmpl := gocv.NewMat()
	result := gocv.NewMat()
	mask := gocv.NewMat()
	defer func() {
		grayImg.Close()
		grayTmpl.Close()
		result.Close()
		mask.Close()
	}()

	// 将图像转换为RGBA图像
	gocv.CvtColor(img, &grayImg, gocv.ColorBGRToRGBA)
	gocv.CvtColor(tmpl, &grayTmpl, gocv.ColorBGRToRGBA)
	gocv.MatchTemplate(grayImg, grayTmpl, &result, gocv.TmCcoeffNormed, mask)

	_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)

	return maxVal, maxLoc, nil
}
