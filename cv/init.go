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
// 如果 tmpl 图像中存在透明度图层，将会使用透明的部分作为遮罩
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

	// 将图像转换为灰度图像
	gocv.CvtColor(img, &grayImg, gocv.ColorBGRToGray)
	gocv.CvtColor(tmpl, &grayTmpl, gocv.ColorBGRToGray)
	// 如果有透明度通道，就将透明部分作为遮罩
	if tmpl.Channels() == 4 {
		// 提取alpha通道
		alpha := gocv.NewMat()
		gocv.ExtractChannel(tmpl, &alpha, 3)
		gocv.Threshold(alpha, &mask, 0, 255, gocv.ThresholdBinaryInv)
		// w := gocv.NewWindow("Input")
		// w.IMShow(mask)
	}

	gocv.MatchTemplate(grayImg, grayTmpl, &result, gocv.TmCcoeffNormed, mask)

	_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)

	return maxVal, maxLoc, nil
}
