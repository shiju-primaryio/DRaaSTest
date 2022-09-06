echo "Testing List of Protected VMs.."
curl http://localhost:8080/listProtectedVms
echo "Testing Create Bucket.."
curl -X POST -H "Content-Type: application/json" -d '{"Name": "rahulk-test9"}' http://localhost:8080/createVMBucket
curl http://localhost:8080/listProtectedVms
echo "Testing Delete Bucket.."
curl -X POST -H "Content-Type: application/json" -d '{"Name": "rahulk-test9"}' http://localhost:8080/deleteVMBucket
curl http://localhost:8080/listProtectedVms

#Issues with naming 
curl -X POST -H "Content-Type: application/json" -d '{"Name": "RKTPTesting"}' http://localhost:8080/createVMBucket


