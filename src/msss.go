package main

import (
//    "flag"
    "fmt"
    "image"
//    "image/color"
    "image/jpeg" // register the PNG format with the image package
    "image/png" // register the PNG format with the image package
    "os"
    "strings"
    "math"
)

func main() {
    imageToLab(nil)
}


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
    
    fmt.Printf("lab: %v", lab)
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
            epsilon := 0.008856 //actual CIE standard
            kappa := 903.3      //actual CIE standard

            Xr := 0.950456 //reference white
            Yr := 1.0      //reference white
            Zr := 1.088754 //reference white

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
