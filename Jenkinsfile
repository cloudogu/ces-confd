#!groovy
@Library(['github.com/cloudogu/ces-build-lib@1.56.0'])
import com.cloudogu.ces.cesbuildlib.*

node('docker') {
  Git git = new Git(this, "cesmarvin")
  git.committerName = 'cesmarvin'
  git.committerEmail = 'cesmarvin@cloudogu.com'
  GitFlow gitflow = new GitFlow(this, git)
  GitHub github = new GitHub(this, git)
  Changelog changelog = new Changelog(this)
  Docker docker = new Docker(this)
  Gpg gpg = new Gpg(this, docker)

  repositoryOwner = 'cloudogu'
  projectName = 'ces-confd'
  projectPath = "/go/src/github.com/${repositoryOwner}/${projectName}/"
  githubCredentialsId = 'sonarqube-gh'

  stage('Checkout') {
    checkout scm
  }

  docker.image('golang:1.17.8').inside("--volume ${WORKSPACE}:${projectPath} -e GOCACHE=/tmp/gocache") {
    stage('Build') {
      make 'clean package checksum'
      archiveArtifacts 'target/**/*.tar.gz'
      archiveArtifacts 'target/**/*.sha256sum'
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

  if (gitflow.isReleaseBranch()) {
    String releaseVersion = git.getSimpleBranchName();

    stage('Finish Release') {
      gitflow.finishRelease(releaseVersion)
    }

    stage('Build after Release') {
      git.checkout(releaseVersion)
      make 'clean package checksum'
    }

    stage('Sign after Release'){
      gpg.createSignature()
    }

    stage('Add Github-Release') {
      github.createReleaseWithChangelog(releaseVersion, changelog)
      github.addReleaseAsset("${releaseId}", "target/ces-confd-*.tar.gz")
      github.addReleaseAsset("${releaseId}", "target/ces-confd.sha256sum")
      github.addReleaseAsset("${releaseId}", "target/ces-confd.sha256sum.asc")
    }
  }

}

String projectPath

void make(goal) {
  sh "cd ${projectPath} && make ${goal}"
}
