#!groovy
@Library(['github.com/cloudogu/ces-build-lib@1.44.3'])
import com.cloudogu.ces.cesbuildlib.*

node('docker') {

  repositoryOwner = 'cloudogu'
  projectName = 'ces-confd'
  projectPath = "/go/src/github.com/${repositoryOwner}/${projectName}/"
  githubCredentialsId = 'sonarqube-gh'

  stage('Checkout') {
    checkout scm
  }

  docker.image('golang:1.14.13').inside("--volume ${WORKSPACE}:${projectPath} -e GOCACHE=/tmp/gocache") {
    stage('Build') {
      make 'clean package'
    }

    stage('Unit Test') {
      make 'unit-test'
      junit allowEmptyResults: true, testResults: 'target/unit-tests/*-tests.xml'
    }
  }

  stage('SonarQube') {
    String branch = "${env.BRANCH_NAME}"

    def scannerHome = tool name: 'sonar-scanner', type: 'hudson.plugins.sonar.SonarRunnerInstallation'
    withSonarQubeEnv {

      sh "git config 'remote.origin.fetch' '+refs/heads/*:refs/remotes/origin/*'"
      sh "git fetch --all"

      if (branch == "master") {
        echo "This branch has been detected as the master branch."
        sh "${scannerHome}/bin/sonar-scanner"
      } else if (branch == "develop") {
        echo "This branch has been detected as the develop branch."
        sh "${scannerHome}/bin/sonar-scanner -Dsonar.branch.name=${env.BRANCH_NAME} -Dsonar.branch.target=master"
      } else if (env.CHANGE_TARGET) {
        echo "This branch has been detected as a pull request."
        sh "${scannerHome}/bin/sonar-scanner -Dsonar.branch.name=${env.CHANGE_BRANCH}-PR${env.CHANGE_ID} -Dsonar.branch.target=${env.CHANGE_TARGET}"
      } else if (branch.startsWith("feature/") || branch.startsWith("bugfix/")) {
        echo "This branch has been detected as a feature branch."
        sh "${scannerHome}/bin/sonar-scanner -Dsonar.branch.name=${env.BRANCH_NAME} -Dsonar.branch.target=develop"
      } else {
        echo "WARNING: This branch has not been detected. Assuming a feature branch."
        sh "${scannerHome}/bin/sonar-scanner -Dsonar.branch.name=${env.BRANCH_NAME} -Dsonar.branch.target=develop"
      }
    }
    timeout(time: 2, unit: 'MINUTES') { // Needed when there is no webhook for example
      def qGate = waitForQualityGate()
      if (qGate.status != 'OK') {
        unstable("Pipeline unstable due to SonarQube quality gate failure")
      }
    }
  }
}

String projectPath

void make(goal) {
  sh "cd ${projectPath} && make ${goal}"
}
