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
                // We need to checkout again because we can't commit and push changes to the repo that is checkout via scm
                git.checkoutRepo(thisRepo, thisRepoBranch, thisRepoDir)

                git.addHostsToKnownHosts()
                print "Generating New Version"
                sh """
                    cd ${WORKSPACE}/${thisRepoDir}
                    
                    chmod 600 /root/.ssh-git/ssh-key
                    eval `ssh-agent -s`
                    ssh-add /root/.ssh-git/ssh-key
                    
                    VERSION=\$(jx-release-version)
                    echo "VERSION := \${VERSION}" > Makefile
		        """

                def version = new io.stakater.Common().shOutput """
                    cd ${WORKSPACE}/${thisRepoDir}
                    
                    chmod 600 /root/.ssh-git/ssh-key > /dev/null
                    eval `ssh-agent -s` > /dev/null
                    ssh-add /root/.ssh-git/ssh-key > /dev/null

                    jx-release-version
                """
                
                git.commitChanges(thisRepoDir, "Bump Version")

                print "Pushing Tag ${version} to Git"
                sh """
                    cd ${WORKSPACE}/${thisRepoDir}
                    
                    chmod 600 /root/.ssh-git/ssh-key
                    eval `ssh-agent -s`
                    ssh-add /root/.ssh-git/ssh-key
                    
                    git tag ${version}
                    git push --tags
                """

                print "Pushing Tag ${version} to DockerHub"
                sh """
                    cd ${WORKSPACE}
                    docker build -t docker.io/stakater/ingress-monitor-controller:${version} .
                    docker tag docker.io/stakater/ingress-monitor-controller:${version} docker.io/stakater/ingress-monitor-controller:latest
                    docker push docker.io/stakater/ingress-monitor-controller:${version}
	                docker push docker.io/stakater/ingress-monitor-controller:latest
                """
            }
        }
    }
}
