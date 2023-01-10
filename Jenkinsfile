pipeline {
    agent {
        kubernetes {
            cloud 'my-k8s-cluster'
        }
    }
    node {
        stage('List pods') {
            withKubeConfig([credentialsId: '<credential-id>',
                        caCertificate:  '''-----BEGIN CERTIFICATE-----
MIIC5zCCAc+gAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl
cm5ldGVzMB4XDTIzMDExMDIyMzAwNloXDTMzMDEwNzIyMzAwNlowFTETMBEGA1UE
AxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALNb
UmTTZ1/KnQ71IclcSzAuSuPSQhutwtbZ76EyVsymrTyDsKkHvEZz1L0RlO0kBJ/Z
pXlT/JHR3eBudGQMRUoCEzbN8ddUzO0HxHvtNCuAB46f7d2O2CeyH6n09Bdh7R/q
+57pZfxernAmBdcw0f1I08MjGF5ppfMC7JwmPRuobq99w0jy1krwGgZOh5zFRkzy
tJ5Rvc1LKSagJ892q7SGv2zDCoKle0oCxwZqy190WHUKHp7tZcj+dET4XRQt0IdE
F9jdy8vjyzXzH6yGvoNnv1TsxjOQq9CBIxVCoEJyxxM7ysrVHlUHndHdiRoT0arZ
voLSu5My/ljeT341XisCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB
/wQFMAMBAf8wHQYDVR0OBBYEFCy25z1PP8BoAMggC2u28vdkuHJPMA0GCSqGSIb3
DQEBCwUAA4IBAQBQSnMajv4S3CVJOqNJlsQWMaVJBzwDB4qVTy/o7zXpGfmSm908
z9EHRJmBPbCuBq4yOPAJ2AQuEjt9IhodCBHYzOhS04zFtreWk+/YOLiD/y9M0Ca6
HkLhapTM4rexyz5YHpZTNnfhnpn3urLCo+MsV+u0blD6w7/rVrylcXHcF3pYhd50
0toZtZjjkG1jxmC06pk9CXu0ZIxpUm95XOYOuQ7GwN0ryXKtuQWpjpMT/sbbIUoG
wRn/+0e2/qi+V0PqjhJzrD3LuKAs31uiPxNFpGl6ILBxEGkuEqQiBZBRNBXcbL/R
sA+sLRd/zR9m5it301+cc3rGiV2hX0ibxFnn
-----END CERTIFICATE-----''',
                        serverUrl: '192.168.0.8:6443',
                        contextName: '',
                        clusterName: 'kubernetes-admin@kubernetes',
                        namespace: 'default'
                        ]) {
                sh 'kubectl get pods'
                        }
        }
    }
}
