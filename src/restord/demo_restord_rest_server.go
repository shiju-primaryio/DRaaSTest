// Copyright 2022 PrimaryIO. All rights reserved.

package main

import "C"

import (
	"fmt"
        "time"
        "os"
        "os/exec"
        "io"
        "strings"
        "log"
	"encoding/json"
	"io/ioutil"
        "net"
        "net/http"
	"github.com/gin-gonic/gin"
)

var ipaddr string

// Get the DR Status
var getDRStatusForSiteRequest struct {
        SiteName string `json:"sitename"`
}

// VM configuration details
type VM struct {
	Name        string `json:"name,omitempty"`
	CPUs        int    `json:"num_cpus,omitempty"`
	Memory      int    `json:"memory,omitempty"`
	GuestID     string `json:"guest_id,omitempty"`
	Disks       []Disk `json:"disks,omitempty"`
}

// Disk configuration details
type Disk struct {
	UnitNumber      int    `json:"unit_number"`   // unit number can be zero as well
	Size            int    `json:"size,omitempty"`
	Label           string `json:"label,omitempty"`
	ThinProvisioned bool   `json:"thin_provisioned"`  // removed omitempty , as value could be false i.e. 0.
	Policyname      string `json:"policy_name"`  // This can be empty, so omitempty is not applicable 
}
// Start DR for Site
var startDRForSiteRequest struct {
        SiteName string `json:"sitename"`
        VmList  []VM `json:"vmlist,omitempty"`
}

var attachPolicyToVMsForSiteRequest struct {
        SiteName string `json:"sitename"`
        VmList  []VM `json:"vmlist,omitempty"`
}

var failover_progress string = "Failover_not_started"

func getDRStatusForSite(c *gin.Context) {
    c.JSON(http.StatusOK, failover_progress)
    return
}

var getDRVcenterDetailsRequest struct {
        SiteName string `json:"sitename"`
}
var getDRVcenterDetailsResponse struct {
        vCenterIP string `json:"vcenterip"`
        UserName string `json:"username"`
        Password string `json:"password"`
}

type vCenterDetails struct {
        VcenterIP string
        UserName string
        Password string
}

func getDRVcenterDetails(c *gin.Context) {
    b, err := os.ReadFile("demo_dr_infra_vm/vcenter_details.txt")
    if err != nil {
        fmt.Print(err)
	c.IndentedJSON(http.StatusInternalServerError, "Error:vCenter Server is still not created")
        return
    }

    vCenterDet := vCenterDetails{} 
    str := string(b) // convert content to a 'string'

    vCenterDet.VcenterIP= strings.TrimRight(str,"\r\n") // convert content to a 'string'
    vCenterDet.UserName= "administrator@primaryio.cloud"
    vCenterDet.Password= "PrimaryIO@123" 
    c.JSON(http.StatusOK, vCenterDet)

    return
}

