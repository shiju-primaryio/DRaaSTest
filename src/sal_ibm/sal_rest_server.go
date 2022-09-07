// Copyright 2022 PrimaryIO. All rights reserved.

package main

import "C"

import (
	"fmt"
        "time"
        "os"
        "io"
        "log"

        //REST Server Gin
        "net/http"
	"github.com/gin-gonic/gin"
)

// List all of your available buckets (Vms related data) in IBM cloud
func listProtectedVms(c *gin.Context) {
  result := listBuckets()

  if result == nil {
    fmt.Printf("\n Error: Listing of buckets failed...\n\n")
    c.IndentedJSON(http.StatusInternalServerError, "Error: Unable to list protected Vms...")
    return
  }

  w := c.Writer
	header := w.Header()
	header.Set("Transfer-Encoding", "chunked")
	header.Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<html>
		<body>
		<h1> List Of Protected VMs (list of buckets in IBM Cloud): </h1>
  `))
  w.(http.Flusher).Flush()
  for _,b := range result {
	w.Write([]byte(fmt.Sprintf(`
		<h4> %s</h4>
		`, b)))
		w.(http.Flusher).Flush()
    		//fmt.Printf(aws.StringValue(b.Name) + "\n")
  }
  w.Write([]byte(`
		
		</body>
		</html>
  `))
  w.(http.Flusher).Flush()
}

func jsonWriteMessage(c *gin.Context, retString string) {

    w := c.Writer
    header := w.Header()
    header.Set("Transfer-Encoding", "chunked")
    header.Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf(`
		<html>
		<body>
		<h1> %s </h1>
		</body>
		</html>
        `,retString)))
    w.(http.Flusher).Flush()
}

var newVMName struct {
	Name string `json:"name"`
}

// createVMBucket adds an bucket from JSON received in the request body.
func createVMBucket(c *gin.Context) {

    // Call BindJSON to bind the received JSON to bucket
    if err := c.BindJSON(&newVMName); err != nil {
	fmt.Printf("\n Error: Creating a new bucket failed...\n\n")
	c.IndentedJSON(http.StatusInternalServerError, "Error: Unable to create bucket...")
        return
    }
 
    retString  := createBucket(newVMName.Name)

    jsonWriteMessage(c,retString)

}

// deleteVMBucket deletes the bucket from JSON received in the request body.
func deleteVMBucket(c *gin.Context) {

    // Call BindJSON to bind the received JSON to bucket
    if err := c.BindJSON(&newVMName); err != nil {
        fmt.Printf("\n Error: Deleting the bucket failed...\n\n")
	c.IndentedJSON(http.StatusInternalServerError, "Error: Unable to delete the bucket...")
        return
    }
 
    fmt.Printf("\nDeleting the bucket named '" + newVMName.Name + "'...\n\n")
    retString  := deleteBucket(newVMName.Name)

    jsonWriteMessage(c,retString)
    //c.IndentedJSON(http.StatusCreated, newVMName.Name)

}

var getObjName struct {
	BucketName string `json:"bucketname"`
	ObjKey string `json:"objkey"`
}

//Get VAIO Obj
func getVaioObj(c *gin.Context) {

    var retString string

    // Call BindJSON to bind the received JSON to getNewObjName structure
    if err := c.BindJSON(&getObjName); err != nil {
        retString = "\nError: Getting the object from bucket failed...\n\n"
	c.IndentedJSON(http.StatusInternalServerError, retString)
        return
    }

    fmt.Printf("\nGetting the object from bucket " + getObjName.BucketName +" with key "+ getObjName.ObjKey + "...\n\n")
    retString = readSyncObjectBucket(svc, getObjName.BucketName, getObjName.ObjKey)

    jsonWriteMessage(c,retString)
}

var addNewObjName struct {
	BucketName string `json:"bucketname"`
	ObjKey string `json:"objkey"`
	Data string `json:"data"`
}

//Add VAIO Obj
func addVaioObj(c *gin.Context) {

    var retString string

    // Call BindJSON to bind the received JSON to addNewObjName structure
    if err := c.BindJSON(&addNewObjName); err != nil {
        retString = "\nError: Adding the object into the bucket failed...\n\n"
	c.IndentedJSON(http.StatusInternalServerError, retString)
        return
    }

    retString = writeSyncObjectBucket(svc, addNewObjName.BucketName, addNewObjName.ObjKey, addNewObjName.Data)

    jsonWriteMessage(c,retString)
}

// Create Log file and start logging 
func startLogging() {

    //Create directory for logging
    if err := os.MkdirAll("sal_logs", os.ModePerm); err != nil {
        log.Fatal(err)
    }
    
    // Logs will be stored at filename.
    filename := "sal_logs/sal_rest_server_"+time.Now().Format("2006-01-02-15-04-05")+".log"
    f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        exitErrorf("Unable to create log file %v", err)
    }
    fmt.Printf("\n Starting Gin Rest Server at port 8080.\n\n Note: Logs will be stored at " + filename +" ...\n\n")

    gin.DefaultWriter = io.MultiWriter(f)
    // If you need to write the logs to file and console at the same time.
    //gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}

func main() {

    // Force log's color
    gin.ForceConsoleColor()

    rest_server := gin.Default()


    // Setup IBM Cloud environment with credentials, endpoint etc.
    setupIBMCloud()

    //start Logging
    startLogging()

    rest_server.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
    	// your custom format
	return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
	)
     }))


    // GET & POST Requests
    rest_server.GET("/listProtectedVms", listProtectedVms) // List buckets
    rest_server.POST("/createVMBucket", createVMBucket)    // Create VM Bucket
    rest_server.POST("/deleteVMBucket", deleteVMBucket)    // Delete VM Bucket
    rest_server.GET("/getVaioObj", getVaioObj)      	   // Retrieve Object
    rest_server.POST("/addVaioObj", addVaioObj)            // Add VAIO Object

    // Start the server 
    rest_server.Run("localhost:8080")
}
