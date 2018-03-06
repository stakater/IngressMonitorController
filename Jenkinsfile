#!/usr/bin/groovy
@Library('github.com/stakater/fabric8-pipeline-library@master')

def utils = new io.fabric8.Utils()

clientsNode(clientsImage: 'stakater/pipeline-tools:1.1') {
    container(name: 'clients') {
        String workspaceDir = WORKSPACE
        stage('Checkout') {
            checkout scm
        }

        stage('Download Dependencies') {
            sh """
                cd ${workspaceDir}
                glide update
                cp -r ./vendor/* /src/go/
            """
        }

        if (utils.isCI()) {
            stage('CI: Test') {
                sh """
                    cd ${workspaceDir}
                    go test
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