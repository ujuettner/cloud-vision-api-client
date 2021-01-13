package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"cloud.google.com/go/storage"
	vision "cloud.google.com/go/vision/apiv1"
	flag "github.com/spf13/pflag"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func main() {
	gcsBucket := flag.StringP("bucket", "b", "", "GCS bucket")
	imageFile := flag.StringP("image", "i", "", "image file")
	uploadTimeoutSecs := flag.Float64P("timeout", "t", 60.0, "timeout in seconds for image upload")
	flag.Parse()

	if *gcsBucket == "" {
		fmt.Println("No GCS bucket given!")
		os.Exit(2)
	}
	if *imageFile == "" {
		fmt.Println("No image file given!")
		os.Exit(2)
	}

	ctx := context.Background()

	uri, err := uploadImage(ctx, *imageFile, *gcsBucket, time.Second*time.Duration(*uploadTimeoutSecs))
	if err != nil {
		fmt.Printf("Error uploading the image %v to bucket %v: %v\n", *imageFile, *gcsBucket, err)
		return
	}

	// TODO: extract to func
	iac, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		fmt.Printf("Error creating the image annotator client: %v\n", err)
		return
	}
	defer iac.Close()

	res, err := iac.AnnotateImage(ctx, &visionpb.AnnotateImageRequest{
		Image: vision.NewImageFromURI(uri),
		Features: []*visionpb.Feature{
			{Type: visionpb.Feature_LANDMARK_DETECTION, MaxResults: 8},
			{Type: visionpb.Feature_LABEL_DETECTION, MaxResults: 4},
		},
	})
	if err != nil {
		fmt.Printf("Error getting image annotations: %v\n", err)
		return
	}

	// TODO: format output???
	// see https://pkg.go.dev/google.golang.org/genproto/googleapis/cloud/vision/v1#AnnotateImageResponse
	fmt.Printf("%v\n", res)
}

func uploadImage(ctx context.Context, imageFile, bucket string, timeout time.Duration) (string, error) {
	sc, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("Error creating the storage client: %v", err)
	}
	defer sc.Close()

	imageFileHandle, err := os.Open(imageFile)
	if err != nil {
		return "", fmt.Errorf("Error opening image file %v: %v", imageFile, err)
	}
	defer imageFileHandle.Close()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	object := sc.Bucket(bucket).Object(path.Base(imageFile))
	writer := object.NewWriter(ctx)
	if _, err = io.Copy(writer, imageFileHandle); err != nil {
		return "", fmt.Errorf("Error copying image file : %v", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("Error closing the writer: %v", err)
	}

	return fmt.Sprintf("gs://%s/%s", object.BucketName(), object.ObjectName()), nil
}
