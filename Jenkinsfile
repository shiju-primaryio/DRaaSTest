pipeline {
  agent any

  credentials {
    kubernetes {
      configName = 'my-cluster'
      serviceAccount = '/path/to/service-account.json'
    }
  }

  stages {
    stage('Connect to cluster') {
      steps {
        script {
          def cluster = kubernetes.cluster('my-cluster')
          kubernetes.configureConnection(cluster)
        }
      }
    }
    stage('List pods') {
      steps {
        sh 'kubectl get pods'
      }
    }
  }
}