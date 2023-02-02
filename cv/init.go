package cv

import (
	"errors"
	"image"

	"gocv.io/x/gocv"
)

var ErrIMEmpty = errors.New("empty image")
var ErrVTooLow = errors.New("value too low")

func Find(src, target string) (image.Point, error) {
	imsrc := gocv.IMRead(src, gocv.IMReadColor)
	if imsrc.Empty() {
		return image.Point{}, ErrIMEmpty
	}
	defer imsrc.Close()
	imtarget := gocv.IMRead(target, gocv.IMReadColor)
	if imtarget.Empty() {
		return image.Point{}, ErrIMEmpty
	}
	defer imtarget.Close()

	result := gocv.NewMat()
	defer result.Close()

	mask := gocv.NewMat()
	defer mask.Close()
	gocv.MatchTemplate(imsrc, imtarget, &result, gocv.TmCcoeffNormed, mask)

	_, max, _, maxLoc := gocv.MinMaxLoc(result)
	if max < 0.9 {
		return image.Point{}, ErrVTooLow
	}
	return maxLoc, nil
}
