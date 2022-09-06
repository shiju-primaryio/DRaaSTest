 
Steps to configure gin REST Server:

  1. Unstall go and related packages as with ubuntu 20 go package is of 1.13 version and gin needs >= 1.16 version
	#apt-get remove golang-go
	#apt-get remove --auto-remove golang-go
  2. Install latest go packages
	# curl -OL https://go.dev/dl/go1.19.linux-amd64.tar.gz
	# tar -C /usr/local -xvf go1.19.linux-amd64.tar.gz
	# add following line in ~/.profile
		export PATH=$PATH:/usr/local/go/binA
	Check version
	# go version
		go version go1.19 linux/amd64
        Go Configuration
	# mkdir go_test
        # go mod init go_test
	# go mod tidy

  3. Install gin and IBM cloud related packages
	# go get -u github.com/gin-gonic/gin
	# go get github.com/IBM/ibm-cos-sdk-go/aws
	# go get github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam
        # go get github.com/IBM/ibm-cos-sdk-go/aws/session
        # go get github.com/IBM/ibm-cos-sdk-go/service/s3
        # go get "github.com/IBM/ibm-cos-sdk-go/service/s3/s3manager"

  4. Execute make
        #make
  5. Execute sal
        # sal_make & 	
     Note: SAL execution logs will be stored at Log file sal_logs/sal_rest_server_<timestamp> 
