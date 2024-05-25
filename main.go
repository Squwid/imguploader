package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

var (
	watchPath  = os.Getenv("IMGUPLOADER_WATCH_PATH")
	bucketName = os.Getenv("IMGUPLOADER_BUCKET")
	url        = os.Getenv("IMGUPLOADER_URL")
)

func init() {
	wp, err := homedir.Expand(watchPath)
	if err != nil {
		logrus.WithError(err).Fatalf("Error expanding homepath")
	}
	watchPath = wp
}

func main() {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		logrus.WithError(err).Fatalf("Error creating storage client")
	}
	defer client.Close()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.WithError(err).Fatalf("Error creating new watcher")
	}
	defer watcher.Close()

	// Check for files that were not uploaded yet.
	fs, err := os.ReadDir(watchPath)
	if err != nil {
		logrus.WithError(err).Fatalf("Error on initial read")
	}
	for _, f := range fs {
		if !f.IsDir() {
			if err := maybeUploadFile(client,
				filepath.Join(watchPath, f.Name())); err != nil {
				logrus.WithError(err).Error("Error uploading")
			}
		}
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					if err := maybeUploadFile(client, event.Name); err != nil {
						logrus.WithError(err).Error("Error uploading")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.WithError(err).Errorf("Watcher error")
			}
		}
	}()

	if err := watcher.Add(watchPath); err != nil {
		logrus.WithError(err).Fatalf("Error adding watcher")
	}
	logrus.WithField("Path", watchPath).Infof("Watching")
	<-done
}

func maybeUploadFile(client *storage.Client, file string) error {
	ext := filepath.Ext(file)
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		return upload(client, file, ext)
	}
	return nil
}

func upload(client *storage.Client, file, ext string) error {
	cloudFile := fmt.Sprintf("%s%s", rstr(10), ext)

	writer := client.Bucket(bucketName).Object(cloudFile).
		NewWriter(context.Background())
	writer.ChunkSize = 1024 * 1024

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, writer.ChunkSize)
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if _, err := writer.Write(buf[:n]); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}

	// Remove file once it is dragged in and uploaded
	// TODO: Configure this via CLI or something.
	if err := os.Remove(file); err != nil {
		return err
	}

	fmt.Printf("Uploaded image to %s/%s\n", url, cloudFile)
	return nil
}

// rstr generates a random alpha-numeric string of length length.
func rstr(length int) string {
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	result := make([]rune, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}
