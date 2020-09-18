package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

func handler(w http.ResponseWriter, r *http.Request) {
	pictureUrl := "http://192.168.1.10/bildvisare/bilderFotoram/"
	root := "/media/DVDebian/StorageLarge/bilderFotoram/"

	var files []string
	var imgFile *os.File
	var metaData *exif.Exif
	var jsonByte []byte

	type Photo struct {
		DateTime           string `json:"DateTime"`
		DateTimeDigitized  string `json:"DateTimeDigitized"`
		DateTimeOriginal   string `json:"DateTimeOriginal"`
		Orientation        int    `json:"Orientation"`
		PixelXDimension    int    `json:"PixelXDimension"`
		PixelYDimension    int    `json:"PixelYDimension"`
		Make               string `json:"Make"`
		Model              string `json:"Model"`
		Filename           string
		Onlyfilename       string
		Pictureurl         string
		Completepictureurl string
		Exif               bool
		Error              bool
	}
	var photo Photo

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	noOFiles := len(files)
	fmt.Print("Found ", noOFiles, " files.\n")
	if !(noOFiles > 0) {
		log.Fatal("No files found!")
		photo.Error = false
	}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(noOFiles)
	fileName := files[n]
	fmt.Print("File ", n, " - ", fileName, "\n")

	imgFile, err = os.Open(fileName)
	if err != nil {
		photo.Error = true
		log.Fatal(err)
	}

	// Check filetype
	buff := make([]byte, 512) // why 512 bytes ? see http://golang.org/pkg/net/http/#DetectContentType
	_, err = imgFile.Read(buff)

	filetype := http.DetectContentType(buff)
	fmt.Println(filetype)

	metaData, err = exif.Decode(imgFile)
	if err != nil {
		fmt.Println("No exif value found.")
		photo.Exif = false
	} else {
		jsonByte, err = metaData.MarshalJSON()
		if err != nil {
			log.Fatal(err.Error())
		}
		json.Unmarshal([]byte(jsonByte), &photo)
		photo.Exif = true
	}
	photo.Filename = fileName

	// fileName is the whole path. Get just the filename.
	fn := strings.SplitAfter(fileName, "/")
	photo.Onlyfilename = fn[len(fn)-1]
	photo.Pictureurl = pictureUrl
	photo.Completepictureurl = pictureUrl + photo.Onlyfilename

	var jsonData []byte
	jsonData, err = json.MarshalIndent(photo, "", "    ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(jsonData))
}

func main() {
	//var err error
	//for _, file := range files {
	//fmt.Println(file)
	//}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
