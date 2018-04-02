#!/usr/bin/groovy
@Library('github.com/stakater/fabric8-pipeline-library@master')

def utils = new io.fabric8.Utils()

String chartPackageName = ""

toolsNode(toolsImage: 'stakater/pipeline-tools:1.5.1') {
    container(name: 'tools') {
        withCurrentRepo { def repoUrl, def repoName, def repoOwner, def repoBranch ->
            String srcDir = WORKSPACE + "/src"
            def kubernetesDir = WORKSPACE + "/kubernetes"

            def chartTemplatesDir = kubernetesDir + "/templates/chart"
            def chartDir = kubernetesDir + "/chart"
            def manifestsDir = kubernetesDir + "/manifests"
            
            def dockerImage = repoOwner.toLowerCase() + "/" + repoName.toLowerCase()
            def dockerImageVersion = ""

            // Slack variables
            def slackChannel = "${env.SLACK_CHANNEL}"
            def slackWebHookURL = "${env.SLACK_WEBHOOK_URL}"
            
            def git = new io.stakater.vc.Git()
            def helm = new io.stakater.charts.Helm()
            def common = new io.stakater.Common()
            def chartManager = new io.stakater.charts.ChartManager()
            def docker = new io.stakater.containers.Docker()
            def stakaterCommands = new io.stakater.StakaterCommands()
            def slack = new io.stakater.notifications.Slack()
            try {
                stage('Download Dependencies') {
                    sh """
                        cd ${srcDir}
                        glide update
                        cp -r ./vendor/* /go/src/
                    """
                }

                stage('Run Tests') {
                    sh """
                        cd ${srcDir}
                        go test
                    """
                }

                stage('Build Binary') {
                    sh """
                        cd ${srcDir}
                        go build -o ..out/${repoName.toLowerCase()}
                    """
                }

                if (utils.isCI()) {
                    stage('CI: Publish Dev Image') {
                        dockerImageVersion = stakaterCommands.getBranchedVersion("${env.BUILD_NUMBER}")
                        docker.buildImageWithTag(dockerImage, dockerImageVersion)
                        docker.pushTag(dockerImage, dockerImageVersion)
                    }
                } else if (utils.isCD()) {
                    stage('CD: Tag and Push') {
                        print "Generating New Version"
                        def version = common.shOutput("jx-release-version --gh-owner=${repoOwner} --gh-repository=${repoName}")
                        dockerImageVersion = version
                        sh """
                            echo "VERSION := ${version}" > Makefile
                        """

                        sh """
                            export VERSION=${version}
                            export DOCKER_IMAGE=${dockerImage}
                            gotplenv ${chartTemplatesDir}/Chart.yaml.tmpl > ${chartDir}/${repoName}/Chart.yaml
                            gotplenv ${chartTemplatesDir}/values.yaml.tmpl > ${chartDir}/${repoName}/values.yaml

                            helm template ${chartDir}/${repoName} -x templates/deployment.yaml > ${manifestsDir}/deployment.yaml
                            helm template ${chartDir}/${repoName} -x templates/configmap.yaml > ${manifestsDir}/configmap.yaml
                            helm template ${chartDir}/${repoName} -x templates/rbac.yaml > ${manifestsDir}/rbac.yaml
                        """
                        
                        git.commitChanges(WORKSPACE, "Bump Version to ${version}")

                        print "Pushing Tag ${version} to Git"
                        git.createTagAndPush(WORKSPACE, version)
                        git.createRelease(version)

                        print "Pushing Tag ${version} to DockerHub"
                        docker.buildImageWithTag(dockerImage, "latest")
                        docker.tagImage(dockerImage, "latest", version)
                        docker.pushTag(dockerImage, version)
                        docker.pushTag(dockerImage, "latest")
                    }
                    
                    stage('Chart: Init Helm') {
                        helm.init(true)
                    }

                    stage('Chart: Prepare') {
                        helm.lint(chartDir, repoName)
                        chartPackageName = helm.package(chartDir, repoName)
                    }

                    stage('Chart: Upload') {
                        String cmUsername = common.getEnvValue('CHARTMUSEUM_USERNAME')
                        String cmPassword = common.getEnvValue('CHARTMUSEUM_PASSWORD')
                        chartManager.uploadToChartMuseum(chartDir, repoName, chartPackageName, cmUsername, cmPassword)
                    }
                }
            }
            catch(e) {
                slack.sendDefaultFailureNotification(slackWebHookURL, slackChannel, [slack.createErrorField(e)])
            
                def commentMessage = "Yikes! You better fix it before anyone else finds out! [Build ${env.BUILD_NUMBER}](${env.BUILD_URL}) has Failed!"
                git.addCommentToPullRequest(commentMessage)

                throw e
            }
            stage('Notify') {
                def dockerImageWithTag = "${dockerImage}:${dockerImageVersion}"
                slack.sendDefaultSuccessNotification(slackWebHookURL, slackChannel, [slack.createDockerImageField(dockerImageWithTag)])

                def commentMessage = "Image is available for testing. `docker pull ${dockerImageWithTag}`"
                git.addCommentToPullRequest(commentMessage)
            }
        }
    }
}
