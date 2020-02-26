#!/usr/bin/env groovy
@Library('github.com/stakater/stakater-pipeline-library@fix-pipeline-volumes') _

goBuildViaGoReleaser {
    publicChartRepositoryURL = 'https://stakater.github.io/stakater-charts'
    publicChartGitURL = 'git@github.com:stakater/stakater-charts.git'
    toolsImage = 'stakater/pipeline-tools:v2.0.13'
    dockerRepositoryURL = 'docker.pkg.github.com'
}