package main

import (
	"fmt"
)

func settings() {
	pictureUrl := "http://192.168.1.10/bildvisare/bilderFotoram/"
	root := "/media/DVDebian/StorageLarge/bilderFotoram/"
	fmt.Print("Found settings: " + pictureUrl + ", " + root)
}
