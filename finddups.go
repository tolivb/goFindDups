package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type fresult struct {
	cnt  int
	path string
	md5  string
	size int64
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		log.Fatalf("Please provide at least 1 dir to scan for duplicate files")
	}

	//check if the provided dirs are OK
	dirs := make(map[string]*os.File)
	for i := 0; i < len(args); i++ {
		dirFD, err := os.Open(args[i])
		if err != nil {
			log.Fatalf("Unable to open dir: " + args[i] + "\n" + err.Error())
		}

		defer dirFD.Close()

		dirStats, err := dirFD.Stat()
		if err != nil {
			log.Fatalf("Unable to stat " + args[i])
		}

		if !dirStats.IsDir() {
			log.Fatalf(args[i] + " is not a directory")
		}
		dirs[args[i]] = dirFD
	}

	//sync channels
	var wg sync.WaitGroup

	files := make(chan *fresult, 5)
	results := make(chan *fresult, 5)

	//start md5 workers
	cpus := runtime.NumCPU() - 1
	fmt.Printf("Starting with %d workers\n\n", cpus)
	for i := 0; i < cpus; i++ {
		wg.Add(1)
		go func(files chan *fresult, results chan *fresult) {
			defer wg.Done()
			for f := range files {

				if f.size == 0 {
					f.md5 = "empty"
				} else {
					md5x, err := md5sum(f.path)
					if err != nil {
						fmt.Printf("ERROR: %v\n\n", err)
						continue
					}
					f.md5 = md5x
				}
				results <- f
			}
		}(files, results)
	}

	//start the walker
	go func(dirs map[string]*os.File, files chan *fresult) {
		defer close(files)
		for k, _ := range dirs {
			err := filepath.Walk(k, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					fmt.Printf("ERROR: %v\n\n", err)
					return nil
				}

				if info.IsDir() {
					return nil
				}
				files <- &fresult{0, path, "", info.Size()}
				return nil
			})

			if err != nil {
				fmt.Printf("error walking the path %q: %v\n", k, err)
				return
			}
		}
	}(dirs, files)

	//Finito
	go func() {
		wg.Wait()
		close(results)
	}()

	filesList := map[string]*fresult{}
	var wastedSpace int64
	rand.Seed(time.Now().UTC().UnixNano())
	tmpDirName := "finddups_" + strconv.FormatInt(rand.Int63(), 10)

	for res := range results {
		v, ok := filesList[res.md5]

		if ok {
			filesList[res.md5].cnt += 1
			fmt.Println("DUPS:", v.path, res.path)
			fmt.Println("SH:mkdir -p '" + tmpDirName + "/" + path.Dir(res.path) + "'")
			fmt.Println("SH:mv '" + res.path + "' '" + tmpDirName + "/" + res.path + "'\n")
			wastedSpace += v.size
		} else {
			filesList[res.md5] = &fresult{1, res.path, res.md5, res.size}
		}
	}

	fmt.Printf("SUMUP: Wasted space: %.3f MB\n", float64(wastedSpace)/1024/1024)
}

func md5sum(file string) (string, error) {
	fd, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer fd.Close()
	h := md5.New()

	if _, err := io.Copy(h, fd); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
