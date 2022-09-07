echo "Testing List of Protected VMs.."
curl http://localhost:8080/listProtectedVms
echo "Testing Create Bucket.."
curl -X POST -H "Content-Type: application/json" -d '{"Name": "rahulk-test9"}' http://localhost:8080/createVMBucket
curl http://localhost:8080/listProtectedVms
echo "Testing Delete Bucket.."
curl -X POST -H "Content-Type: application/json" -d '{"Name": "rahulk-test9"}' http://localhost:8080/deleteVMBucket
curl http://localhost:8080/listProtectedVms

#Test Read and Write
curl -X POST -H "Content-Type: application/json" -d '{"bucketname": "rahulk3-test3", "objkey": "test_data1","data":"Hello World!!" }' http://localhost:8080/addVaioObj
curl -X GET   -H "Content-type: application/json"   -H "Accept: application/json"   -d '{"bucketname": "rahulk3-test3", "objkey": "test_data1" }' http://localhost:8080/getVaioObj

#Issues with naming 
curl -X POST -H "Content-Type: application/json" -d '{"Name": "RKTPTesting"}' http://localhost:8080/createVMBucket


