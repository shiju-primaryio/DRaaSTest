pipeline {
    agent {
        kubernetes {
            cloud 'my-k8s-cluster'
            label 'my-node'
            namespace 'my-namespace'
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
