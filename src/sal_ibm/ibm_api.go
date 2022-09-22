// Copyright 2022 PrimaryIO. All rights reserved.

package main

import "C"

import (
	"fmt"
        "os"
        "bytes"
        "time"
        "io/ioutil"
        "net/http"
        "github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"github.com/IBM/ibm-cos-sdk-go/aws/awserr"
	"strconv"
//	"strings"
//	"github.com/IBM/ibm-cos-sdk-go/service/s3/s3manager"
//       "log"
//        "io"
)

var region = "us-south"
var svc *s3.S3

// Constants for IBM COS values
const (
    maxPartSize        = int64(5 * 1024 * 1024) //5MB
    maxRetries         = 3
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

// Set versioning configuration on a bucket
func enableBucketVersioning(bucketName string) {
	//svc := s3.New(session.New())
	input := &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3.VersioningConfiguration{
			MFADelete: aws.String("Disabled"),
			Status:    aws.String("Enabled"),
		},
	}

	result, err := svc.PutBucketVersioning(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
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

  //Enable Versioning for bucket
  enableBucketVersioning(bucketName)

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
func writeSyncObjectBucket(svc *s3.S3, bucketName string, vmdk_name string,blockNumber int, blockData []byte) (string,error) {
    //Object name is FileName_BlockNumber
    key := vmdk_name+"_"+strconv.Itoa(blockNumber)
    content := bytes.NewReader(blockData)

    input := s3.PutObjectInput{
        Bucket:        aws.String(bucketName),
        Key:           aws.String(key),
        Body:          content,
    }

    retry_iter := 1
    for retry_iter <= 5 {
    	// Call Function to upload (Put) an object 
        //fmt.Println("writing object .. ")
    	result, err := svc.PutObject(&input)

        if err != nil {
                // Print the error, cast err to awserr.Error to get the Code and
                // Message from an error.
                fmt.Println(err.Error())
                return err.Error(),err
        } else {
                fmt.Println(result)
		returnString := fmt.Sprintf("Object %s is written sucessfully..\n", key)
		return returnString,nil
        }
    }
    return "Unable to write Object into Obj Store",nil
}

// Read the Object from the bucket (synchronous)
func readSyncObjectBucket(svc *s3.S3, bucketName string, vmdk_name string,blockNumber int) string {
    //Object name is FileName_BlockNumber
    key := vmdk_name+"_"+strconv.Itoa(blockNumber)

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


// Read the specific version of Object from the bucket at specific UTC time
func readSyncObjectMyBucketVersion(svc *s3.S3, bucketName string, vmdk_name string,blockNumber int, utcLatestTime string) string {

    //Object name is FileName_BlockNumber
    key := vmdk_name+"_"+strconv.Itoa(blockNumber)

    //Fetch version-id of object at specific UTC time
    obj_version_id := fetchObjectVersionIdAtUtcTime(svc,bucketName,key,utcLatestTime)

    var input s3.GetObjectInput

    if (obj_version_id != "") {
	    // users will need to create bucket, key (flat string name)
	    input = s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		VersionId: aws.String(obj_version_id),
	    }
    } else {
	    // users will need to create bucket, key (flat string name)
	    input = s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	    }
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

func fetchObjectVersionIdAtUtcTime(svc *s3.S3,bucketname string, objkey string,utclatesttime string) string{

	//List Object Versions
	input := &s3.ListObjectVersionsInput{
	    Bucket: aws.String(bucketname),
	    Prefix: aws.String(objkey),
	}

	result, err := svc.ListObjectVersions(input)
	if err != nil {
	    if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		default:
		    fmt.Println(aerr.Error())
		}
	    } else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
	    }
	    return ""
	}
        latestdatetime, _ := time.Parse("2006-01-02 15:04:05 -0700 UTC", utclatesttime)
        for _, object := range result.Versions {
		objectdatetimeModified, _ := time.Parse("2006-01-02 15:04:05 -0700 UTC", (*object.LastModified).UTC().String())
		if latestdatetime.After(objectdatetimeModified) {
			fmt.Printf("versionId: %v  %v \n", *object.VersionId,*object.LastModified)
			return *object.VersionId
		}
		fmt.Printf("versionId: %v  %v \n", *object.VersionId,*object.LastModified)
	}
	fmt.Println(result)
	return ""
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


func uploadFileObjectBucket(svc *s3.S3, bucketName string, s3_object_name string, file_name string) string {

	file, err := os.Open(file_name)
	if err != nil {
		fmt.Printf("err opening file: %s", err)
		return "err opening file"
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	fileType := http.DetectContentType(buffer)
	file.Read(buffer)

	path := file.Name()
	input := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s3_object_name),
		Key:         aws.String(path),
		ContentType: aws.String(fileType),
	}

	resp, err := svc.CreateMultipartUpload(input)
	if err != nil {
		fmt.Println(err.Error())
		return "Failed to uploaded file"
	}
	fmt.Println("Created multipart upload request")

	var curr, partLength int64
	var remaining = size
	var completedParts []*s3.CompletedPart
	partNumber := 1
	for curr = 0; remaining != 0; curr += partLength {
		if remaining < maxPartSize {
			partLength = remaining
		} else {
			partLength = maxPartSize
		}
		completedPart, err := uploadPart(svc, resp, buffer[curr:curr+partLength], partNumber)
		if err != nil {
			fmt.Println(err.Error())
			err := abortMultipartUpload(svc, resp)
			if err != nil {
				fmt.Println("RK print1")
				fmt.Println(err.Error())
			}
			fmt.Println("RK print2")
			return "Failed to uploaded file"
		}
		remaining -= partLength
		partNumber++
		completedParts = append(completedParts, completedPart)
	}

	completeResponse, err := completeMultipartUpload(svc, resp, completedParts)
	if err != nil {
		fmt.Println(err.Error())
		return err.Error()
	}

	fmt.Printf("Successfully uploaded file: %s\n", completeResponse.String())
	return "Successfully uploaded file"
}


func completeMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, completedParts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}
	return svc.CompleteMultipartUpload(completeInput)
}

func uploadPart(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, fileBytes []byte, partNumber int) (*s3.CompletedPart, error) {
	tryNum := 1
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(fileBytes),
		Bucket:        resp.Bucket,
		Key:           resp.Key,
		PartNumber:    aws.Int64(int64(partNumber)),
		UploadId:      resp.UploadId,
		ContentLength: aws.Int64(int64(len(fileBytes))),
	}

	for tryNum <= maxRetries {
		uploadResult, err := svc.UploadPart(partInput)
		if err != nil {
			if tryNum == maxRetries {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				}
				return nil, err
			}
			fmt.Printf("Retrying to upload part #%v\n", partNumber)
			tryNum++
		} else {
			fmt.Printf("Uploaded part #%v\n", partNumber)
			return &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}, nil
		}
	}
	return nil, nil
}

func abortMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput) error {
	fmt.Println("Aborting multipart upload for UploadId#" + *resp.UploadId)
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
	}
	_, err := svc.AbortMultipartUpload(abortInput)
	return err
}

func createMultipartUploadObjectBucket(svc *s3.S3, bucketName string, s3_object_name string) (*s3.CreateMultipartUploadOutput, error) {

	input := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(s3_object_name),
	//	ContentType: aws.String(fileType), //TODO: Need to check whether this field is required
	}

	resp, err := svc.CreateMultipartUpload(input)
	if err != nil {
		fmt.Println(err.Error())
		return nil,err
		//return "Failed to uploaded file"
		//return
	}
	fmt.Println("Created multipart upload request")

	return resp,nil
}

func completeMultipartUploadObjectBucket(svc *s3.S3, bucketName string, s3_object_name string, UploadId string, completedParts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucketName),
		Key:      aws.String(s3_object_name),
		UploadId: aws.String(UploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}
	//fmt.Printf("\n\nComplete upload request  #%v\n",completeInput)
	return svc.CompleteMultipartUpload(completeInput)
}

func uploadPartObjectBucket(svc *s3.S3, bucketName string, s3_object_name string, UploadId string, fileBytes []byte, partNumber int) (*s3.CompletedPart, error) {
	tryNum := 1
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(fileBytes),
		Bucket:   aws.String(bucketName),
		Key:      aws.String(s3_object_name),
		PartNumber:    aws.Int64(int64(partNumber)),
		UploadId: aws.String(UploadId),
		ContentLength: aws.Int64(int64(len(fileBytes))),
	}

        //fmt.Printf("\nUploadimg a new obj named '" + bucketName+ " "+s3_object_name +" "+string(partNumber)+ "'...\n")
	//fmt.Printf("trying to upload part #%v\n",partInput)
	for tryNum <= maxRetries {
		uploadResult, err := svc.UploadPart(partInput)
		//fmt.Printf("Retrying to upload part #%v\n", uploadResult)
		if err != nil {
			if tryNum == maxRetries {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				}
				return nil, err
			}
			fmt.Printf("Retrying to upload part #%v\n", partNumber)
			tryNum++
		} else {
			fmt.Printf("Uploaded part #%v\n", partNumber)
			return &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}, nil
		}
	}
	return nil, nil
}
