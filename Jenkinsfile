pipeline {
    agent {
        kubernetes {
            cloud 'my-k8s-cluster'
        }
    }
    stages {
        stage('Get Pods') {
            steps {
                sh 'kubectl get pods'
            }
        }
    }
}