func startDRForSite(c *gin.Context) {

    var retString string

    os.Remove("demo_dr_infra_vm/vcenter_details.txt")
    // Call BindJSON to bind the received JSON startDRForSiteRequest structure
    if err := c.BindJSON(&startDRForSiteRequest); err != nil {
        retString := "\nError: Checking startDRForSite failed...\n\n"
	c.IndentedJSON(http.StatusInternalServerError, retString)
        return
    }

    file, _ := json.MarshalIndent(startDRForSiteRequest, "", " ")

    _ = ioutil.WriteFile("demo_dr_infra_vm/CreateModifyVmList.json", file, 0644)
    //_ = ioutil.WriteFile("dr_infra_tf/CreateModifyVmList.json", file, 0644)

    resp,err1 := executeDRForSiteScript(startDRForSiteRequest.SiteName, startDRForSiteRequest.VmList)
    if err1 != nil {
        fmt.Println(err1.Error())
        retString = "\nError: executing DRForSite script failed...\n\n"
	c.IndentedJSON(http.StatusInternalServerError, retString)
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": resp})
    return
}

func attachPolicyToVMsForSite(c *gin.Context) {

    var retString string

    // Call BindJSON to bind the received JSON attachPolicyToVMsForSiteRequest structure
    if err := c.BindJSON(&attachPolicyToVMsForSiteRequest); err != nil {
        retString := "\nError: Checking attachPolicyToVMsForSite failed...\n\n"
	c.IndentedJSON(http.StatusInternalServerError, retString)
        return
    }

    os.Rename("demo_dr_infra_vm/CreateModifyVmList.json","demo_dr_infra_vm/CreateModifyVmListBackup.json")
    file, _ := json.MarshalIndent(attachPolicyToVMsForSiteRequest, "", " ")
    _ = ioutil.WriteFile("demo_dr_infra_vm/CreateModifyVmList.json", file, 0644)

    //_ = ioutil.WriteFile("dr_infra_tf/CreateModifyVmList.json", file, 0644)
    //_ = ioutil.WriteFile("dr_infra_tf/CreateModifyVmList.json", file, 0644)
    //_ = ioutil.WriteFile("dr_infra_tf/CreateVmList.json", file, 0644)

    resp,err1 := attachPolicyToVMsForSiteScript()
    if err1 != nil {
        fmt.Println(err1.Error())
        //retString = resp+"\nError: executing attachPolicy for failover Site script failed...\n\n"
        retString = "\nError: executing attachPolicy for failover Site script failed...\n\n"
	c.IndentedJSON(http.StatusInternalServerError, retString)
        return
    }
    c.JSON(http.StatusOK, gin.H{"response": resp})
    return
}

func executeDRForSiteScript(sitename string,vmlists []VM ) (string,error) {
    failover_progress ="Failover_Started"
    cmd, err := exec.Command("/bin/sh", "./demo_create_vms_using_terraform.sh").Output()
    if err != nil {
    fmt.Printf("error %s", err)
    failover_progress="Error occured while creating infrastructre "
    }
    output := string(cmd)
    failover_progress="Failover_Infra_Created"
    return output,err
}

func attachPolicyToVMsForSiteScript() (string,error) {
    failover_progress ="PrimaryIO_Policy_Attachment_Started"
    cmd, err := exec.Command("/bin/sh", "./demo_attach_policy_using_terraform.sh").Output()
    if err != nil {
    fmt.Printf("error %s", err)
    failover_progress="Error occured while attaching policy to VMs"
    }
    output := string(cmd)
    failover_progress="PrimaryIO_Policy_Attachment_Succeeded"
    return output,err
}

func exitErrorf(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}

// Create Log file and start logging 
func startLogging() {

    //Create directory for logging
    if err := os.MkdirAll("restord_logs", os.ModePerm); err != nil {
        log.Fatal(err)
    }
    
    // Logs will be stored at filename.
    filename := "restord_logs/restord_rest_server_"+time.Now().Format("2006-01-02-15-04-05")+".log"
    f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        exitErrorf("Unable to create log file %v", err)
    }

    fmt.Printf("\n Note: Logs will be stored at " + filename +" ...\n\n")
    gin.DefaultWriter = io.MultiWriter(f)
    // If you need to write the logs to file and console at the same time.
    //gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}

func GetLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}

func main() {

    // Force log's color
    gin.ForceConsoleColor()

    rest_server := gin.Default()


    //start Logging
    startLogging()

    os.Remove("demo_dr_infra_vm/vcenter_details.txt")
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

    // GET POST Requests
    rest_server.GET("/getDRStatusForSite", CORSMiddleware(), getDRStatusForSite) // get DR status for site
    rest_server.POST("/startDRForSite", CORSMiddleware(), startDRForSite)    // Start DR for site
    rest_server.POST("/attachPolicyToVMsForSite", CORSMiddleware(), attachPolicyToVMsForSite)    // Attach policy for VMs for site
    rest_server.GET("/getDRVcenterDetails",  CORSMiddleware(),getDRVcenterDetails) // get DR status for site

    //start Logging
    //startLogging()

    ipaddr = GetLocalIP()
    //fmt.Printf("\n Local ipaddress = %s \n",ipaddr)

    ipaddr_port_str := ipaddr+":8080"
    fmt.Printf("\n Starting RestoreD Rest Server on ip "+ipaddr+" at port 8080.\n")

    // Start the server 
    //rest_server.Run(ipaddr_port_str)
    rest_server.RunTLS(ipaddr_port_str,"./certs/restordserver.crt","./certs/restordserver.key")
}


func CORSMiddleware() gin.HandlerFunc {
	//Hack for preflight request. Need to find a better way
    return func(c *gin.Context) {
        // c.Writer.Header().Set("Access-Control-Allow-Origin", "https://192.168.1.10:4200")
        // c.Writer.Header().Set("Access-Control-Allow-Origin", "https://192.168.29.93:4200")
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        //c.Writer.Header().Set("Access-Control-Allow-Origin", "https://localhost:4200")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
		c.Next()
    }
}
