package ImageService

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

func startProcessor(workToDo chan string, finishedWorkMap *map[string]bool) {
	var workId string
	counter := 0
	finishedWorkCommunicator := make(chan string)
	for i := 0; i < 40; i++ {
		go startImageProcessorWorker(workToDo, finishedWorkCommunicator)
	}
	for {
		select {
		case workId = <-finishedWorkCommunicator:
			globalFinishedWorkMapMutex.Lock()
			(*finishedWorkMap)[workId] = true
			globalFinishedWorkMapMutex.Unlock()
			counter++
			fmt.Println("Finished:", counter)
		}
	}
}

func startImageProcessorWorker(workToDo chan string, finishedWorkCommunicator chan string) {
	for {
		workId := <-workToDo
		if modifyImage(workId) {
			finishedWorkCommunicator <- workId
		}
	}
}

func modifyImage(workId string) bool {
	file, err := os.Open("/tmp/" + workId + ".png")
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return false
	}
	myImage, err := png.Decode(file)
	if err != nil {
		fmt.Println(err)
		return false
	}
	file.Close()
	file, err = os.Create("/tmp/" + workId + ".png")
	if err != nil {
		fmt.Println(err)
		return false
	}
	m := image.NewRGBA(myImage.Bounds())
	for i := 0; i < m.Rect.Max.X; i++ {
		for j := 0; j < m.Rect.Max.Y; j++ {
			r, g, b, _ := myImage.At(i, j).RGBA()
			myColor := new(color.RGBA)
			myColor.R = uint8((g * g) / 255)
			myColor.G = uint8((r * r) / 255)
			myColor.B = uint8((b * b) / 255)
			myColor.A = uint8(255)
			m.Set(i, j, myColor)
		}
	}
	png.Encode(file, m)
	file.Close()
	return true
}
