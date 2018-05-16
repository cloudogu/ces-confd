#!groovy
@Library('github.com/cloudogu/ces-build-lib@8dd2371')
import com.cloudogu.ces.cesbuildlib.*

node('docker') {

  repositoryOwner = 'cloudogu'
  repositoryName = 'ces-confd'
  project = "github.com/${repositoryOwner}/${repositoryName}"
  githubCredentialsId = 'sonarqube-gh'

  stage('Checkout') {
    checkout scm
  }


  new Docker(this).image('cloudogu/golang:latest')
    .mountJenkinsUser()
    .mountDockerSocket()
    .installDockerClient('18.03.1')
    .inside("--volume ${WORKSPACE}:/go/src/${project} -e USER=jenkins") {

    stage('Build') {
      make 'clean'
      make 'build'
      archiveArtifacts 'target/**/*.tar.gz'
    }

    stage('Unit Test') {
      make 'unit-test'
      junit allowEmptyResults: true, testResults: 'target/*-tests.xml'
    }

    stage('Static Analysis') {
      def commitSha = sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
      withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: githubCredentialsId, usernameVariable: 'USERNAME', passwordVariable: 'REVIEWDOG_GITHUB_API_TOKEN']]) {
        withEnv(["CI_PULL_REQUEST=${env.CHANGE_ID}", "CI_COMMIT=${commitSha}", "CI_REPO_OWNER=${repositoryOwner}", "CI_REPO_NAME=${repositoryName}"]) {
          make 'static-analysis'
        }
      }
    }
  }

}

String repositoryOwner
String repositoryName
String project
String githubCredentialsId

void make(goal) {
  sh "cd /go/src/${project} && make ${goal}"
}
