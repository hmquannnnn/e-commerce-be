pipeline {
    agent {
        docker {
            image 'golang:tip-alpine3.23'
            args '-u root'
        }
    }

    stages {
        stage('Prepare Environment') {
            steps {
                sh 'apk add --no-cache git'
                sh 'git config --global --add safe.directory "*"'
                script {
                    if (fileExists('prev_commit.txt')) {
                        env.PREV_BUILD_SUCCESS_COMMIT = readFile('prev_commit.txt').trim()
                        echo "PREV_BUILD_SUCCESS_COMMIT=${env.PREV_BUILD_SUCCESS_COMMIT}"
                    } else {
                        echo 'No prev_commit.txt (first build or workspace clean)'
                    }
                }
            }
        }

        stage('Detect services changed') {
            steps {
                script {
                    def prevBuildSuccessCommit = env.PREV_BUILD_SUCCESS_COMMIT
                    echo "PREV_BUILD_SUCCESS_COMMIT=${prevBuildSuccessCommit}"
                    if(prevBuildSuccessCommit == null) {
                        echo 'No prev_commit.txt (first build or workspace clean)'
                        env.BUILD_USER = 'true'
                        env.BUILD_FILE = 'true'
                        env.BUILD_API_GATEWAY = 'true'
                        env.BUILD_PRODUCT = 'true'
                        env.BUILD_ORDER = 'true'
                        env.BUILD_PAYMENT = 'true'

                        echo 'all services changed'
                        return
                    }
                    
                    def changedFiles = sh(
                        script: 'git diff --name-only ${prevBuildSuccessCommit} HEAD',
                        returnStdout: true
                    ).trim()

                    echo "Changed files:\n${changedFiles}"

                    def service = [
                        pkg        : 'pkg/',
                        apiGateway : 'api-gateway/',
                        user       : 'services/user',
                        file       : 'services/file',
                        product    : 'services/product',
                        order      : 'services/order',
                        payment    : 'services/payment'
                    ]

                    if (changedFiles.contains(service.pkg)) {
                        env.BUILD_USER = 'true'
                        env.BUILD_FILE = 'true'
                        env.BUILD_API_GATEWAY = 'true'
                        env.BUILD_PRODUCT = 'true'
                        env.BUILD_ORDER = 'true'
                        env.BUILD_PAYMENT = 'true'

                        echo 'all services changed'
                    } else if (changedFiles.contains(service.apiGateway)) {
                        env.BUILD_API_GATEWAY = 'true'
                        echo 'api-gateway changed'
                    } else if (changedFiles.contains(service.user)) {
                        env.BUILD_USER = 'true'
                        echo 'user service changed'
                    } else if (changedFiles.contains(service.file)) {
                        env.BUILD_FILE = 'true'
                        echo 'file service changed'
                    } else if (changedFiles.contains(service.product)) {
                        env.BUILD_PRODUCT = 'true'
                        echo 'product service changed'
                    } else if (changedFiles.contains(service.order)) {
                        env.BUILD_ORDER = 'true'
                        echo 'order service changed'
                    } else if (changedFiles.contains(service.payment)) {
                        env.BUILD_PAYMENT = 'true'
                        echo 'payment service changed'
                    }
                }
            }
        }
        
        stage('Build') {
            parallel {
                stage('Build API Gateway') {
                    when { expression { return env.BUILD_API_GATEWAY == 'true' } }
                    steps {
                        dir ('api-gateway') {
                            sh 'go mod download'
                            sh 'go build -o bin/api-gateway'
                        }
                    }
                }
                
                stage('Build User Service') {
                    when { expression { return env.BUILD_USER == 'true' } }
                    steps {
                        dir ('services/user') {
                            sh 'go mod download'
                            sh 'go build -o bin/user-service'
                        }
                    }
                }
                
                stage('Build File Service') {
                    when { expression { return env.BUILD_FILE == 'true' } }
                    steps {
                        dir ('services/file') {
                            sh 'go mod download'
                            sh 'go build -o bin/file-service'
                        }
                    }
                }

                stage('Build Product Service') {
                    when { expression { return env.BUILD_PRODUCT == 'true' } }
                    steps {
                        dir ('services/product') {
                            sh 'go mod download'
                            sh 'go build -o bin/product-service'
                        }
                    }
                }

                stage('Build Order Service') {
                    when { expression { return env.BUILD_ORDER == 'true' } }
                    steps {
                        dir ('services/order') {
                            sh 'go mod download'
                            sh 'go build -o bin/order-service'
                        }
                    }
                }

                stage('Build Payment Service') {
                    when { expression { return env.BUILD_PAYMENT == 'true' } }
                    steps {
                        dir ('services/payment') {
                            sh 'go mod download'
                            sh 'go build -o bin/payment-service'
                        }
                    }
                }
            }
        }
    }

    post {
        success {
            script {
                def commit = sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
                writeFile file: 'prev_commit.txt', text: commit
                echo "Saved PREV_BUILD_SUCCESS_COMMIT=${commit} to prev_commit.txt"
            }
        }
    }
}
