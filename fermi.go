package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

var (
	tmin        float64        = 241747201
	tmax        float64        = 520473605
	splitNumber int            = 5
	CPUnumber   int            = 10
	wg          sync.WaitGroup = sync.WaitGroup{}
	mutex                      = sync.Mutex{}
)

func split(lo, hi float64) []float64 {
	step := (hi - lo) / float64(splitNumber)
	res := make([]float64, splitNumber+1)
	for i := 0; i != splitNumber+1; i++ {
		res[i] = lo + step*float64(i)
	}
	return res
}

func clean(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	} else if err == nil {
		os.Remove(path)
	}
}

func gtselect(tmin, tmax float64, i, outer int) {
	cmd := exec.Command("gtselect", "evclass=128", "evtype=3", "infile=/home/wwc129/fermi2/more_fermi_go/filtered_"+
		strconv.Itoa(outer)+".fits",
		"outfile=./split/gtselect_"+strconv.Itoa(outer)+"_"+strconv.Itoa(i)+".fits", "ra=294.915", "dec=21.6227", "rad=20",
		"tmin="+strconv.FormatFloat(tmin, 'f', 0, 64),
		"tmax="+strconv.FormatFloat(tmax, 'f', 0, 64), "emin=100", "emax=100000", "zmax=90")

	/* if err := os.MkdirAll("Split_and_Merge", os.ModePerm); err != nil {
		log.Fatal(err)
	} */

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		log.Fatal("gtselect wrong")
	}
	//c <- 1
}

func gtmktime(n, outer int, c chan int) {
	cmd := exec.Command("gtmktime", "scfile=./p1937sc00.fits", "filter=DATA_QUAL>0 && LAT_CONFIG==1", "roicut=no",
		"evfile=./split/gtselect_"+strconv.Itoa(outer)+"_"+strconv.Itoa(n)+".fits",
		"outfile=./split/gti_"+strconv.Itoa(outer)+"_"+strconv.Itoa(n)+".fits")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("gtmktime wrong")
		log.Fatal(err)
	}
	if err := os.Remove("./split/gtselect_" + strconv.Itoa(outer) + "_" + strconv.Itoa(n) + ".fits"); err != nil {
		fmt.Println("cannot delete the gtselect file!!")
		log.Fatal(err)
	}
	c <- 1
}

func tempo2(n, outer int) {
	fmt.Println(n, "Loop")
	cmd := exec.Command("tempo2", "-gr", "fermi", "-graph", "0", "-ft1", "./split/gtselect_"+
		strconv.Itoa(outer)+"_"+strconv.Itoa(n)+".fits",
		"-ft2", "/data/wwc129/b1937/fermi2/tempSpace/p1937sc00_"+strconv.Itoa(outer)+".fits",
		"-f", "/data/wwc129/b1937/fermi2/tempSpace/PSRJ1939+2134_2PC_"+strconv.Itoa(outer)+".par", "-phase")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		log.Fatal("tempo2 wrong, in line 88 !!!", outer, n)
	}
}

func merge(outer int) {
	/* if err := os.Remove("./split/files_" + strconv.Itoa(outer) + ".txt"); err != nil {
		fmt.Println(err)
	} */
	//merge gti files for gtselect to use
	target, err := os.Create("./split/files_" + strconv.Itoa(outer) + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	defer target.Close()
	allFiles, err := filepath.Glob("./split/gti_" + strconv.Itoa(outer) + "_?.fits")
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range allFiles {
		target.WriteString(v + "\n")
	}

	cmd2 := exec.Command("gtselect", "infile=./split/files_"+strconv.Itoa(outer)+".txt",
		"outfile=./split/gti_total_"+strconv.Itoa(outer)+".fits", "ra=294.915", "dec=21.6227",
		"rad=20", "tmin=0", "tmax=600000000", "emin=100", "emax=100000", "zmax=90", "clobber=yes")
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		log.Fatal(err)
	}
}

func selectphase(outer int) {
	outerString := strconv.Itoa(outer)
	cmd1 := exec.Command("gtselect", "infile=./split/gti_total_"+outerString+".fits",
		"outfile="+outdir+"/gti1_"+outerString+".fits",
		"ra=294.915", "dec=21.6227", "rad=20", "tmin=0", "tmax=600000000", "emin=100", "emax=100000", "zmax=90", "clobber=yes",
		"phasemin=0.0", "phasemax=0.2")

	cmd2 := exec.Command("gtselect", "infile=./split/gti_total_"+outerString+".fits",
		"outfile="+outdir+"/gti2_"+outerString+".fits",
		"ra=294.915", "dec=21.6227", "rad=20", "tmin=0", "tmax=600000000", "emin=100", "emax=100000", "zmax=90", "clobber=yes",
		"phasemin=0.5", "phasemax=0.7")

	wgTemp := sync.WaitGroup{}
	wgTemp.Add(2)

	go func() {
		if err := cmd1.Run(); err != nil {
			log.Fatal(err)
		}
		wgTemp.Done()
	}()

	go func() {
		if err := cmd2.Run(); err != nil {
			log.Fatal(err)
		}
		wgTemp.Done()
	}()

	wgTemp.Wait()

}

//this function use "ftmerge", so be careful
func mergeAgain(outer int) {
	outerString := strconv.Itoa(outer)
	cmd := exec.Command("ftmerge", outdir+"/gti1_"+outerString+".fits,"+outdir+"/gti2_"+outerString+".fits",
		outdir+"/gti_"+outerString+".fits")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func fermiOuter(outer int) {
	clean("/home/wwc129/fermi2/split")
	splitNumber = 5
	runtime.GOMAXPROCS(CPUnumber)
	channel := make([]chan int, CPUnumber)

	init := func() {
		for i := 0; i != CPUnumber; i++ {
			channel[i] = make(chan int)
			go func(c chan int) {
				c <- 1
			}(channel[i])
		}
	}
	consume := func() {
		for i := 0; i != CPUnumber; i++ {
			<-channel[i]
		}
	}
	TimeIntervals := split(tmin, tmax)

	init()
	func() {
		for i := 0; i != splitNumber; i++ {
			// <-channel[i%CPUnumber]
			min, max := TimeIntervals[i], TimeIntervals[i+1]
			gtselect(min, max, i, outer)
		}
		consume()
	}()

	// init()
	func() {
		for i := 0; i != splitNumber; i++ {
			tempo2(i, outer)
		}
		// consume()
	}()

	init()
	func() {
		for i := 0; i != splitNumber; i++ {
			<-channel[i%CPUnumber]
			go gtmktime(i, outer, channel[i%CPUnumber])
		}
		consume()
	}()

	merge(outer)
	selectphase(outer)
	mergeAgain(outer)
	fmt.Println("finished fermiOuter")
}
