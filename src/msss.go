package main

/*
 * Computes the saliency mean factor based on Radhakrishna Achanta and and S. Suesstrunk
 * salient detection algorithm published in "Saliency Detection using Maximum Symmetric Surround",
 * Proceedings of IEEE International Conference on Image Processing (ICIP), 2010.
 *
 * This code is a simple scripted go implementation of the C++ code, published at
 * http://ivrl.epfl.ch/page-75168-en.html
 *
 * Original code (c) 2010 Radhakrishna Achanta [EPFL]. All rights reserved.
 * Go code (c) 2017 Sascha Kohlmann
 */

import (
    "flag"
    "fmt"
    "image"
    "image/jpeg" // register the PNG format with the image package
    "image/png" // register the PNG format with the image package
    "os"
    "strings"
    "math"
)


type LAB struct {
    L []float64
    A []float64
    B []float64
}

func (l *LAB) InitLabSize(size int) {
    if size < 0 {
        panic("Unable to init LAB with size < 0")
    }
    l.L = make([]float64, size)
    l.A = make([]float64, size)
    l.B = make([]float64, size)
}

func imageToLab(src image.Image) *LAB {
    lab := new(LAB)
    
    bounds := src.Bounds()
    pixel := bounds.Max.X * bounds.Max.Y
    lab.InitLabSize(pixel)

    for x := 0; x < bounds.Max.X; x++ {
        for y := 0; y < bounds.Max.Y; y++ {
            
            color := src.At(x, y)
            sR, sG, sB, _ := color.RGBA()
            
            R := float64(sR) / 255.0
            G := float64(sG) / 255.0
            B := float64(sB) / 255.0
            
            var r, g, b float64
            
            if R <= 0.04045 {
                r = R / 12.92
            } else {
                r = math.Pow((R + 0.055) / 1.055, 2.4)
            }
            if G <= 0.04045 {
                g = G / 12.92
            } else {
                g = math.Pow((G + 0.055) / 1.055, 2.4)
            }
            if B <= 0.04045 {
                b = B / 12.92
            } else {
                b = math.Pow((B + 0.055) / 1.055, 2.4)
            }

            X := r * 0.4124564 + g * 0.3575761 + b * 0.1804375
            Y := r * 0.2126729 + g * 0.7151522 + b * 0.0721750
            Z := r * 0.0193339 + g * 0.1191920 + b * 0.9503041

            //------------------------
            // XYZ to LAB conversion
            //------------------------
            const epsilon float64 = 0.008856 //actual CIE standard
            const kappa float64 = 903.3      //actual CIE standard

            const Xr float64 = 0.950456 //reference white
            const Yr float64 = 1.0      //reference white
            const Zr float64 = 1.088754 //reference white

            xr := X / Xr
            yr := Y / Yr
            zr := Z / Zr

            var fx, fy, fz float64
            if xr > epsilon {
                fx = math.Pow(xr, 1.0 / 3.0)
            } else {
                fx = (kappa * xr + 16.0) / 116.0
            }
            if yr > epsilon {
                fy = math.Pow(yr, 1.0 / 3.0);
            } else {
                fy = (kappa * yr + 16.0) / 116.0
            }
            if zr > epsilon {
                fz = math.Pow(zr, 1.0 / 3.0)
            } else {
                fz = (kappa * zr + 16.0) / 116.0
            }

            idx := x + y * bounds.Max.X
            lab.L[idx] = 116.0 * fy - 16.0;
            lab.A[idx] = 500.0 * (fx - fy);
            lab.B[idx] = 200.0 * (fy - fz);
        }
    }    
    return lab
}

func gaussianSmooth(srcImg []float64, width int, height int, kernel []float64) []float64 {
    smoothImg := make([]float64, len(srcImg))
    tmpImg := make([]float64, len(srcImg))
    center := len(kernel) / 2
    
    rows := height
    cols := width

    // Blur in the x direction.    
    idx := 0
    for r := 0; r < rows; r++ {
        for c := 0; c < cols; c++ {
            var kernelsum float64
            var sum float64
            for cc := 	(-center); cc <= center; cc++ {
                if ((c + cc) >= 0) && ((c + cc) < cols) {
                    sum += srcImg[r * cols + (c + cc)] * kernel[center + cc]
                    kernelsum += kernel[center + cc]
                }
            }
            tmpImg[idx] = sum / kernelsum
            idx++
        } 
    }

    // Blur in the y direction.
    idx = 0
    for r := 0; r < rows; r++ {
        for c := 0; c < cols; c++ {
            var kernelsum float64
            var sum float64
            
            for rr := (-center); rr <= center; rr++ {
                if ((r + rr) >= 0) && ((r + rr) < rows) {
                    sum += tmpImg[(r + rr) * cols + c] * kernel[center + rr]
                    kernelsum += kernel[center + rr]
                }   
            }            
            
            smoothImg[idx] = sum / kernelsum;
            idx++;
        } 
    }

    return smoothImg 
}

func createIntegralImage(srcImg []float64, width int, height int) [][]float64 {
    intImg := make([][]float64, height)
    for i := range intImg {
        intImg[i] = make([]float64, width)
    }
    idx := 0
    
    for j := 0; j < height; j++ {
        var sumRow float64
        for k := 0; k < width; k++ {
            sumRow += srcImg[idx]
            idx++
            if 0 == j {
                intImg[j][k] = sumRow
            } else {
                intImg[j][k] = intImg[j - 1][k] + sumRow
            }
        }
    }
    
    return intImg
}

