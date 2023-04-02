package cv_test

import (
	"fmt"
	"image"
	"testing"

	"github.com/HumXC/give-me-time/cv"
	"gocv.io/x/gocv"
)

func TestXxx(t *testing.T) {
	// 加载图像
	imgA := gocv.IMRead("../test/project1/screen.jpg", gocv.IMReadColor)
	imgB := gocv.IMRead("../test/project1/number.png", gocv.IMReadUnchanged)
	// 将图像转换为灰度图像
	grayA := gocv.NewMat()
	grayB := gocv.NewMat()
	gocv.CvtColor(imgA, &grayA, gocv.ColorBGRToGray)
	gocv.CvtColor(imgB, &grayB, gocv.ColorBGRToGray)
	// 创建一个掩码，将alpha通道中的不透明部分设置为1，透明部分设置为0
	mask := gocv.NewMat()

	// 如果有透明度通道，就将透明部分作为遮罩
	if imgB.Channels() == 4 {
		// 提取alpha通道
		alpha := gocv.NewMat()
		gocv.ExtractChannel(imgB, &alpha, 3)
		gocv.Threshold(alpha, &mask, 0, 255, gocv.ThresholdBinaryInv)
		// w := gocv.NewWindow("Input")
		// w.IMShow(mask)
	}
	// 使用模板匹配算法查找B.png在A.png中的位置
	result := gocv.NewMat()
	gocv.MatchTemplate(grayA, grayB, &result, gocv.TmCcoeffNormed, mask)

	// 找到最佳匹配位置
	_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)
	fmt.Println(maxVal)

	// 匹配的区域
	rect := image.Rectangle{
		Min: image.Point{X: maxLoc.X, Y: maxLoc.Y},
		Max: image.Point{X: maxLoc.X + imgB.Cols(), Y: maxLoc.Y + imgB.Rows()},
	}
	result = imgA.Region(rect)

	// 找出被遮罩的内容
	ignored := gocv.NewMat()
	result.CopyToWithMask(&ignored, mask)
	// 获取边界框
	contour := gocv.FindContours(mask, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	rect = gocv.BoundingRect(contour.At(0))
	// 裁剪图像
	cropped := ignored.Region(rect)
	// 显示输入图像和结果矩阵
	w1 := gocv.NewWindow("Input")
	w1.IMShow(cropped)
	w1.WaitKey(0)
}
func TestFind(t *testing.T) {
	want := image.Point{
		X: 103,
		Y: 174,
	}
	big := gocv.IMRead("test/big.png", gocv.IMReadUnchanged)
	small := gocv.IMRead("test/small.png", gocv.IMReadUnchanged)
	_, p, err := cv.Find(big, small)
	if err != nil {
		t.Fatal(err)
		return
	}
	if !p.Eq(want) {
		t.Errorf("want: %v, got: %v", want, p)
		return
	}
}

func BenchmarkFind(b *testing.B) {
	big := gocv.IMRead("../test/big.png", gocv.IMReadUnchanged)
	small := gocv.IMRead("../test/small.png", gocv.IMReadUnchanged)
	for i := 0; i < b.N; i++ {
		_, _, err := cv.Find(big, small)
		if err != nil {
			b.Fatal(err)
		}
	}
}
