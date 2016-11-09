package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/ahmdrz/goinsta"
	"log"
	"os"
	"path/filepath"
	"net/http"
	"io"
	"github.com/google/uuid"
	"strconv"
)

type Cloner struct{
	username string
	password string
	tmpDir string
	imageSource string
	client *goinsta.Instagram
}

func (cloner *Cloner)cloneRandomPhoto(ch chan string){
	filePath, err := cloner.downloadRandomPhoto()
	if err != nil{
		ch <- err.Error()
		return
	}
	resp, err := cloner.client.UploadPhoto(filePath, "https://nemeanthemes.com", cloner.client.NewUploadID(), 100, goinsta.Filter_Lark) // default quality is 87
	err = os.Remove(filePath)
	if err != nil{
		ch <- err.Error()
		return
	}
	ch <-resp.Status
	return
}

func (cloner *Cloner)downloadRandomPhoto() (string, error) {
	absDir, err := filepath.Abs(cloner.tmpDir)
	if err != nil {
		return "",err
	}

	if err := os.MkdirAll(absDir, 0700); err != nil {
		return "",err
	}
	filename, err:= uuid.NewUUID()
	if err != nil {
		return "", err
	}
	fileFullPath := filepath.Join(absDir, filename.String() + ".jpg")
	f, err := os.Create(fileFullPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	resp, err := http.Get(cloner.imageSource)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return fileFullPath, nil
}


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	batchSizeString := os.Getenv("BATCH_SIZE")
	batchSize, err := strconv.Atoi(batchSizeString)
	if err != nil {
		batchSize = 1
	}
	cloner := &Cloner{
		username : os.Getenv("USERNAME"),
		password : os.Getenv("PASSWORD"),
		tmpDir:os.Getenv("TMP_DIR"),
		imageSource:os.Getenv("IMAGE_SOURCE"),
		client : goinsta.New(os.Getenv("USERNAME"), os.Getenv("PASSWORD")),
	}

	if err := cloner.client.Login(); err != nil {
		panic(err)
	}
	defer cloner.client.Logout()

	result := make(chan string)

	for i := 0; i < batchSize; i++ {
		go cloner.cloneRandomPhoto(result)
	}

	for i := 0; i < batchSize; i++ {
		fmt.Println("#" + strconv.Itoa(i) + ":", <-result)
	}
}
