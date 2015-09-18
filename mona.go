package main

import (
	// "encoding/base64"
	"fmt"
	"image"
    "image/draw"
    "image/color"
	"log"
	"os"
	// "strings"

	// Package image/jpeg is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand JPEG formatted images. Uncomment these
	// two lines to also understand GIF and PNG images:
	// _ "image/gif"
	// _ "image/png"
	_ "image/jpeg"
    "image/png"
    "math/rand"
    "time"
    "sort"
    "runtime"
)


type Point struct {
	x int
	y int
}

type Gene struct {
	r uint32
	g uint32
	b uint32
	a uint32
	points []Point
}

type Individual struct {
	img image.Image
	genes []Gene
	fitness int
}

type ByFitness []Individual
func (a ByFitness) Len() int { return len(a) }
func (a ByFitness) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFitness) Less(i, j int) bool { return a[i].fitness > a[j].fitness }


func (gene *Gene) randomGene(maxX, maxY int) {
    gene.r = rand.Uint32()
    gene.g = rand.Uint32()
    gene.b = rand.Uint32()
    // gene.a = rand.Uint32()
    gene.a = 65535
	gene.a = rand.Uint32()
	gene.points = make([]Point,0)
	gene.points = append(gene.points,Point{x:rand.Intn(maxX), y:rand.Intn(maxY)})
}

func (individual *Individual) randomIndividual(genes int, maxX, maxY int) {
    individual.genes = make([]Gene,0);
    for i := 0; i < genes; i++ {
        gene := Gene{}
        gene.randomGene(maxX, maxY)
        individual.genes = append(individual.genes, gene)
    }
}


func (individual *Individual) calculateFitness(m image.Image) {
    bounds := m.Bounds()

    individual.generateImage(bounds.Max.X,bounds.Max.Y)

    maxDiff := bounds.Max.X * bounds.Max.Y * (255 + 255 + 255)
    // fmt.Printf("max diff: %v\n",maxDiff)
    totalDiff := 0

    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			oR, oG, oB, _ := m.At(x, y).RGBA()
            nR, nG, nB, _ := individual.img.At(x, y).RGBA()

            oR = oR/257
            nR = nR/257
            oG = oG/257
            nG = nG/257
            oB = oB/257
            nB = nB/257

            rDiff := int(oR) - int(nR)
            if (rDiff < 0) { rDiff = rDiff * -1; }
            gDiff := int(oG) - int(nG)
            if (gDiff < 0) { gDiff = gDiff * -1; }
            bDiff := int(oB) - int(nB)
            if (bDiff < 0) { bDiff = bDiff * -1; }

            totalDiff += (rDiff + gDiff + bDiff)

			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 12 reduces this to the range [0, 15].
			// fmt.Printf("%v %v %v\n",r,g,b)
		}
	}

	// fmt.Printf("total diff: %v\n",totalDiff)
    individual.fitness = int((float32(maxDiff - totalDiff) / float32(maxDiff)) * 10000)

    individual.deleteImage()
}

func (individual *Individual) deleteImage() {
    individual.img = nil
}

func (individual *Individual) generateImage(x,y int) {

    m := image.NewRGBA(image.Rect(0, 0, x, y))
    bg := color.RGBA{0, 0, 0, 255}
    //bg := color.RGBA{255, 255, 255, 255}

    draw.Draw(m, m.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

    for i := range(individual.genes) {

        r := uint8(individual.genes[i].r/257)
        g := uint8(individual.genes[i].g/257)
        b := uint8(individual.genes[i].b/257)
        a := uint8(individual.genes[i].a/257)

        c := color.RGBA{r,g,b,a}

        // fmt.Printf("%v\n",len(individual.genes[i].points))

        // presume 1 point for now
        rect := image.Rectangle{image.Point{individual.genes[i].points[0].x,individual.genes[i].points[0].y},image.Point{individual.genes[i].points[0].x + 5,individual.genes[i].points[0].y + 5}}

        // draw.Draw(m, rect, &image.Uniform{c}, image.ZP,draw.Src)
		draw.Draw(m, rect, &image.Uniform{c}, image.ZP,draw.Over)

        // do drawmask instead?


        //for j := range(individual.genes[i].points) {
        //    fmt.Printf("%v,%v\n",individual.genes[i].points[j].x,individual.genes[i].points[j].y)
        //}
    }

    individual.img = m

}


func breed(population []Individual, maxX, maxY int) {
    newPopulation := make([]Individual, 0)
    for i := 0; i < len(population)/2; i++ {
        newPopulation = append(newPopulation, makeChild(population[i], population[rand.Intn(len(population)/2-1)],maxX,maxY))
        newPopulation = append(newPopulation, makeChild(population[i], population[rand.Intn(len(population)/2-1)],maxX,maxY))
    }
    // swap new and old pop
    for i:= range newPopulation {
        population[i] = newPopulation[i]
    }
}

func makeChild(mum, dad Individual, maxX, maxY int) Individual{
    child := Individual{}
    child.genes = make([]Gene, len(mum.genes))

    for i := range(child.genes) {
        c := rand.Intn(99)
        if c == 50 {
            child.genes[i].randomGene(maxX,maxY)
        } else if c > 49 {
            child.genes[i] = mum.genes[i]
        } else {
            child.genes[i] = dad.genes[i]
        }
    }
    return child
}


func main() {
    runtime.GOMAXPROCS(4)

    populationSize := 1000
    genes := 3000
    generations := 1000

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




    fmt.Printf("generating random population\n")

    population := make([]Individual, populationSize)
    for i := range(population) {
        population[i] = Individual{}
        population[i].randomIndividual(genes, bounds.Max.X, bounds.Max.Y)
    }


    jobs := make(chan int, len(population))
    results := make(chan int, len(population))

    // start workers
    for w := 1; w <= 2; w++ {
        go worker(w, jobs, results, population, m)
    }



    for j := 0; j < generations; j++ {

        fmt.Printf("generation %v\n",j)

        if (j > 0) {
            // breed
            breed(population, bounds.Max.X, bounds.Max.Y)
        }

        fmt.Printf("calculating fitness\n")



        for i := range(population) {
            jobs <- i
            // population[i].calculateFitness(m)
        }



        // wait for workers, should probably use wg.wait?
        for _ = range population {
            _ = <- results
        }


        sort.Sort(ByFitness(population))

        fmt.Printf("best fitness: %v\n",population[0].fitness)
        population[0].generateImage(bounds.Max.X, bounds.Max.Y)

        toimg, _ := os.Create("best.png")
        png.Encode(toimg, population[0].img)
        toimg.Close()

    }

}


func worker(id int, jobs <-chan int, results chan<- int, population []Individual, img image.Image) {
    for {
        job := <- jobs
        population[job].calculateFitness(img)
        results <- 1
    }
}
