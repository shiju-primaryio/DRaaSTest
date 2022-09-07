// Copyright 2022 PrimaryIO. All rights reserved.

package main

import "C"

import (
	"fmt"
        "os"
        "bytes"
        "io/ioutil"
        "github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
//	"github.com/IBM/ibm-cos-sdk-go/aws/awserr"
//	"strings"
//	"github.com/IBM/ibm-cos-sdk-go/service/s3/s3manager"
//       "log"
//        "io"
)

var region = "us-south"
var svc *s3.S3

// Constants for IBM COS values
const (
    apiKey            = "jfAqZFDTqJrG5E9t6w1kW98gnG0ZHFqJEdXHlwA9TfuD"
    serviceInstanceID = "crn:v1:bluemix:public:cloud-object-storage:global:a/573fa71d42694fb78477738a1c20dc41:86a44647-4731-465a-80b6-32a9ebb81e27::"
    authEndpoint      = "https://iam.cloud.ibm.com/identity/token"
    serviceEndpoint   = "https://s3.us-south.cloud-object-storage.appdomain.cloud"
    bucketLocation    = "us-south"
)

var buckets []string

// List all of your available buckets in IBM cloud
func listBuckets() []string  {
  buckets = nil
  result, err := svc.ListBuckets(nil)

  if err != nil {
    //exitErrorf("Unable to list buckets, %v", err)
    return nil
  }
  fmt.Printf("listbucket \n")
  for _, b := range result.Buckets {
	buckets = append(buckets,aws.StringValue(b.Name))
  }
  return buckets
}

// Create a bucket (VM)
func createBucket(bucketName string) string {
  var returnString string
  fmt.Printf("\nCreating a new bucket named '" + bucketName + "'...\n")

  _, err := svc.CreateBucket(&s3.CreateBucketInput{
 	  Bucket: aws.String(bucketName),
	   CreateBucketConfiguration: &s3.CreateBucketConfiguration{
	     LocationConstraint: aws.String(region),
	   },
	 })

  if err != nil {
    //exitErrorf("Unable to create bucket, %v", err)
    returnString = fmt.Sprintf("Unable to create bucket, %v", err)
    fmt.Printf(returnString+"\n")
    return(returnString)
  }
  
  // Wait until bucket is created before finishing
  fmt.Printf("Waiting for bucket %q to be created...\n", bucketName)

  err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
    Bucket: aws.String(bucketName),
  })
  returnString = fmt.Sprintf("Bucket %s is created sucessfully..\n", bucketName)
  fmt.Printf(returnString+"\n")
  return(returnString)
}

// Delete the bucket (VM)
func deleteBucket(bucketName string) string {
  var returnString string
  fmt.Printf("\nDeleting the bucket named '" + bucketName + "'...\n")

  _, err := svc.DeleteBucket(&s3.DeleteBucketInput{
    Bucket: aws.String(bucketName),
  })

  if err != nil {
    //exitErrorf("Unable to delete bucket, %v", err)
    returnString = fmt.Sprintf("Unable to delete the bucket, %v", err)
    fmt.Printf(returnString+"\n")
    return(returnString)
  }
  
  // Wait until bucket is deleted before finishing
  fmt.Printf("Waiting for bucket %q to be deleted...\n", bucketName)
 
  err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
    Bucket: aws.String(bucketName),
  })
  returnString = fmt.Sprintf("Bucket %s is deleted sucessfully..\n", bucketName)
  fmt.Printf(returnString+"\n")
  return(returnString)
}


func exitErrorf(msg string, args ...interface{}) {
  // If there's an error, display it.
  fmt.Fprintf(os.Stderr, msg+"\n", args...)
  os.Exit(1)
}

// Write the object into the bucket (synchronous) . If object already exists, it is overwritten, otherwise it will be newly created 
func writeSyncObjectBucket(svc *s3.S3, bucketName string, s3_object_name string,data string) string {
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
                return err.Error()
        } else {
                fmt.Println(result)
		returnString := fmt.Sprintf("Object %s is written sucessfully..\n", key)
		return returnString
        }
    }
    return "Unable to write Object into Obj Store"
}

// Read the Object from the bucket (synchronous)
func readSyncObjectBucket(svc *s3.S3, bucketName string, s3_object_name string) string {
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
                return err.Error()
        } else {
		data, _ := ioutil.ReadAll(result.Body)
		fmt.Println(string(data))
		return string(data)
        }
    }
    return "Unable to retieve Object from Obj Store"
}


func setupIBMCloud() {

  // Create config
  var conf = aws.NewConfig().
    WithRegion("us-south").
    WithEndpoint(serviceEndpoint).
    WithCredentials(ibmiam.NewStaticCredentials(aws.NewConfig(), authEndpoint, apiKey, serviceInstanceID)).
    WithS3ForcePathStyle(true)

  s3Config := aws.NewConfig()
  s3Config.CredentialsChainVerboseErrors = aws.Bool(true)

  sess, err := session.NewSession(s3Config)
  if err != nil {
    fmt.Printf("Error initializing s3 uploader. %v" + err.Error())
    os.Exit(0)
  }
  svc = s3.New(sess, conf)
}
