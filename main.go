package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcutil/base58"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gopkg.in/yaml.v2"
)

type Config struct {
	EndPoint       string `yaml:"endPoint"`
	AccessUser     string `yaml:"accessUser"`
	AccessPassword string `yaml:"accessPassword"`
	Bucket         string `yaml:"bucket"`
}

var config Config

func main() {
	// Load config
	ex, err := os.Executable()
	if err != nil {
		log.Fatalf("cannot get executable: %v\n", err)
		os.Exit(1)
	}
	exPath := filepath.Dir(ex)

	file, err := os.ReadFile(filepath.Join(exPath, "config.yaml"))
	if err != nil {
		log.Fatalf("cannot read config file: %v\n", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("cannot unmarshal config file: %v\n", err)
		os.Exit(1)
	}

	// Initialize minio client object.
	useSSL := true
	client, err := minio.New(config.EndPoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			config.AccessUser, config.AccessPassword, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("cannot connect to minio: %v\n", err)
		os.Exit(1)
	}

	ctx := context.TODO()
	for _, file := range os.Args[1:] {
		ext := filepath.Ext(file)
		fileName := filepath.Base(file)
		fileNameWithoutExt := fileName[:len(fileName)-len(ext)]
		contentType := mime.TypeByExtension(filepath.Ext(file))
		if contentType == "" {
			contentType = "text/plain"
		}
		newName := fmt.Sprintf("%s-%s%s",
			fileNameWithoutExt, GetBase58Md5(file), ext)
		_, err := client.FPutObject(
			ctx, config.Bucket, newName, file,
			minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			log.Fatalf("failed to upload: %v\n", err)
			return
		}
		fmt.Printf("https://%s/%s/%s\n",
			config.EndPoint, config.Bucket, newName)
	}

}

func GetBase58Md5(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Cannot read file: %v", err)
		return ""
	}

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatalf("Cannot hash file: %v", err)
		return ""
	}

	return base58.Encode(hash.Sum(nil))
}
