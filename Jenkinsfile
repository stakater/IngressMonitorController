#!/usr/bin/groovy
@Library('github.com/stakater/fabric8-pipeline-library@fix-go-release')

def dummy

properties([
    disableConcurrentBuilds()
])

goBuildViaGoReleaser {
    chartRepositoryURL = 'https://chartmuseum.release.stakater.com'
    publicChartRepositoryURL = 'https://stakater.github.io/stakater-charts'
    publicChartGitURL = 'git@github.com:stakater/stakater-charts.git'
}
