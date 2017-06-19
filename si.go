package main

import (
    "flag"
    "fmt"
    "image"
    "image/color"
    _ "image/jpeg" // register the PNG format with the image package
    "image/png" // register the PNG format with the image package
    "os"
    "strings"
    "math"
    "sync"
    "time"
)

var verbose *bool

func loadImage(name string) image.Image {
    infile, err := os.Open(name)
    check(err)
    defer infile.Close()

    src, _, err := image.Decode(infile)
    check(err)
    return src
}


func main() {
    
    if len(os.Args) == 1 {
        usage(os.Args[0])
        return
    }

    siImgName := flag.String("o", "", "Name of the SI image - optional")
    verbose = flag.Bool("v", false, "Prints additional information to stderr")
    help := flag.Bool("h", false, "Prints this help")
    // kernelName := flag.String("k", "sobel", "Name of the SI kernel - optional")
    flag.Parse()
    srcImgName := os.Args[len(os.Args) - 1]
    
    if *help {
        usage(os.Args[0])
        return
    }
    
    
    startLoad := time.Now()
    src := loadImage(srcImgName)
    verbosePrintExecDuration(startLoad, "load image")

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
        { 3, 0,  -3},
        {10, 0, -10},
        { 3, 0,  -3},
    }

    scharrY_kernel3x3 := [3][3]int{
        { 3,  10,  3},
        { 0,   0,  0},
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
    
    sem := make(chan int, height)
   
    startGray := time.Now()
    for y := 0; y < height; y++ {
        go func(y int) {
            for x := 0; x < width; x++ {
                oldPixel := src.At(x, y)
                pixel := color.GrayModel.Convert(oldPixel)
                grayImage.Set(x, y, pixel)
            }
            sem <- 1
        }(y)
    }
    for y := 0; y < height; y++ {<- sem}
    verbosePrintExecDuration(startGray, "to gray")

    newImage := image.NewGray(image.Rect(0, 0, width, height))
     
    var SIsum int64
    var SIrm int64


    semk := make(chan int, height - half_kernel_size)
    var mu sync.Mutex

    startSi := time.Now()
    for y := 1; y < height - half_kernel_size; y++ {
        go func(y int) {
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
                mu.Lock()
                SIsum += SIr 
                SIrm += (SIr * SIr)
    
                newImage.SetGray(x, y, color.Gray{uint8(SIr)})
                mu.Unlock()
            }
            semk <- 1
        }(y)
    }
    for y := 1; y < height - half_kernel_size; y++ {<- semk}
    verbosePrintExecDuration(startSi, "si calc")

    pixel := width * height
    SImean := (1.0 / float64(pixel)) * float64(SIsum)

    if *verbose {
        SIrms := math.Sqrt((1.0 / float64(pixel)) * float64(SIrm))
        fmt.Fprintf(os.Stderr, "width=%d\n", width)
        fmt.Fprintf(os.Stderr, "height=%d\n", height)
        fmt.Fprintf(os.Stderr, "pixel=%d\n", pixel)
        fmt.Fprintf(os.Stderr, "SIsum=%dd\n", SIsum)
        fmt.Fprintf(os.Stderr, "SIrm=%dd\n", SIrm)
        fmt.Fprintf(os.Stderr, "SIrms=%f\n", SIrms)
    }

    fmt.Printf("%f", SImean)

    if strings.Compare(*siImgName, "") != 0 {
        saveFile, err := os.Create(*siImgName)
        check(err)
        defer saveFile.Close()
        err = png.Encode(saveFile, newImage)
        check(err)
    }
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

type Kernel struct {
    X [][]int
    Y [][]int
}

func header() {
    fmt.Fprintf(os.Stderr, "Spartial information of images.\n")
    fmt.Fprintf(os.Stderr, "Copyright (c) 2017 Sascha Kohlmann.\n")
}

func usage(prgName string) {
    fmt.Fprintf(os.Stderr, "usage: %s [options] image\n\n", prgName	)
    header()
    fmt.Fprintf(os.Stderr, "\nOptions:\n")
    fmt.Fprintf(os.Stderr, "  -h       : prints this help\n")
    fmt.Fprintf(os.Stderr, "  -k name  : name of the kernel to use. Default: sobel\n")
    fmt.Fprintf(os.Stderr, "  -o name  : stores a control image with <name>\n")
    fmt.Fprintf(os.Stderr, "  -v       : prints additional information on stderr\n")
}

func verbosePrintExecDuration(t time.Time, prefix string) {
    if *verbose {
        fmt.Fprintf(os.Stderr, "%s: %d\n", prefix, time.Since(t).Nanoseconds())
    }
}
