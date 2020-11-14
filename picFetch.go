package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/cors"
	"github.com/xor-gate/goexif2/exif"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Paths struct {
		PictureUrl string `yaml:"pictureUrl"`
		Root       string `yaml:"root"`
	} `yaml:"paths"`
}

var rowCnt = 0
var root string
var pictureUrl string

func handler(w http.ResponseWriter, r *http.Request) {
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
		ImageWidth         int    `json:"ImageWidth"`
		ImageHeight        int    `json:"ImageHeight"`
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

	fmt.Println("Root: " + root)

	var err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
	fileType := http.DetectContentType(buff)
	fmt.Println("filetype: ", fileType) // image/jpeg, video/mp4,

	metaData, err = exif.Decode(imgFile)
	fmt.Print("Metadata:")
	fmt.Println(metaData)

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

	writeToGD(fileName, fileType, photo.Exif)

	var jsonData []byte
	jsonData, err = json.MarshalIndent(photo, "", "    ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println("jsonData: " + string(jsonData))

	fmt.Fprintf(w, string(jsonData)) // Print on web page

	fmt.Println("Done")

}

//func (config Config) Run() {
func Run() {
	/*
		fmt.Println("Settings:")
		root := config.Paths.Root
		fmt.Println("Root: " + root)
		pictureUrl := config.Paths.PictureUrl
		fmt.Println("pictureUrl: " + pictureUrl)
	*/
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", doNothing)
	mux.HandleFunc("/", handler)

	//pictureUrl := "http://192.168.1.10/bildvisare/bilderFotoram/"
	//root := "/media/DVDebian/StorageLarge/bilderFotoram/"

	cfgPath, err := ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Settings:")
	root = cfg.Paths.Root
	fmt.Println("Root: " + root)
	pictureUrl = cfg.Paths.PictureUrl
	fmt.Println("pictureUrl: " + pictureUrl)

	newHandler := cors.Default().Handler(mux)
	log.Fatal(http.ListenAndServe(":8080", newHandler))

}

func main() {
	fmt.Println("Start")
	/*
		cfgPath, err := ParseFlags()
		if err != nil {
			log.Fatal(err)
		}
		cfg, err := NewConfig(cfgPath)
		if err != nil {
			log.Fatal(err)
		}
	*/
	//	fmt.Println(cfg)

	//cfg.Run()
	Run()
}

func doNothing(w http.ResponseWriter, r *http.Request) {}

func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	home, _ := os.UserHomeDir()
	configfilepath := home + "/.config/config.yml"
	fmt.Println("Open " + configfilepath)

	if _, err := os.Stat(configfilepath); os.IsNotExist(err) {
		fmt.Println("Config file config.yml in $HOME/.config/ is missing.")
		os.Exit(1)
	}

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", configfilepath, "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}

func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
