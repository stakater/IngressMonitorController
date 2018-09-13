#!/usr/bin/groovy
@Library('github.com/stakater/fabric8-pipeline-library@remove-Nested-Openshift-Vendor')

def dummy

properties([
    disableConcurrentBuilds()
])

goBuildAndRelease {
    removeNestedOpenshiftVendor = true
}
