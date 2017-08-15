package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

const (
	outdir string = "more_fermi_go"
	allErr string = "/home/wwc129/fermi2/more_fermi_go/all.log"
	bins   int    = 10
)

var (
	emin      float64        = 100
	emax      float64        = 100000
	wgFlux    sync.WaitGroup = sync.WaitGroup{}
	mutexFlux                = sync.Mutex{}
	errOut    *os.File
)

//make sure the "outdir" is empty before begining
/* func cleanFlux() {
	allfiles, err := filepath.Glob(outdir + "/*")
	if err != nil {
		fmt.Fprintln(errOut, `in "clean" function`, err)
	}
	for _, v := range allfiles {
		if err := os.Remove(v); err != nil {
			fmt.Fprintln(errOut, err)
		}
	}
} */

//create a log file, which contains all the global err information
/* func init() {

	wgFlux.Add(1)
	//first of all, create the outdir (usually is "more_fermi_go"), and create log file called "all.log"
	//"all.log" is aimed to record all the information when doing calculation.
	//the log file is closed in the "main" function
	cleanFlux()
	if err := os.Mkdir(outdir, os.ModePerm); err != nil {
		fmt.Println(err)
	}

	errout, err := os.Create(allErr)
	if err != nil {
		log.Fatal(err)
	}
	errOut = errout
	wgFlux.Done()
	wgFlux.Wait()
} */

func IsExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			// fmt.Println(err)
			// return false
			log.Fatal(err)
		} else {
			return false
		}
	}
	return true
}

//split the emin, emax into different bins
func splitEnergy(lo, hi float64, n int) []float64 {
	res := []float64{}
	Loglo, Loghi := math.Log(lo), math.Log(hi)
	step := (Loghi - Loglo) / float64(n)
	for i := 0; i != n+1; i++ {
		res = append(res, math.Exp(Loglo+float64(i)*step))
	}
	return res
}

//should be "gtselect", but already has "gtselect" in "fermi.go"
func gtselectFlux(ith int) {
	lo, hi := splitEnergy(emin, emax, bins)[ith], splitEnergy(emin, emax, bins)[ith+1]
	fmt.Println("running gtselect")
	loString, hiString := strconv.FormatFloat(lo, 'f', 10, 64), strconv.FormatFloat(hi, 'f', 10, 64)
	ithString := strconv.Itoa(ith)

	cmd := exec.Command("gtselect", "infile=./Prepare/gti.fits", "outfile="+outdir+"/gti_"+ithString+".fits",
		"ra=294.915", "dec=21.6227", "rad=20", "tmin=0", "tmax=600000000", "emin="+loString, "emax="+hiString, "zmax=90")
	if err := cmd.Run(); err != nil {
		fmt.Println("wrong in gtselectFlux")
		log.Fatal(err)
	}

	// cmd := exec.Command("python", "gtselect.py", loString, hiString, ithString, outdir)
	if !IsExists(filepath.Join(outdir, "Err")) {
		if err := os.Mkdir(filepath.Join(outdir, "Err"), os.ModePerm); err != nil {
			fmt.Println(err)
		}
	}
	outErr, err := os.Create(filepath.Join(outdir, "Err", "gtselect_"+ithString+".err"))
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	defer outErr.Close()
	cmd.Stderr = outErr
	cmd.Stdout = os.Stdout
	outErr.Write([]byte("Loop " + ithString + ":\n"))
	fmt.Println("finished")
}

