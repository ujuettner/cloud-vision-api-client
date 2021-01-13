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
	timeout := flag.DurationP("timeout", "t", 30.0*time.Second, "timeout for all remote calls")
	maxResults := flag.Int32P("maxresults", "m", 3, "maximum number of results per category")
	flag.Parse()

	if *gcsBucket == "" {
		fmt.Println("No GCS bucket given!")
		os.Exit(2)
	}
	if *imageFile == "" {
		fmt.Println("No image file given!")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	uri, err := uploadImage(ctx, *imageFile, *gcsBucket)
	if err != nil {
		fmt.Printf("Error uploading the image %v to bucket %v: %v\n", *imageFile, *gcsBucket, err)
		return
	}

	res, err := getImageAnnotations(ctx, uri, *maxResults)
	if err != nil {
		fmt.Printf("Error getting image annotations: %v\n", err)
		return
	}

	printResults(res)
}

func uploadImage(ctx context.Context, imageFile, bucket string) (string, error) {
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

func getImageAnnotations(ctx context.Context, imageUri string, maxResults int32) (*visionpb.AnnotateImageResponse, error) {
	iac, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error creating the image annotator client: %v\n", err)
	}
	defer iac.Close()

	res, err := iac.AnnotateImage(ctx, &visionpb.AnnotateImageRequest{
		Image: vision.NewImageFromURI(imageUri),
		Features: []*visionpb.Feature{
			{Type: visionpb.Feature_LANDMARK_DETECTION, MaxResults: maxResults},
			{Type: visionpb.Feature_LOGO_DETECTION, MaxResults: maxResults},
			{Type: visionpb.Feature_LABEL_DETECTION, MaxResults: maxResults},
			{Type: visionpb.Feature_TEXT_DETECTION, MaxResults: maxResults},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Error getting image annotations: %v\n", err)
	}

	return res, nil
}

func printResults(res *visionpb.AnnotateImageResponse) {
	fmt.Println("Landmarks:")
	for i, v := range res.GetLandmarkAnnotations() {
		fmt.Printf("\t%d: %s (%f)\n", i, v.Description, v.Score)
	}
	fmt.Println("Logos:")
	for i, v := range res.GetLogoAnnotations() {
		fmt.Printf("\t%d: %s (%f)\n", i, v.Description, v.Score)
	}
	fmt.Println("Labels:")
	for i, v := range res.GetLabelAnnotations() {
		fmt.Printf("\t%d: %s (%f)\n", i, v.Description, v.Score)
	}
	fmt.Println("Text:")
	for i, v := range res.GetTextAnnotations() {
		fmt.Printf("\t%d: %s\n", i, v.Description)
	}
}
