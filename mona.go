package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	// _ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math/rand"
	"runtime"
	"sort"
	"time"
)

type Point struct {
	x int
	y int
	w int
	h int
}

type Gene struct {
	r      uint8 // red 0-255
	g      uint8 // green 0-255
	b      uint8 // blue 0-255
	a      uint8 // alpha 0-255
	points []Point
}

type Pixel struct {
	r uint8 // red 0-255
	g uint8 // green 0-255
	b uint8 // blue 0-255
}

type Individual struct {
	pixels  [][]Pixel
	img     image.Image
	genes   []Gene
	fitness int
}

type ByFitness []Individual

func (a ByFitness) Len() int           { return len(a) }
func (a ByFitness) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFitness) Less(i, j int) bool { return a[i].fitness > a[j].fitness }

func (gene *Gene) randomGene(maxX, maxY int) {
	gene.r = uint8(rand.Intn(255))
	gene.g = uint8(rand.Intn(255))
	gene.b = uint8(rand.Intn(255))
	gene.a = uint8(rand.Intn(127)+128)

	gene.points = make([]Point, 0)
	gene.points = append(gene.points, Point{x: (rand.Intn(maxX+20) - 20), y: (rand.Intn(maxY+20) - 20), w: (rand.Intn(15)), h: (rand.Intn(15))})
}

func (individual *Individual) randomIndividual(genes int, maxX, maxY int) {
	individual.genes = make([]Gene, 0)
	for i := 0; i < genes; i++ {
		gene := Gene{}
		gene.randomGene(maxX, maxY)
		individual.genes = append(individual.genes, gene)
	}
}

func (individual *Individual) calculateFitness(m image.Image, original [][]Pixel) {
	bounds := m.Bounds()

	individual.generateImagePixels(bounds.Max.X, bounds.Max.Y)

	maxDiff := uint64(bounds.Max.X * bounds.Max.Y * (255 + 255 + 255))

	totalDiff := uint64(0)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			oR := original[x][y].r
			oG := original[x][y].g
			oB := original[x][y].b

			nR := individual.pixels[x][y].r
			nG := individual.pixels[x][y].g
			nB := individual.pixels[x][y].b

			rDiff := 0
			gDiff := 0
			bDiff := 0

			if oR > (nR) {
				rDiff = int(oR - (nR))
			} else {
				rDiff = int((nR) - oR)
			}
			if oG > (nG) {
				gDiff = int(oG - (nG))
			} else {
				gDiff = int((nG) - oG)
			}
			if oB > (nB) {
				bDiff = int(oB - (nB))
			} else {
				bDiff = int((nB) - oB)
			}

			totalDiff += uint64(rDiff + gDiff + bDiff)

			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 12 reduces this to the range [0, 15].
			// fmt.Printf("%v %v %v\n",r,g,b)
		}
	}

	// fmt.Printf("total diff: %v\n",totalDiff)
	individual.fitness = int((float32(maxDiff-totalDiff) / float32(maxDiff)) * 10000)

	individual.deleteImage()
}

func (individual *Individual) deleteImage() {
	individual.img = nil
}

func (individual *Individual) generateImagePixels(x, y int) {

	// make pixels 2d array
	individual.pixels = make([][]Pixel, x)
	for i := 0; i < x; i++ {
		individual.pixels[i] = make([]Pixel, y)
	}
	// fill with black
	for i := 0; i < y; i++ {
		for j := 0; j < x; j++ {
			individual.pixels[j][i] = Pixel{r: 0, g: 0, b: 0}
		}
	}

	for i := range individual.genes {

		r := uint8(individual.genes[i].r)
		g := uint8(individual.genes[i].g)
		b := uint8(individual.genes[i].b)
		// a := uint8(individual.genes[i].a)

		// rectangle
		x1 := individual.genes[i].points[0].x
		x2 := individual.genes[i].points[0].x + individual.genes[i].points[0].w

		y1 := individual.genes[i].points[0].y
		y2 := individual.genes[i].points[0].y + individual.genes[i].points[0].h


		// cut off any of the shape outside of the image dimensions
		if x1 < 0 {
			x1 = 0
		}
		if y1 < 0 {
			y1 = 0
		}
		if x2 >= x {
			x2 = x - 1
		}
		if y2 >= y {
			y2 = y - 1
		}

		for k := y1; k <= y2; k++ {
			for j := x1; j <= x2; j++ {
				var currentPixel = individual.pixels[j][k];

				// work out some dodgy version of alpha mix between current and new pixel
				// return in int in case they overflowed 255
				newr := (int(currentPixel.r) + int(r) ) / 2;
				newg := (int(currentPixel.g) + int(g) ) / 2;
				newb := (int(currentPixel.b) + int(b) ) / 2;

				// make sure they wont overflow uint8 (0-255)
				if newr > 255 { newr = 255; }
				if newg > 255 { newg = 255; }
				if newb > 255 { newb = 255; }

				individual.pixels[j][k] = Pixel{r: uint8(newr), g: uint8(newg), b: uint8(newb)}
				// need to make use of a (alpha) from above - if used
			}
		}

	}

}

