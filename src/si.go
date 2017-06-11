package main


import (
	"fmt"
    "image"
    "image/color"
    "image/png" // register the PNG format with the image package
    "os"
    "math"
)
    
func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {
	infile, err := os.Open(os.Args[1])
    check(err)
    
    defer infile.Close()
    
    src, err := png.Decode(infile)
    check(err)

    sobelX_kernel3x3 := [3][3]int{
 		{-1, 0, 1},
 		{-2, 0, 2},
 		{-1, 0, 1},
 	}

    sobelY_kernel3x3 := [3][3]int{
 		{ 1,  2,  1},
 		{ 0,  0,  0},
 		{-1, -2, -1},
 	}
/*
    scharrX_kernel3x3 := [3][3]int{
 		{3, 0, -3},
 		{10, 0, -10},
 		{3, 0, -3},
 	}

    scharrY_kernel3x3 := [3][3]int{
 		{ 3,  10,  3},
 		{ 0,  0,  0},
 		{-3, -10, -3},
 	}
*/
  
    kernelX := sobelX_kernel3x3
    kernelY := sobelY_kernel3x3
    const kernel_size int = len(kernelX)
    const half_kernel_size int = kernel_size / 2

    dimension := src.Bounds().Max
 	width := dimension.X
 	height := dimension.Y

 	grayImage := image.NewGray(image.Rect(0, 0, width, height))
 	for y := 1; y < height; y++ {
 		for x := 1; x < width; x++ {
 			oldPixel := src.At(x, y)
            pixel := color.GrayModel.Convert(oldPixel)
            grayImage.Set(x, y, pixel)
 		}
 	}

 	newImage := image.NewGray(image.Rect(0, 0, width, height))
 	
 	var SIsum int64
 	var SIrm int64


 	for y := 1; y < height - half_kernel_size; y++ {
 		for x := 1; x < width - half_kernel_size; x++ {

 			var magX float64
 			var magY float64

 			for a := 0; a < kernel_size; a++ {
 				for b := 0; b < kernel_size; b++ {
 					xn := x + a - half_kernel_size
 					yn := y + b - half_kernel_size
 					
 					idx := xn + yn * width
 					
 					magX += float64(grayImage.Pix[idx]) * float64(kernelX[a][b])
 					magY += float64(grayImage.Pix[idx]) * float64(kernelY[a][b])
 				}
 			}

 			SIr := int64(math.Sqrt(float64((magX * magX) + (magY * magY))))
            SIsum += SIr 
            SIrm += (SIr * SIr)

 			newImage.SetGray(x, y, color.Gray{uint8(SIr)})
 		}
 	}
 	
 	pixel := width * height
//  fmt.Printf("width=%d\n", width)
//  fmt.Printf("height=%d\n", height)
//  fmt.Printf("pixel=%d\n", pixel)
 	
 	SImean := (1.0 / float64(pixel)) * float64(SIsum)
// 	SIrms := math.Sqrt((1.0 / float64(pixel)) * float64(SIrm))
 	
//  fmt.Printf("SIsum=%dd\n", SIsum)
//  fmt.Printf("SIrm=%dd\n", SIrm)
 	fmt.Printf("SImean=%f\n", SImean)
//  fmt.Printf("SIrms=%f\n", SIrms)
 	
 	saveFile, err := os.Create("./edges.png")
 	check(err)
    defer saveFile.Close()
 	err = png.Encode(saveFile, newImage)
 	check(err)
}
