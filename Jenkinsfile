pipeline {
    agent { docker { image 'golang:tip-alpine3.23' } }
    stages {
        stage('build') {
            steps {
                sh 'go version'
            }
        }
    }
}