package justeat

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strings"

	// Image detection
	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func (j *JEClient) DownloadLogo(url, filename string) {
	_, err := os.Stat(fmt.Sprintf("logos/%s.jpg", filename))
	if err == nil {
		return
	} else if os.IsNotExist(err) {
	} else {
		log.Println("really bad thing happened on disk")
	}

	err = os.MkdirAll("logos", 0777)
	if err != nil {
		// TODO: Proper logger
		log.Println("failed to mkdir")
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Println("failed to get image")
		return
	}

	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Println("failed to decode gif")
		return
	}

	newImage := image.NewRGBA(image.Rect(0, 0, 160, 160))
	draw.BiLinear.Scale(newImage, newImage.Bounds(), img, img.Bounds(), draw.Over, nil)

	var out bytes.Buffer
	err = jpeg.Encode(bufio.NewWriter(&out), newImage, nil)
	if err != nil {
		log.Println("failed to encode image")
		return
	}

	err = os.WriteFile(fmt.Sprintf("logos/%s.jpg", filename), out.Bytes(), 0666)
	if err != nil {
		log.Println("failed to save image")
		return
	}
}

func (j *JEClient) DownloadFoodImage(path string, restaurantID string, itemID string) {
	path = strings.Replace(path, "{transformations}", "h_160,w_160", -1)

	_, err := os.Stat(fmt.Sprintf("logos/%s/%s.jpg", restaurantID, itemID))
	if err == nil {
		return
	} else if os.IsNotExist(err) {
	} else {
		log.Println("really bad thing happened on disk")
	}

	err = os.MkdirAll(fmt.Sprintf("logos/%s", restaurantID), 0777)
	if err != nil {
		// TODO: Proper logger
		log.Println("failed to mkdir")
		return
	}

	resp, err := http.Get(path)
	if err != nil {
		log.Println("failed to get image")
		return
	}

	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Println("failed to decode gif")
		return
	}

	newImage := image.NewRGBA(image.Rect(0, 0, 160, 160))
	draw.BiLinear.Scale(newImage, newImage.Bounds(), img, img.Bounds(), draw.Over, nil)

	var out bytes.Buffer
	err = jpeg.Encode(bufio.NewWriter(&out), newImage, nil)
	if err != nil {
		log.Println("failed to encode image")
		return
	}

	err = os.WriteFile(fmt.Sprintf("logos/%s/%s.jpg", restaurantID, demae.CompressUUID(itemID)), out.Bytes(), 0666)
	if err != nil {
		log.Println("failed to save image")
		return
	}
}
