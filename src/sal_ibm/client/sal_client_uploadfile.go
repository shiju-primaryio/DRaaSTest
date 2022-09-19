// Copyright 2022 PrimaryIO. All rights reserved.

package main

import (
	 "encoding/json"
	 "bytes"
	 "fmt"
	 "time"
	 "os"
	 "io"
	 "io/ioutil"
	 "net/http"
	 "github.com/IBM/ibm-cos-sdk-go/service/s3"
	 "log"
         "flag"
	 "path/filepath"
)

func listbuckets() {

	 client := &http.Client{}

	 req, err := http.NewRequest("GET", "http://localhost:8080/listProtectedVms", nil)
	 if err != nil {
	  fmt.Print(err.Error())
	 }

	 req.Header.Add("Accept", "application/json")
	 req.Header.Add("Content-Type", "application/json")
	 resp, err := client.Do(req)
	 if err != nil {
	  fmt.Print(err.Error())
	 }
	 defer resp.Body.Close()
	 bodyBytes, err := ioutil.ReadAll(resp.Body)
	 if err != nil {
	  fmt.Print(err.Error())
 	}
  
 	fmt.Print(string(bodyBytes))
}

func listbuckets_simple() {

	 resp, err := http.Get("http://localhost:8080/listProtectedVms")
	 if err != nil {
	  fmt.Print(err.Error())
	 }
	 bodyBytes, err := ioutil.ReadAll(resp.Body)
	 if err != nil {
	  fmt.Print(err.Error())
	 }
	 fmt.Print(string(bodyBytes))
    
}

func createbuckets() {
	 jsonData := map[string]string {"vmbucketname": "rahulk341-test31"}
	 jsonValue, _ := json.Marshal(jsonData)
	 resp, err := http.Post("http://localhost:8080/createVMBucket","application/json",bytes.NewBuffer(jsonValue))
	 if err != nil {
	  fmt.Print(err.Error())
	 }
	 bodyBytes, err := ioutil.ReadAll(resp.Body)
	 if err != nil {
	  fmt.Print(err.Error())
	 }
	 fmt.Print(string(bodyBytes))
	    
}

type StartUploadResponse struct {
 	UploadId     string `json:"UploadId"`
}


func StartUpload(bucketname string, objkey string) string {

	 jsonData := map[string]string {"vmbucketname": bucketname,"vmdkname": objkey }
	 jsonValue, _ := json.Marshal(jsonData)
	 resp, err := http.Post("http://localhost:8080/startUploadFileObj","application/json",bytes.NewBuffer(jsonValue))
	 if err != nil {
	  	fmt.Print(err.Error())
          	return ""
 	 }
	 bodyBytes, err := ioutil.ReadAll(resp.Body)
	 if err != nil {
	  	fmt.Print(err.Error())
          	return ""
 	 }
	 //fmt.Print(string(bodyBytes))

	 var responseObject StartUploadResponse
	 json.Unmarshal(bodyBytes, &responseObject)

	 //fmt.Printf("\nAPI Response as struct %+v\n", responseObject)
	 //fmt.Printf("\nUploadId is %+v\n", responseObject.UploadId)
         return responseObject.UploadId 
}

type UploadRequest struct {
        VmBucketName string `json:"vmbucketname"`
	VmdkName string `json:"vmdkname"`
        //ObjKey string `json:"objkey"`
        UploadId string `json:"uploadId"`
        FileBytes []byte `json:"fileBytes"`
        PartNumber int `json:"partNumber"`
}

type RequestComplete struct {
        VmBucketName string `json:"vmbucketname"`
	VmdkName string `json:"vmdkname"`
        //ObjKey string `json:"objkey"`
        UploadId string `json:"uploadId"`
	CompletedUploadParts []*s3.CompletedPart `json:"completedUploadParts"`
}

type UploadPartResponse struct {
	 ETag string `json:"ETag"`
	 PartNumber int `json:"PartNumber"`
	//ETag":"\"37b15ab4d3226d1f2325f03ea526e375\"","PartNumber
}
func UploadPartData(vmbucketname string, objkey string,uploadId string, partId int, fileBytes []byte ) (*s3.CompletedPart, error) {
	
         var responseObject s3.CompletedPart

	 RequestObj := UploadRequest{
		VmBucketName: vmbucketname,
		VmdkName: objkey,
        	UploadId: uploadId,
        	FileBytes: fileBytes,
        	PartNumber: partId,
	 }
 
	 jsonValue, _ := json.Marshal(RequestObj)
	 //fmt.Print(string(jsonValue))
	 resp, err := http.Post("http://localhost:8080/uploadPartFileObj","application/json",bytes.NewBuffer(jsonValue))
	 if err != nil {
	  	fmt.Print(err.Error())
          	return nil,err
 	 }
	 bodyBytes, err := ioutil.ReadAll(resp.Body)
	 if err != nil {
	  	fmt.Print(err.Error())
          	return nil,err
 	 }
	 //fmt.Print(string(bodyBytes))

	 err = json.Unmarshal(bodyBytes, &responseObject)

	 //fmt.Printf("\n\n Upload part API Response as struct %+v\n", responseObject)
	 log.Printf("Uploaded part %+v of file %+v", partId,objkey)
         return &responseObject,err 
}

