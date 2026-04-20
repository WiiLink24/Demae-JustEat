package justeat

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/WiiLink24/DemaeJustEat/logger"
	"golang.org/x/image/draw"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	// Image detection
	_ "golang.org/x/image/webp"
)

func (j *JEClient) DownloadLogo(url, filename string) {
	_, err := os.Stat(fmt.Sprintf("logos/%s.jpg", filename))
	if err == nil {
		return
	} else if os.IsNotExist(err) {
	} else {
		logger.Error(Image, err.Error())
		return
	}

	err = os.MkdirAll("logos", 0777)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(Image, err.Error())
		}
	}(resp.Body)
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}

	newImage := image.NewRGBA(image.Rect(0, 0, 160, 160))
	draw.BiLinear.Scale(newImage, newImage.Bounds(), img, img.Bounds(), draw.Over, nil)

	var out bytes.Buffer
	err = jpeg.Encode(bufio.NewWriter(&out), newImage, nil)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}

	err = os.WriteFile(fmt.Sprintf("logos/%s.jpg", filename), out.Bytes(), 0666)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}
}

func (j *JEClient) DownloadFoodImage(path string, restaurantID string, itemID string) {
	path = strings.ReplaceAll(path, "{transformations}", "h_160,w_160")

	_, err := os.Stat(fmt.Sprintf("logos/%s/%s.jpg", restaurantID, itemID))
	if err == nil {
		return
	} else if os.IsNotExist(err) {
	} else {
		logger.Error(Image, err.Error())
		return

	}

	err = os.MkdirAll(fmt.Sprintf("logos/%s", restaurantID), 0777)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}

	resp, err := http.Get(path)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(Image, err.Error())
		}
	}(resp.Body)
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}

	newImage := image.NewRGBA(image.Rect(0, 0, 160, 160))
	draw.BiLinear.Scale(newImage, newImage.Bounds(), img, img.Bounds(), draw.Over, nil)

	var out bytes.Buffer
	err = jpeg.Encode(bufio.NewWriter(&out), newImage, nil)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}

	err = os.WriteFile(fmt.Sprintf("logos/%s/%s.jpg", restaurantID, itemID), out.Bytes(), 0666)
	if err != nil {
		logger.Error(Image, err.Error())
		return
	}
}