func gtmktimeFlux(ith int) {
	fmt.Println("running gtmktime")
	ithString := strconv.Itoa(ith)
	cmd := exec.Command("python", "gtmktime.py", ithString, outdir, "1")
	outErr, err := os.Create(filepath.Join(outdir, "Err", "gtmktime_"+ithString, ".err"))
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	outErr.Write([]byte("Loop " + ithString + ":\n"))
	defer outErr.Close()
	cmd.Stderr = outErr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func gtbinCmap(ith int) {
	fmt.Println("running gtbin_cmap")
	ithString := strconv.Itoa(ith)
	cmd := exec.Command("python", "gtbin_cmap.py", ithString, outdir, "1")
	outErr, err := os.Create(filepath.Join(outdir, "Err", "cmap_"+ithString+".err"))
	defer outErr.Close()
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	outErr.Write([]byte("Loop " + ithString + ":\n"))
	cmd.Stderr = outErr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func gtbinCcube(ith int) {
	lo, hi := splitEnergy(emin, emax, bins)[ith], splitEnergy(emin, emax, bins)[ith+1]
	fmt.Println("running gtbin_ccube")
	loString, hiString := strconv.FormatFloat(lo, 'f', 10, 64), strconv.FormatFloat(hi, 'f', 10, 64)
	ithString := strconv.Itoa(ith)
	cmd := exec.Command("python", "gtbin_ccube.py", ithString, loString, hiString, outdir, "1")
	outErr, err := os.Create(filepath.Join(outdir, "Err", "ccube_"+ithString+".err"))
	defer outErr.Close()
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	outErr.Write([]byte("Loop " + ithString + ":\n"))
	cmd.Stderr = outErr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func model(ith int) {
	fmt.Println("generating input sources, adding my sources")
	ithString := strconv.Itoa(ith)
	cmd := exec.Command("python", "model.py", ithString, outdir, "1")
	outErr, err := os.Create(filepath.Join(outdir, "Err", "model_"+ithString+".err"))
	defer outErr.Close()
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	outErr.Write([]byte("Loop " + ithString + ":\n"))
	cmd.Stderr = outErr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	//below is for adding my own source
	mutexFlux.Lock()
	file, err := os.OpenFile(filepath.Join(outdir, "input_"+ithString+".xml"), os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Fprintln(errOut, err)
	}

	addMysource, err := ioutil.ReadFile("add_mysource.txt")
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	mySouceData := string(addMysource)

	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	dataString := string(data)
	dataSlice := strings.Split(dataString, "\n")
	beforeLines := strings.Join(dataSlice[:len(dataSlice)-1], "\n")
	file.Seek(0, 0)
	file.Truncate(0)
	file.WriteString(beforeLines + "\n")
	file.WriteString(mySouceData + "\n")
	mutexFlux.Unlock()
}

func gtltcube(ith int) {
	ithString := strconv.Itoa(ith)
	fmt.Println("running gtltcube")
	cmd := exec.Command("python", "gtltcube.py", ithString, outdir, "1")
	outErr, err := os.Create(filepath.Join(outdir, "Err", "gtltcube_"+ithString+".err"))
	defer outErr.Close()
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	outErr.Write([]byte("Loop " + ithString + ":\n"))
	cmd.Stderr = outErr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

}

func gtexpcube2(ith int) {
	lo, hi := splitEnergy(emin, emax, bins)[ith], splitEnergy(emin, emax, bins)[ith+1]
	fmt.Println("running gtexpcube2")
	loString, hiString := strconv.FormatFloat(lo, 'f', 10, 64), strconv.FormatFloat(hi, 'f', 10, 64)
	ithString := strconv.Itoa(ith)
	cmd := exec.Command("python", "gtexpcube2.py", ithString, loString, hiString, outdir, "1")
	outErr, err := os.Create(filepath.Join(outdir, "Err", "gtexpcube2_"+ithString+".err"))
	defer outErr.Close()
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	outErr.Write([]byte("Loop " + ithString + ":\n"))
	cmd.Stderr = outErr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func gtsrcmaps(ith int) {
	ithString := strconv.Itoa(ith)
	fmt.Println("running gtsrcmaps")
	cmd := exec.Command("python", "gtsrcmaps.py", ithString, outdir, "1")
	outErr, err := os.Create(filepath.Join(outdir, "Err", "gtsrcmaps_"+ithString+".err"))
	defer outErr.Close()
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	outErr.Write([]byte("Loop " + ithString + ":\n"))
	cmd.Stderr = outErr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func gtlike(ith int) {
	ithString := strconv.Itoa(ith)
	fmt.Println("running gtsrcmaps")
	cmd := exec.Command("python", "gtlike.py", ithString, outdir, "1")
	outErr, err := os.Create(filepath.Join(outdir, "Err", "gtlike"+ithString+".err"))
	defer outErr.Close()
	if err != nil {
		fmt.Fprintln(errOut, err)
	}
	outErr.Write([]byte("Loop " + ithString + ":\n"))
	cmd.Stderr = outErr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func process(ith int) {
	//gtselectFlux(ith)
	//fermiOuter(ith)

	//gtmktimeFlux(ith)
	/* gtbinCmap(ith)
	gtbinCcube(ith)
	model(ith)
	gtltcube(ith) */
	gtexpcube2(ith)
	gtsrcmaps(ith)
	gtlike(ith)
	wgFlux.Done()
}

func main() {
	debug.SetGCPercent(800)
	//gtselectFlux should be running seperately, or different goroutines will read the same gti file.
	/* for i := 0; i != bins; i++ {
		gtselectFlux(i)
	} */
	wgFlux.Add(bins)
	for i := 0; i != bins; i++ {
		go process(i)
	}
	wgFlux.Wait()
	defer errOut.Close()

}