func UploadFilePartByPart(bucketname string, objkey string,uploadId string) ([]*s3.CompletedPart) {
        sourceFileStat, err := os.Stat(objkey)
        if err != nil {
                //return 0, err
		return nil 
        }
        if !sourceFileStat.Mode().IsRegular() {
                //return 0, fmt.Errorf("%s is not a regular file", objkey)
        	return nil
        }
        source, err := os.Open(objkey)
        if err != nil {
                //return 0, err
		return nil 
        }
        defer source.Close()


        buf := make([]byte,BUFFERSIZE)
	var partNum = 1 
        var completedParts []*s3.CompletedPart
	//var completedParts []*s3.CompleteMultipartUploadOutput
        for {
                n, err := source.Read(buf)
                if err != nil && err != io.EOF {
                        //return err
                        return nil
                }
                if n == 0 {
                        break
                }

		/*
                if _, err := destination.Write(buf[:n]); err != nil {
                        return err
                }
		*/
		completedPart, err := UploadPartData(bucketname, objkey, uploadId, partNum, buf[:n] )
		if err != nil {
                        fmt.Println(err.Error())
			return nil
		}
		completedParts = append(completedParts, completedPart)
		partNum = partNum+1
        	//return
        }
	//fmt.Printf("\nCompleted Parts API Response as struct %+v\n", completedParts)
	return completedParts
}


func CompleteUpload(vmbucketname string, objkey string, uploadId string, completedUploadParts []*s3.CompletedPart) string {

	RequestObj := RequestComplete{
		VmBucketName: vmbucketname,
		VmdkName: objkey,
        	UploadId: uploadId,
		CompletedUploadParts: completedUploadParts,
	} 
	 jsonValue, _ := json.Marshal(RequestObj)
	 resp, err := http.Post("http://localhost:8080/completeUploadFileObj","application/json",bytes.NewBuffer(jsonValue))
	 if err != nil {
	  	fmt.Print(err.Error())
          	return ""
 	 }
	 bodyBytes, err := ioutil.ReadAll(resp.Body)
	 if err != nil {
	 	fmt.Print(string(bodyBytes))
	  	fmt.Print(err.Error())
          	return ""
 	 }
	 //fmt.Print(string(bodyBytes))

	 return "upload completed"
}


func uploadFile(bucketname string, objkey string)  {
	fmt.Printf("\n")
   	start := time.Now()
         log.Printf("Starting upload of file %+v into bucket %+v \n",objkey,bucketname)

	 //Step 1: Start Upload File Object
         uploadId := StartUpload(bucketname,objkey)
	 if uploadId == "" {
         	log.Printf("Unable to upload file %s . Please check bucket exists or not. \n",objkey)
		return
	 }
	 log.Printf("Started upload of file %+v . UploadId is %+v\n", objkey,uploadId)

         //Step 2: Upload Object part by Part
	 completedParts := UploadFilePartByPart(bucketname, objkey,uploadId)
	 //fmt.Printf("\nCompleted Parts API Response as struct %+v\n", completedParts)
	 
         //Step 3: Complete the Upload 
         ret := CompleteUpload(bucketname,objkey,uploadId,completedParts)
	 if ret == "" {
		return
	 }
	 
   	timeElapsed := time.Since(start)
   	log.Printf("-------- The File upload of file %s took %s -------- \n\n", objkey, timeElapsed)
}

var BUFFERSIZE = 5 * 1024 * 1024  //BufferSize.. UploadSize 5MB default upload size

func main() {

    //Command Line Parameter parsing
    UploadPartSizeinMBPtr := flag.Int("UploadPartSize", 5, 
                "Upload part size in MB to be uploaded. UploadPartSize >=5 ")    
    fileNamePtr := flag.String("Filename", "test_media_VAIO7.mp4", 
                "Filename to be uploaded. Path could be relative or absolute path. ")    
    bucketNamePtr := flag.String("Bucketname", "rahulk341-test31", 
                "Bucketname to be used for file upload. Please note that bucketname should exist.")    
    flag.Parse()    //---print out the message---

    if (*UploadPartSizeinMBPtr < 5) {
	fmt.Printf("\nError: UploadPartSize should be more than 5MB.\n\n")
	flag.Usage()
        return
    }
 

    BUFFERSIZE = *UploadPartSizeinMBPtr * 1024 * 1024

    pFileName, err := filepath.Abs(*fileNamePtr)

    if err != nil {

        log.Fatal(err)
    }

    stat, err := os.Stat(*fileNamePtr)
    if err != nil {

    	fmt.Printf("\nError: Filename %s does not exist. Please check and execute again ..\n\n",*fileNamePtr)
	flag.Usage()
	return
    }

    filename := filepath.Base(pFileName)
    var FileSizeInMB float64
    FileSizeInMB = (float64) (stat.Size())/(1024*1024) 
   
 
    fmt.Printf("\nFileName to be uploaded is %s \nFile size in MB is %f ",filename,FileSizeInMB)

    bucketName := *bucketNamePtr
    
    //listbuckets()
    //listbuckets_simple() 
    //createbuckets()
   //uploadFile("rahulk341-test31",filename)
   uploadFile(bucketName,filename)
}
