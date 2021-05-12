package main

import (
	"ImageResize/settings"
	"bytes"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/panjf2000/ants/v2"
	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
	"image"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func isFile(file string) bool {
	if s, err := os.Stat(file); s.IsDir() || err != nil {
		return false
	}

	return true
}
func isDirectory(file string) bool {
	if s, err := os.Stat(file); s.IsDir() || err != nil {
		return true
	}

	return false
}

// listFilesInDir returns an array of files found in a directory
func listFilesInDir(path string) ([]string, error) {
	var fileList []string
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fileList, err
	}

	for _, f := range files {
		fileList = append(fileList, filepath.Join(path, f.Name()))
	}
	return fileList, nil
}

func fileNameWithoutExtension(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}

// returns image and imageformat
func openFile(file string) (image.Image, string) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Println("Error opening file!")
	}
	defer f.Close()

	img, fmtName, err := image.Decode(f)
	if err != nil {
		fmt.Println("error opening file " + file)
		os.Exit(1)
	}
	return img, fmtName
}

func findJpgSizeCompression(logFileName string, img image.Image, targetByteSize int) bytes.Buffer {
	var buf bytes.Buffer

	for iQuality := 100; iQuality > 1; iQuality-- {
		buf.Reset()

		opt := jpeg.Options{
			Quality: iQuality,
		}

		err := jpeg.Encode(&buf, img, &opt)
		if err != nil {
			fmt.Println("Error encoding file!")
		}

		imageSize := buf.Len()
		if imageSize <= targetByteSize {
			fmt.Println("compression: " + strconv.Itoa(iQuality) + "% for " + logFileName)
			break
		}
	}

	return buf
}

func compressFile(outputFile string, img image.Image, targetByteSize int) {
	imageBuffer := findJpgSizeCompression(outputFile, img, targetByteSize)

	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating file!")
	}
	defer f.Close()

	_, err = f.Write(imageBuffer.Bytes())
	if err != nil {
		fmt.Println("Error writing file!")
		os.Exit(1)
	}
}

var Workers = 5
var ShowSpinner = false

type CompressWorker struct {
	File string
	Bar *mpb.Bar
}

func main() {
	var Files []string

	// load config
	var config = settings.Config
	config.GetConf()

	Workers = config.Workers

	var wantedFileSize datasize.ByteSize
	err := wantedFileSize.UnmarshalText([]byte(config.TargetFileSize))
	if err != nil {
		fmt.Println("Invalid filesize format. (for example: '10 MB', '28 kilobytes')")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	var spinner *mpb.Progress
	if ShowSpinner {
		spinner = mpb.New(mpb.WithWaitGroup(&wg), mpb.WithOutput(os.Stderr))
	}

	pool, _ := ants.NewPoolWithFunc(Workers, func(compressWorker interface{}) {
		start := time.Now()

		file := compressWorker.(CompressWorker).File
		spinnerBar := compressWorker.(CompressWorker).Bar

		// compress file
		fileImage, _ := openFile(file)
		newFilename := strings.ReplaceAll(config.NewFilename, "{filename}", fileNameWithoutExtension(filepath.Base(file)))
		newFilename = strings.ReplaceAll(newFilename, "{ext}", "jpg")

		compressFile(newFilename, fileImage, int(wantedFileSize.Bytes()))

		// add file to worker file list
		Files = append(Files, file)

		if ShowSpinner && spinnerBar != nil {
			spinnerBar.Increment()
			spinnerBar.DecoratorEwmaUpdate(time.Since(start))
			spinnerBar.SetCurrent(100)
		}
		defer wg.Done()
	}, ants.WithExpiryDuration(3e+10))
	defer pool.Release()



	if len(os.Args) < 2 {
		fmt.Println("You must drag and drop a file!")
		os.Exit(1)
	} else {
		fileArray := os.Args[1:]


		if isDirectory(fileArray[0]) {
			fileArray, _ = listFilesInDir(fileArray[0])
		}

		for _, file := range fileArray {
			file = strings.Replace(file, "File ", "", 1)
			if isFile(file) {

				var bar *mpb.Bar
				if ShowSpinner {

					bar = spinner.AddSpinner(int64(100),
						mpb.SpinnerOnLeft,
						mpb.PrependDecorators(
							// simple name decorator
							decor.Name(filepath.Base(file)),
							// decor.DSyncWidth bit enables column width synchronization
							decor.Percentage(decor.WCSyncSpace),
						),
						mpb.AppendDecorators(
							// replace ETA decorator with "done" message, OnComplete event
							decor.OnComplete(
								// ETA decorator with ewma age of 60
								decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
							),
						),
						mpb.BarWidth(5),
						mpb.BarRemoveOnComplete(),
					)
				}
				wg.Add(1)
				_ = pool.Invoke(CompressWorker{File: file, Bar: bar})

			} else {
				fmt.Println("File " + file + " does not exist!")
			}
		}
		wg.Wait()
	}

	fmt.Println("Finished !!!")
}
