package lab

import (
    "image"
    "math"
)

type LAB struct {
    L []float64
    A []float64
    B []float64
    stride int
}

func (l *LAB) Stride() int {
    return l.stride
}

func (l *LAB) initLabSize(size int) {
    if size < 0 {
        panic("Unable to init LAB with size < 0")
    }
    l.L = make([]float64, size)
    l.A = make([]float64, size)
    l.B = make([]float64, size)
}

func ImageToLab(src image.Image) *LAB {
    lab := new(LAB)
    
    bounds := src.Bounds()
    lab.stride = bounds.Max.X
    pixel := bounds.Max.X * bounds.Max.Y
    lab.initLabSize(pixel)

    sem := make(chan int,  bounds.Max.Y)

    for y := 0; y < bounds.Max.Y; y++ {
        go func(y int) {
            for x := 0; x < bounds.Max.X; x++ {
                
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
            sem <- 1
        } (y)
    }
    for y := 0; y < bounds.Max.Y; y++ {<- sem}
    return lab
}
