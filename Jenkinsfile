pipeline {
    agent any

    environment {
        DOCKERHUB_USER = 'hmquannnnn'
        GITHUB_REPO    = 'hmquannnnn/uav-store-be'
        GITOPS_REPO    = 'hmquannnnn/uav-store-infra'
        GITOPS_BRANCH  = 'dev'
        GITOPS_DIR     = 'gitops-repo'
    }

    stages {
        stage('Prepare Environment') {
            steps {
                sh 'git config --global --add safe.directory "*"'
                script {
                    env.GIT_SHA = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    echo "GIT_SHA=${env.GIT_SHA}"

                    if (fileExists('prev_commit.txt')) {
                        env.PREV_BUILD_SUCCESS_COMMIT = readFile('prev_commit.txt').trim()
                        echo "PREV_BUILD_SUCCESS_COMMIT=${env.PREV_BUILD_SUCCESS_COMMIT}"
                    } else {
                        echo 'No prev_commit.txt (first build or workspace clean)'
                    }
                }
            }
        }

        stage('Detect Services Changed') {
            steps {
                script {
                    def prev = env.PREV_BUILD_SUCCESS_COMMIT
                    if (prev == null) {
                        echo 'First build — marking all services for build'
                        env.BUILD_USER        = 'true'
                        env.BUILD_FILE        = 'true'
                        env.BUILD_API_GATEWAY = 'true'
                        env.BUILD_PRODUCT     = 'true'
                        env.BUILD_ORDER       = 'true'
                        env.BUILD_PAYMENT     = 'true'
                        return
                    }

                    def changedFiles = sh(
                        script: "git diff --name-only ${prev} HEAD",
                        returnStdout: true
                    ).trim()

                    echo "Changed files:\n${changedFiles}"

                    def paths = [
                        pkg        : 'pkg/',
                        apiGateway : 'api-gateway/',
                        user       : 'services/user',
                        file       : 'services/file',
                        product    : 'services/product',
                        order      : 'services/order',
                        payment    : 'services/payment',
                    ]

                    if (changedFiles.contains(paths.pkg)) {
                        env.BUILD_USER        = 'true'
                        env.BUILD_FILE        = 'true'
                        env.BUILD_API_GATEWAY = 'true'
                        env.BUILD_PRODUCT     = 'true'
                        env.BUILD_ORDER       = 'true'
                        env.BUILD_PAYMENT     = 'true'
                        echo 'pkg/ changed — rebuilding all services'
                    } else {
                        if (changedFiles.contains(paths.apiGateway)) { env.BUILD_API_GATEWAY = 'true'; echo 'api-gateway changed' }
                        if (changedFiles.contains(paths.user))       { env.BUILD_USER        = 'true'; echo 'user-service changed' }
                        if (changedFiles.contains(paths.file))       { env.BUILD_FILE        = 'true'; echo 'file-service changed' }
                        if (changedFiles.contains(paths.product))    { env.BUILD_PRODUCT     = 'true'; echo 'product-service changed' }
                        if (changedFiles.contains(paths.order))      { env.BUILD_ORDER       = 'true'; echo 'order-service changed' }
                        if (changedFiles.contains(paths.payment))    { env.BUILD_PAYMENT     = 'true'; echo 'payment-service changed' }
                    }
                }
            }
        }

        stage('Docker Build & Push') {
            agent {
                docker {
                    image 'docker:25-cli'
                    args  "--entrypoint='' -v /var/run/docker.sock:/var/run/docker.sock -u root"
                    reuseNode true
                }
            }
            steps {
                script {
                    def sha  = env.GIT_SHA
                    def user = env.DOCKERHUB_USER

                    def services = [
                        [build: env.BUILD_API_GATEWAY == 'true', name: 'api-gateway',     dockerfile: 'api-gateway/Dockerfile',      context: '.'],
                        [build: env.BUILD_USER == 'true',        name: 'user-service',    dockerfile: 'services/user/Dockerfile',    context: '.'],
                        [build: env.BUILD_FILE == 'true',     name: 'file-service',    dockerfile: 'services/file/Dockerfile',    context: '.'],
                        [build: env.BUILD_PRODUCT == 'true',     name: 'product-service', dockerfile: 'services/product/Dockerfile', context: '.'],
                        [build: env.BUILD_ORDER == 'true',      name: 'order-service',   dockerfile: 'services/order/Dockerfile',   context: '.'],
                        [build: env.BUILD_PAYMENT == 'true',     name: 'payment-service', dockerfile: 'services/payment/Dockerfile', context: '.'],
                    ]

                    withCredentials([usernamePassword(
                        credentialsId: 'dockerhub-credentials',
                        usernameVariable: 'DOCKER_USER',
                        passwordVariable: 'DOCKER_PASS'
                    )]) {
                        sh 'echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin'

                        def anyBuilt = false
                        services.each { svc ->
                            if (svc.build) {
                                anyBuilt = true

                                echo "===== Building ${svc.name} ====="
                                def img = "${user}/${svc.name}"
                                sh "docker pull ${img}:latest || true"
                                sh "docker build --cache-from ${img}:latest -f ${svc.dockerfile} -t ${img}:latest -t ${img}:${sha} ${svc.context}"
                                sh "docker push ${img}:latest"
                                sh "docker push ${img}:${sha}"
                                echo "Pushed ${img}:latest and ${img}:${sha}"
                            }
                        }

                        if (!anyBuilt) {
                            echo 'No services changed — skipping Docker build'
                        }

                        sh 'docker logout'
                    }
                }
            }
        }

        stage('Update K8s Manifests') {
            steps {
                script {
                    def sha  = env.GIT_SHA
                    def user = env.DOCKERHUB_USER
                    def gitopsDir = env.GITOPS_DIR
                    def branch = env.GITOPS_BRANCH
                    def gitopsRepo = env.GITOPS_REPO

                    def serviceMappings = [
                        [build: env.BUILD_API_GATEWAY == 'true', name: 'api-gateway',     yaml: 'deployment/k8s/07-api-gateway.yaml'],
                        [build: env.BUILD_USER == 'true',        name: 'user-service',    yaml: 'deployment/k8s/06-user-service.yaml'],
                        [build: env.BUILD_FILE == 'true',        name: 'file-service',    yaml: 'deployment/k8s/10-file-service.yaml'],
                        [build: env.BUILD_PRODUCT == 'true',     name: 'product-service', yaml: 'deployment/k8s/11-product-service.yaml'],
                        [build: env.BUILD_ORDER == 'true',       name: 'order-service',   yaml: 'deployment/k8s/14-order-service.yaml'],
                        [build: env.BUILD_PAYMENT == 'true',     name: 'payment-service', yaml: 'deployment/k8s/17-payment-service.yaml'],
                    ]

                    def updated = false
                    serviceMappings.each { svc ->
                        if (svc.build) {
                            updated = true
                        }
                    }

                    if (updated) {
                        withCredentials([usernamePassword(
                            credentialsId: 'github-credentials',
                            usernameVariable: 'GIT_USER',
                            passwordVariable: 'GIT_TOKEN'
                        )]) {
                            dir(gitopsDir) {
                                deleteDir()
                            }

                            sh "git clone --branch ${branch} https://\${GIT_USER}:\${GIT_TOKEN}@github.com/${gitopsRepo}.git ${gitopsDir}"

                            serviceMappings.each { svc ->
                                if (svc.build) {
                                    sh "sed -i 's|image: .*${svc.name}.*|image: ${user}/${svc.name}:${sha}|g' ${gitopsDir}/${svc.yaml}"
                                }
                            }

                            sh """
                                cd ${gitopsDir}
                                git config user.email "jenkins@uav-store"
                                git config user.name "Jenkins"

                                git add deployment/k8s/
                                git diff --staged --quiet || git commit -m "ci: update image tags to ${sha} [ci skip]"

                                git pull --rebase origin ${branch}

                                git push origin HEAD:${branch}
                            """
                        }
                        echo "K8s manifests updated to ${sha}"
                    } else {
                        echo 'No manifests to update'
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
                echo "Saved prev_commit=${commit}"
            }
        }
        failure {
            echo 'Build failed'
        }
    }
}
