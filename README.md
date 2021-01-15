A simple client program to play around with [Cloud Vision](https://cloud.google.com/vision/) API (Go package docs are [here](https://pkg.go.dev/cloud.google.com/go/vision/apiv1)).

## Auth & Enabling Cloud Vision API

1. `gcloud auth application-default login`
2. `gcloud services enable vision.googleapis.com`
   * Checker whether the API is enabled: `gcloud services list`

## Create a GCS Bucket

1. `gsutil mb gs://cloud-vision-api-client-test-001`
2. Set bucket's lifecycle to delete all objects older than 1 day: `gsutil lifecycle set bucket-lifecycle.json gs://cloud-vision-api-client-test-001/`
   * Check bucket's lifecycle: `gsutil lifecycle get gs://cloud-vision-api-client-test-001/`
3. Disable versioning: `gsutil versioning set off gs://cloud-vision-api-client-test-001/`
   * Check whether versioning is disabled: `gsutil versioning get gs://cloud-vision-api-client-test-001/`
