// Copyright 2022 PrimaryIO. All rights reserved.

package main

import (
	 "encoding/json"
	 "bytes"
	 "fmt"
//	 "time"
	 "os"
	 "io"
	 "io/ioutil"
	 "net/http"
	 "github.com/IBM/ibm-cos-sdk-go/service/s3"
	 "log"
         "flag"
	 "path/filepath"
)

type addVaioDataRequest struct {
        VmBucketName string `json:"vmbucketname"`
        VmdkName string `json:"vmdkname"`
        BlockNumber int `json:"blocknumber"`
        BlockData []byte `json:"blockdata"`
}

type addVaioDataResponse struct {
        RetString string `json:"retstring"`
}

func sendVaioData(vmbucketname string, vmdkname string, blocknumber int, blockdata []byte ) (string, error) {

         var responseObject s3.CompletedPart

	 RequestVaioObj := addVaioDataRequest{
		VmBucketName: vmbucketname,
		VmdkName: vmdkname,
		BlockNumber: blocknumber,
		BlockData: blockdata,
	}

	jsonValue, _ := json.Marshal(RequestVaioObj)
	 //fmt.Print(string(jsonValue))
	 resp, err := http.Post("http://localhost:8080/addVaioObj","application/json",bytes.NewBuffer(jsonValue))
	 if err != nil {
		fmt.Print(err.Error())
		return "error",err
	}
	 bodyBytes, err := ioutil.ReadAll(resp.Body)
	 if err != nil {
	  	fmt.Print(err.Error())
          	return "error",err
 	 }
	 //fmt.Print(string(bodyBytes))

	 err = json.Unmarshal(bodyBytes, &responseObject)

	 //log.Printf("Sent VAIO Data %+v of file", objkey)
	 return "success",err
}


func addVaioData(bucketname string, objkey string) {
        sourceFileStat, err := os.Stat(objkey)
        if err != nil {
                fmt.Errorf("Error while doing stat for the file %s", objkey)
		return
        }
        if !sourceFileStat.Mode().IsRegular() {
                fmt.Errorf("Error:file %s is not regular file", objkey)
		return
        }
        source, err := os.Open(objkey)
        if err != nil {
                fmt.Errorf("Error:file %s open failed", objkey)
		return
        }
        defer source.Close()


        buf := make([]byte,BUFFERSIZE)
	var blockNum = 1 
        for {
                n, err := source.Read(buf)
                if err != nil && err != io.EOF {
                        //return err
                        //return nil
			return
                }
                if n == 0 {
                        break
                }

		/*
                if _, err := destination.Write(buf[:n]); err != nil {
                        return err
                }
		*/
		retString, err := sendVaioData(bucketname, objkey,blockNum, buf[:n] )
		if err != nil {
                        fmt.Println(err.Error())
                        fmt.Println(retString)
			return 
		}
		fmt.Printf("\nAdded block number %d of file %s into the bucket %s\n", blockNum,objkey,bucketname)
		blockNum = blockNum+1
        }
	return
}

var BUFFERSIZE = 4 * 1024 * 1024  //BufferSize.. 4MB default buffer size

func main() {

    //Command Line Parameter parsing
    BufferSizeinMBPtr := flag.Int("BufferSize", 4, 
                "Buffer size in MB to be used to send the filedata. BufferSize >=4 ")    
    fileNamePtr := flag.String("Filename", "test_media_VAIO7.mp4", 
                "Filename to be uploaded. Path could be relative or absolute path. ")    
    bucketNamePtr := flag.String("Bucketname", "rahulk-test19", 
                "Bucketname to be used for file upload. Please note that bucketname should exist.")    
    flag.Parse()    //---print out the message---

    if (*BufferSizeinMBPtr < 4) {
	fmt.Printf("\nError: BufferSize should not be less than 4MB.\n\n")
	flag.Usage()
        return
    }
 

    BUFFERSIZE = *BufferSizeinMBPtr * 1024 * 1024

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
   
 
    fmt.Printf("\n FileName %s has filesize %f MB \n",filename,FileSizeInMB)

    bucketName := *bucketNamePtr
    
   addVaioData(bucketName,filename)
}
