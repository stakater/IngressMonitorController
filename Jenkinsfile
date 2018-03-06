#!/usr/bin/groovy
@Library('github.com/stakater/fabric8-pipeline-library@add-controller-template')

def utils = new io.fabric8.Utils()

controllerNode(clientsImage: 'stakater/pipeline-tools:1.1') {
    container(name: 'clients') {
        String workspaceDir = WORKSPACE + "/src"
        stage('Checkout') {
            checkout scm
        }

        stage('Download Dependencies') {
            sh """
                cd ${workspaceDir}
                glide update
                cp -r ./vendor/* /go/src/
            """
        }

        if (utils.isCI()) {
            stage('CI: Test') {
                sh """
                    cd ${workspaceDir}
                    go test
                """
            }
            stage('CI: Publish Dev Image') {
                sh """
                    cd ${workspaceDir}
                    go build -o /out/ingressmonitorcontroller
                    cd ..
                    docker build -t docker.io/stakater/ingress-monitor-controller:dev .
                    docker push docker.io/stakater/ingress-monitor-controller:latest
                """
            }
        } else if (utils.isCD()) {
            stage('CD: Build') {
                sh """
                    cd ${workspaceDir}
                    go test
                    go build -o /out/ingressmonitorcontroller
                """
            }

            stage('CD: Tag and Push') {
                sh """
                    cd ${workspaceDir}
                    cd ..
                    VERSION=\$(jx-release-version)
                    echo "VERSION := \${VERSION}" > Makefile
                    
                    git add Makefile
                    git commit -m 'release \${VERSION}'
                    git push

                    docker build -t docker.io/stakater/ingress-monitor-controller:\${VERSION} .
                    docker tag docker.io/stakater/ingress-monitor-controller:\${VERSION} docker.io/stakater/ingress-monitor-controller:latest
                    docker push docker.io/stakater/ingress-monitor-controller:\${VERSION}
	                docker push docker.io/stakater/ingress-monitor-controller:latest
                """
            }
        }
    }
}