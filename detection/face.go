package detection

import (
	"github.com/disintegration/imaging"
	pigo "github.com/esimov/pigo/core"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
)

func RotateImage(image image.Image, rotationAngle float64) image.Image {
	return imaging.Rotate(image, rotationAngle, color.White)
}

func findFaces(image image.Image, detectionScore float32) bool {
	cascadeFile, err := ioutil.ReadFile("./cascade/facefinder")
	if err != nil {
		log.Fatalf("Error reading the cascade file: %v", err)
	}

	src := pigo.ImgToNRGBA(image)

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}

	pigo := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err := pigo.Unpack(cascadeFile)
	if err != nil {
		log.Fatalf("Error reading the cascade file: %s", err)
	}

	angle := 0.0 // cascade rotation angle. 0.0 is 0 radians and 1.0 is 2*pi radians

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets := classifier.RunCascade(cParams, angle)

	// Calculate the intersection over union (IoU) of two clusters.
	dets = classifier.ClusterDetections(dets, 0.2)

	// return true if any detection dot score is above 50%
	for _, detection := range dets {
		if detection.Q > detectionScore {
			return true
		}
	}

	return false
}

func FindRotatedImage(srcImage image.Image, detectionScore float32) (image.Image, bool) {
	var tmpImage image.Image

	if findFaces(srcImage, detectionScore) {
		return srcImage, false
	}

	rotations := []float64{90, 180, 270}
	for n := 0; n < len(rotations); n++ {
		tmpImage = RotateImage(srcImage, rotations[n])
		if findFaces(tmpImage, detectionScore) {
			return tmpImage, true
		}
	}

	return srcImage, false
}
