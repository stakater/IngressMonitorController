#!/usr/bin/groovy
@Library('github.com/stakater/fabric8-pipeline-library@master')

def utils = new io.fabric8.Utils()
String gitUsername = "stakater-user"
String gitEmail = "stakater@gmail.com"

String thisRepo = "git@github.com:stakater/IngressMonitorController"
String thisRepoBranch = "master"
String thisRepoDir = "IngressMonitorController"

controllerNode(clientsImage: 'stakater/pipeline-tools:1.2.0') {
    container(name: 'clients') {
        String workspaceDir = WORKSPACE + "/src"
        def git = new io.stakater.vc.Git()
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
                    go build -o ../out/ingressmonitorcontroller
                    cd ..
                    docker build -t docker.io/stakater/ingress-monitor-controller:dev .
                    docker push docker.io/stakater/ingress-monitor-controller:dev
                """
            }
        } else if (utils.isCD()) {
            stage('CD: Build') {
                sh """
                    cd ${workspaceDir}
                    go test
                    go build -o ../out/ingressmonitorcontroller
                """
            }

            stage('CD: Tag and Push') {
                print "Checkout current Repo for pushing version"

                git.setUserInfo(gitUsername, gitEmail)
                git.addHostsToKnownHosts()
                git.checkoutRepo(thisRepo, thisRepoBranch, thisRepoDir)

                print "Generating New Version"
                sh """
                    cd ${WORKSPACE}/${thisRepoDir}
                    VERSION=\$(jx-release-version)
                    echo "VERSION := \${VERSION}" > Makefile
		        """

                def version = new io.stakater.Common().shOutput("cd ${WORKSPACE}/${thisRepoDir}; jx-release-version")
                
                git.commitChanges(thisRepoDir, "Bump Version")

                sh """
                    cd ${WORKSPACE}/${thisRepoDir}
                    git tag -a ${version}
                    git push --tags
                """

                sh """
                    docker build -t docker.io/stakater/ingress-monitor-controller:${version} .
                    docker tag docker.io/stakater/ingress-monitor-controller:${version} docker.io/stakater/ingress-monitor-controller:latest
                    docker push docker.io/stakater/ingress-monitor-controller:${version}
	                docker push docker.io/stakater/ingress-monitor-controller:latest
                """
            }
        }
    }
}
