# Release process

To release a new version do the following:

- merge your changes via a github PR into develop
- on develop, exec: `make go-release`
- follow the steps of gitflow
- wait for the release branch build on jenkins
- download binary and checksum
- add binary and checksum to new github release