func (individual *Individual) generateActualImage(x, y int) {

	m := image.NewRGBA(image.Rect(0, 0, x, y))
	bg := color.RGBA{0, 0, 0, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	for oy := 0; oy < y; oy++ {
		for ox := 0; ox < x; ox++ {
			r := individual.pixels[ox][oy].r
			g := individual.pixels[ox][oy].g
			b := individual.pixels[ox][oy].b
			a := uint8(255)
			c := color.RGBA{r, g, b, a}
			m.Set(ox, oy, c)
		}

	}

	individual.img = m

}

func breed(population []Individual, maxX, maxY int) {
	newPopulation := make([]Individual, 0)
	for i := 0; i < len(population)/2; i++ {
		newPopulation = append(newPopulation, makeChild(population[i], population[rand.Intn(len(population)/2-1)], maxX, maxY))
		newPopulation = append(newPopulation, makeChild(population[i], population[rand.Intn(len(population)/2-1)], maxX, maxY))
	}
	// swap new and old pop
	for i := range newPopulation {
		population[i] = newPopulation[i]
	}
}

func massiveMutate(population []Individual, maxX, maxY int) {
	for i := len(population) / 4; i < len(population)-1; i++ {
		for j := 0; j < len(population[i].genes); j++ {
			population[i].genes[j].randomGene(maxX, maxY)
		}
	}
}

func makeChild(mum, dad Individual, maxX, maxY int) Individual {
	child := Individual{}
	child.genes = make([]Gene, len(mum.genes))

	for i := range child.genes {
		c := rand.Intn(99)
		if c == 50 {
			child.genes[i].randomGene(maxX, maxY)
		} else if c > 49 {
			child.genes[i] = mum.genes[i]
		} else {
			child.genes[i] = dad.genes[i]
		}
	}
	return child
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	populationSize := 1000
	genes := 2000
	generations := 10000

	rand.Seed(time.Now().Unix())

	reader, err := os.Open("080411_mona-lisa.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	bounds := m.Bounds()

	// get original in terms of pixels
	original := make([][]Pixel, bounds.Max.X)
	for i := 0; i < bounds.Max.X; i++ {
		original[i] = make([]Pixel, bounds.Max.Y)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oR, oG, oB, _ := m.At(x, y).RGBA()
			r := uint8(oR / 257)
			g := uint8(oG / 257)
			b := uint8(oB / 257)
			original[x][y] = Pixel{r: r, g: g, b: b}
		}
	}

	fmt.Printf("generating random population\n")

	population := make([]Individual, populationSize)
	for i := range population {
		population[i] = Individual{}
		population[i].randomIndividual(genes, bounds.Max.X, bounds.Max.Y)
	}

	jobs := make(chan int, len(population))
	results := make(chan int, len(population))

	// start workers
	for w := 1; w <= 2; w++ {
		go worker(w, jobs, results, population, m, original)
	}

	var bestScore = 0

	for j := 0; j < generations; j++ {

		fmt.Printf("generation %v\n", j)

		if j > 0 {
			breed(population, bounds.Max.X, bounds.Max.Y)
		}

		fmt.Printf("calculating fitness\n")

		for i := range population {
			jobs <- i
		}

		// wait for workers, should probably use wg.wait?
		for _ = range population {
			_ = <-results
		}

		sort.Sort(ByFitness(population))

		var totalFitness = 0
		for i := range population {
			totalFitness += population[i].fitness
		}

		var averageFitness = totalFitness / len(population)

		fmt.Printf("best fitness: %v\n", population[0].fitness)
		fmt.Printf("average fitness: %v\n", averageFitness)

		if population[0].fitness > bestScore {
			bestScore = population[0].fitness
			fmt.Printf("New best score!\n")
			population[0].generateActualImage(bounds.Max.X, bounds.Max.Y)
			toimg, _ := os.Create("best.png")
			png.Encode(toimg, population[0].img)
			toimg.Close()
		} else {
			fmt.Printf("Score was no better than the current best\n")
		}


	}

}

func worker(id int, jobs <-chan int, results chan<- int, population []Individual, img image.Image, original [][]Pixel) {
	for {
		job := <-jobs
		population[job].calculateFitness(img, original)
		results <- 1
	}
}
