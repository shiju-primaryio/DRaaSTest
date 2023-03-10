package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/CacheboxInc/DRaaS/src/syncd/vmOps"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// @title SyncD OpenAPI
// @version 1.0
// @description This is syncd server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /
// @schemes http
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	env := os.Getenv("ENVIRONMENT")
	env = strings.ToLower(env)
	if env == "production" || env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.POST("/fullDiskSync", FullDiskSyncHandler)

	host := os.Getenv("HOST")
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	log.Printf("Server listening on http://%s:%d/", host, port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), router); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}

// FullDiskSyncHandler godoc
// @Summary Perform a full sync of VM's virtual disk.
// @Description Perform a full sync of VM's virtual disk.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /fullDiskSync [post]
func FullDiskSyncHandler(ctx *gin.Context) {
	var vmInfo FullDiskSyncHandlerRequest
	ctx.BindJSON(&vmInfo)
	fmt.Println("VM UUID: ", vmInfo.VmInstanceUuid)
	client, err := vmOps.GetVCenterClient(ctx)
	if err != nil {
		fmt.Printf("Unable to obtain API client to vCenter.")
		ctx.JSON(http.StatusInternalServerError, FullDiskSyncHandlerResponse{Message: "Failed to connect to vCenter"})
		return
	}
	vmObj, err := vmOps.FindVmByInstanceUuid(client, vmInfo.VmInstanceUuid)
	if err != nil {
		fmt.Printf("No VM found: %v\n", err)
		ctx.JSON(http.StatusNotFound, FullDiskSyncHandlerResponse{Message: "VM not found"})
		return
	}

	temp_uuid := uuid.New()
	clone_vm_name := "clone-" + vmObj.Summary.Config.Name + "-" + temp_uuid.String()
	err = vmOps.CloneVm(client, vmObj, clone_vm_name)
	if err != nil {
		fmt.Printf("VM Clone failed: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, FullDiskSyncHandlerResponse{Message: "VM clone failed"})
		return
	}
	ctx.JSON(http.StatusOK, FullDiskSyncHandlerResponse{Message: "Success"})
}

type FullDiskSyncHandlerRequest struct {
	VmInstanceUuid string `json:"vmInstanceUuid" binding:"required"`
}

type FullDiskSyncHandlerResponse struct {
	Message string `json:"message" binding:"required"`
}
