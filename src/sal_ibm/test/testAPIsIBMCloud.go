// Copyright 2022 PrimaryIO. All rights reserved.

package main

import "C"

import (
	"fmt"
        "io/ioutil"
        "io"
        "bytes"
        "os"
        "github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
//	"github.com/IBM/ibm-cos-sdk-go/aws/awserr"
//	"strings"
//	"github.com/IBM/ibm-cos-sdk-go/service/s3/s3manager"
//       "log"
)

// Constants for IBM COS values
const (
    apiKey            = "jfAqZFDTqJrG5E9t6w1kW98gnG0ZHFqJEdXHlwA9TfuD"
    serviceInstanceID = "crn:v1:bluemix:public:cloud-object-storage:global:a/573fa71d42694fb78477738a1c20dc41:86a44647-4731-465a-80b6-32a9ebb81e27::"
    authEndpoint      = "https://iam.cloud.ibm.com/identity/token"
    serviceEndpoint   = "https://s3.us-south.cloud-object-storage.appdomain.cloud"
    bucketLocation    = "us-south"
)

// Create config

var conf = aws.NewConfig().
    WithRegion("jp-tok").
    WithEndpoint(serviceEndpoint).
    WithCredentials(ibmiam.NewStaticCredentials(aws.NewConfig(), authEndpoint, apiKey, serviceInstanceID)).
    WithS3ForcePathStyle(true)

// List all of your available buckets
func listMyBuckets(svc *s3.S3) {
  result, err := svc.ListBuckets(nil)

  if err != nil {
    exitErrorf("Unable to list buckets, %v", err)
  }

  fmt.Println("My buckets now are:\n")

  for _, b := range result.Buckets {
    fmt.Printf(aws.StringValue(b.Name) + "\n")
  }

  fmt.Printf("\n")
}

// Create a bucket
func createMyBucket(svc *s3.S3, bucketName string, region string) {
  fmt.Printf("\nCreating a new bucket named '" + bucketName + "'...\n\n")

  _, err := svc.CreateBucket(&s3.CreateBucketInput{
   Bucket: aws.String(bucketName),
   CreateBucketConfiguration: &s3.CreateBucketConfiguration{
     LocationConstraint: aws.String(region),
   },
 })

  if err != nil {
    exitErrorf("Unable to create bucket, %v", err)
  }
  
  // Wait until bucket is created before finishing
  fmt.Printf("Waiting for bucket %q to be created...\n", bucketName)

  err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
    Bucket: aws.String(bucketName),
  })
}

// Delete the bucket
func deleteMyBucket(svc *s3.S3, bucketName string) {
  fmt.Printf("\nDeleting the bucket named '" + bucketName + "'...\n\n")

  _, err := svc.DeleteBucket(&s3.DeleteBucketInput{
    Bucket: aws.String(bucketName),
  })

  if err != nil {
    exitErrorf("Unable to delete bucket, %v", err)
  }
  
  // Wait until bucket is deleted before finishing
  fmt.Printf("Waiting for bucket %q to be deleted...\n", bucketName)

 
 err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
    Bucket: aws.String(bucketName),
  })
}

func exitErrorf(msg string, args ...interface{}) {
  // If there's an error, display it.
  fmt.Fprintf(os.Stderr, msg+"\n", args...)
  os.Exit(1)
}


// Write the object into the bucket. If object already exists, it is overwritten, otherwise it will be newly created 
func writeSyncObjectMyBucket(svc *s3.S3, bucketName string, s3_object_name string,data string) {
    key := s3_object_name
    content := bytes.NewReader([]byte(data))

    input := s3.PutObjectInput{
        Bucket:        aws.String(bucketName),
        Key:           aws.String(key),
        Body:          content,
    }

    retry_iter := 1
    for retry_iter <= 5 {
    	// Call Function to upload (Put) an object 
    	result, err := svc.PutObject(&input)

        if err != nil {
                // Print the error, cast err to awserr.Error to get the Code and
                // Message from an error.
                fmt.Println(err.Error())
        } else {
                fmt.Println(result)
                return
                //return true
        }
    }
}

// Read the Object from the bucket
//func readSyncObjectMyBucket(svc *s3.S3, bucketName string, s3_object_name string) string {
func readSyncObjectMyBucket(svc *s3.S3, bucketName string, s3_object_name string) {
    key := s3_object_name

    // users will need to create bucket, key (flat string name)
    input := s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(key),
    }

    retry_iter := 1
    for retry_iter <= 5 {
    	// Call Function to download (Get) an object 
    	result, err := svc.GetObject(&input)

        if err != nil {
                // Print the error, cast err to awserr.Error to get the Code and
                // Message from an error.
                fmt.Println(err.Error())
        } else {
    		fmt.Println(result)
		f, _ := os.Create(key)
		defer f.Close()
		io.Copy(f, result.Body)

		fmt.Println("Downloaded", f.Name())

    		body, _ := ioutil.ReadAll(result.Body)
		s := string(body[:])
		fmt.Println("Downloaded: ",s)

                return 
        }
    }
}

///*
func main() {

  s3Config := aws.NewConfig()
  s3Config.CredentialsChainVerboseErrors = aws.Bool(true)

  sess, err := session.NewSession(s3Config)
  if err != nil {
    fmt.Printf("Error initializing s3 uploader. %v" + err.Error())
    os.Exit(0)
  }
  svc:= s3.New(sess, conf)

  listMyBuckets(svc)
  createMyBucket(svc, "rahulk31-test31", "us-south")
//  writeSyncObjectMyBucket(svc, "rahulk3-test3", "test_data","RK bcbcewcbwobcewocHello World!")
  readSyncObjectMyBucket(svc, "rahulk3-test3", "test_data")
  //data := readSyncObjectMyBucket(svc, "rahulk3-test3", "test_data")
  //fmt.Println(data)
  
  //writeSyncObjectMyBucket(svc, "rahulk3-test3", "test_data","Hello World!")
  //deleteMyBucket(svc, "rahulk3-test3")
  //listMyBuckets(svc)
}

//*/