func getIntegralSum(intImg [][]float64, x1, y1, x2, y2 int) float64 {
    var sum float64
    
    if x1 - 1 < 0 && y1 - 1 < 0 {
        sum = intImg[y2][x2]
    } else if x1 - 1 < 0 {
        sum = intImg[y2][x2] - intImg[y1 - 1][x2]
    } else if y1 - 1 < 0 {
        sum = intImg[y2][x2] - intImg[y2][x1 - 1]
    } else {
        sum = intImg[y2][x2] + intImg[y1 - 1][x1 - 1] - intImg[y1 - 1][x2] - intImg[y2][x1 - 1]
    }
    
    return sum
}

func doNormalize(salMap []float64, width, height int) {
    maxValue := 0.0
    minValue := float64(1 << 30)
    
    size := width * height
    for i := 0; i < size; i++ {
//        fmt.Printf("salMap %v\n", salMap[i])
        if maxValue < salMap[i] {
            maxValue = salMap[i]
        }
        if minValue > salMap[i] {
            minValue = salMap[i]
        }
    }
    
//    fmt.Printf("max: %f - min: %f\n", maxValue, minValue)
    
    _range := maxValue - minValue
    if _range <= 0 {panic("Range lower 0")}
    
    for i := 0; i < size; i++ {
        salMap[i] = ((255.0 * (salMap[i] - minValue)) / _range)
//        fmt.Printf("Normalized: %v\n", salMap[i])
    }
    
}

func max(a, b int) int {
    if a <= b {
        return b
    }
    return a
}

func min(a, b int) int {
    if a <= b {
        return a
    }
    return b
}

func computeMaximumSymmetricSurroundSaliency(source LAB, width, height int, normalize bool) []float64 {
    size := width * height
    saliencyMap := make([]float64, size)
    
    kernel := []float64{1.0, 2.0, 1.0}
    
    ls := gaussianSmooth(source.L, width, height, kernel)
    as := gaussianSmooth(source.A, width, height, kernel)
    bs := gaussianSmooth(source.B, width, height, kernel)

    lint := createIntegralImage(source.L, width, height)
    aint := createIntegralImage(source.A, width, height)
    bint := createIntegralImage(source.B, width, height)

    index := 0
    for j := 0; j < height; j++ {
        yoff := min(j, height - j)
        y1 := max(j - yoff, 0)
        y2 := min(j + yoff, height - 1)

        for k := 0; k < width; k++ {
            xoff := min(k, width - k)
            x1 := max(k - xoff, 0)
            x2 := min(k + xoff, width - 1)

            area := (x2 - x1 + 1) * (y2 - y1 + 1);
            
            lval := getIntegralSum(lint, x1, y1, x2, y2) / float64(area)
            aval := getIntegralSum(aint, x1, y1, x2, y2) / float64(area)
            bval := getIntegralSum(bint, x1, y1, x2, y2) / float64(area)
            
            saliencyMap[index] = (lval - ls[index]) * (lval - ls[index]) + (aval - as[index]) * (aval - as[index]) + (bval - bs[index]) * (bval - bs[index]); //square of the euclidean distance
            index++
        }
    }
    
    if normalize == true {
        doNormalize(saliencyMap, width, height)
    }
    
    return saliencyMap
}


func loadImage(name string) image.Image {
    infile, err := os.Open(name)
    check(err)
    defer infile.Close()
    postfix := strings.ToLower(postfix(name))

    if strings.Compare(postfix, "png") == 0 {
        src, err := png.Decode(infile)
        check(err)
        return src
    } else if strings.Compare(postfix, "jpg") == 0 {
        src, err := jpeg.Decode(infile)
        check(err)
        return src
    }
    panic("Unsupported image type. Must be PNG or JPEG")
}

func postfix(name string) string {
    parts := strings.Split(name, ".")
    if len(parts) != 0 {
        return parts[len(parts) - 1]
    }
    return ""
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {
    if len(os.Args) == 1 {
        usage()
        return
    }

    srcImgName := flag.String("i", "", "Name of the input image")
    saliencyImgName := flag.String("o", "", "Name of the saliency image")
    help := flag.Bool("h", false, "Prints this help")
    flag.Parse()

    if *help {
        usage()
        return
    }

    src := loadImage(*srcImgName)
    lab := imageToLab(src)
    bounds := src.Bounds()
    salMap := computeMaximumSymmetricSurroundSaliency(*lab, bounds.Max.X, bounds.Max.Y, true)
//    fmt.Printf("SalMap: %d - name: %v\n", len(salMap), targetImgName)
    
    gray := image.NewGray(bounds)
    size := len(salMap)
    var Ssum float64
    for i := 0; i < size; i++ {
        Ssum += salMap[i]
        gray.Pix[i] = uint8(salMap[i] * 1.1)
    }
    
    Smean := (1.0 / float64((float64(bounds.Max.X * bounds.Max.Y))) * Ssum)
    fmt.Printf("%f", Smean)
    
    if strings.Compare(*saliencyImgName, "") != 0 {
        saveFile, err := os.Create(*saliencyImgName)
        check(err)
        defer saveFile.Close()
        err = png.Encode(saveFile, gray)
        check(err)
    }
}

func header() {
    fmt.Fprintf(os.Stderr, "Saliency Detection using Maximum Symmetric Surround of images.\n")
    fmt.Fprintf(os.Stderr, "Copyright (c) 2017 Sascha Kohlmann.\n")
}

func usage() {
    fmt.Fprintf(os.Stderr, "usage: si [options] image\n\n")
    header()
    fmt.Fprintf(os.Stderr, "\nOptions:\n")
    fmt.Fprintf(os.Stderr, "  -h       : prints this help\n")
    fmt.Fprintf(os.Stderr, "  -i name  : name of the input image\n")
    fmt.Fprintf(os.Stderr, "  -o name  : stores a saliency control image with <name> - optional\n")
}

