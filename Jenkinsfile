#!/usr/bin/groovy
@Library('github.com/stakater/fabric8-pipeline-library@remove-ssh-node')

def utils = new io.fabric8.Utils()
String gitUsername = "stakater-user"
String gitEmail = "stakater@gmail.com"

String chartPackageName = ""
String chartName = "chart/ingress-monitor-controller"

toolsNode(toolsImage: 'stakater/pipeline-tools:1.4.0') {
    container(name: 'tools') {
        withCurrentRepo { def repoUrl, def repoName, def repoBranch ->
            String workspaceDir = WORKSPACE + "/src"
            def git = new io.stakater.vc.Git()
            def helm = new io.stakater.charts.Helm()
            def common = new io.stakater.Common()
            def chartManager = new io.stakater.charts.ChartManager()

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
                    print "Generating New Version"
                    sh """
                        cd ${WORKSPACE}
                        VERSION=\$(jx-release-version)
                        echo "VERSION := \${VERSION}" > Makefile
                    """

                    def version = new io.stakater.Common().shOutput """
                        cd ${WORKSPACE}
                        
                        chmod 600 /root/.ssh-git/ssh-key > /dev/null
                        eval `ssh-agent -s` > /dev/null
                        ssh-add /root/.ssh-git/ssh-key > /dev/null

                        jx-release-version
                    """
                    
                    git.commitChanges(workspaceDir, "Bump Version")

                    print "Pushing Tag ${version} to Git"
                    sh """
                        cd ${WORKSPACE}
                        
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
                
                stage('Chart: Init Helm') {
                    helm.init(true)
                }

                stage('Chart: Prepare') {
                    helm.lint(WORKSPACE, chartName)
                    chartPackageName = helm.package(WORKSPACE, chartName)
                }

                stage('Chart: Upload') {
                    String cmUsername = common.getEnvValue('CHARTMUSEUM_USERNAME')
                    String cmPassword = common.getEnvValue('CHARTMUSEUM_PASSWORD')
                    chartManager.uploadToChartMuseum(WORKSPACE, chartName, chartPackageName, cmUsername, cmPassword)
                }
            }
        }
    }
}
