# Changelog

## [1.4.0](https://github.com/air-gapped/cooked/compare/v1.3.2...v1.4.0) (2026-04-03)


### Features

* add --frame-ancestors flag for iframe embedding ([8c9536f](https://github.com/air-gapped/cooked/commit/8c9536f02c030a2b12f5b4d219755d3bcf874b37))
* add --trusted-proxies for X-Forwarded-For client IP logging ([e50437a](https://github.com/air-gapped/cooked/commit/e50437ac2386b7d2ea7b43d7f51fa6ca8ce187f1))
* add Helm chart for Kubernetes deployment ([eae7614](https://github.com/air-gapped/cooked/commit/eae761488ecd6c262ea4e6a2108d501972819da6))
* proxy images and assets through /_cooked/raw/ to fix CORS ([0a7c38e](https://github.com/air-gapped/cooked/commit/0a7c38ec9b982d1df5a6390b00ec8a4b8e804d73))

## [1.3.2](https://github.com/air-gapped/cooked/compare/v1.3.1...v1.3.2) (2026-04-03)


### Bug Fixes

* sign Docker images with cosign and attach SBOM attestations ([cfac0c3](https://github.com/air-gapped/cooked/commit/cfac0c378e7a230be7590493d1a8d2d41b65648f))

## [1.3.1](https://github.com/air-gapped/cooked/compare/v1.3.0...v1.3.1) (2026-03-31)


### Bug Fixes

* replace coverage theater TestSetup_JSONOutput with real Setup() test ([e57b850](https://github.com/air-gapped/cooked/commit/e57b850965615c8c06388e91c70087d35df3dba3))

## [1.3.0](https://github.com/air-gapped/cooked/compare/v1.2.3...v1.3.0) (2026-03-31)


### Features

* add dependency review workflow with CVE check and AI security analysis ([a5df1fd](https://github.com/air-gapped/cooked/commit/a5df1fd281fb125793124fc7310edf3b09516f20))
* run update-ca-certificates at container startup for runtime CA injection ([0b460bb](https://github.com/air-gapped/cooked/commit/0b460bb174dc4d75190f25a539db0eea7fd4c9e2))

## [1.2.3](https://github.com/air-gapped/cooked/compare/v1.2.2...v1.2.3) (2026-03-26)


### Bug Fixes

* **deps:** update module github.com/yuin/goldmark to v1.8.2 ([#48](https://github.com/air-gapped/cooked/issues/48)) ([aeaeb27](https://github.com/air-gapped/cooked/commit/aeaeb27f20f654a5518d4ed1700aa7aaf5278023))

## [1.2.2](https://github.com/air-gapped/cooked/compare/v1.2.1...v1.2.2) (2026-03-26)


### Bug Fixes

* **deps:** update module github.com/sirupsen/logrus to v1.9.4 ([#39](https://github.com/air-gapped/cooked/issues/39)) ([2039de7](https://github.com/air-gapped/cooked/commit/2039de728738edb37ce437dc5940300ca72af945))
* **deps:** update module github.com/yuin/goldmark to v1.7.17 ([#37](https://github.com/air-gapped/cooked/issues/37)) ([3483394](https://github.com/air-gapped/cooked/commit/34833946981f477037a39bb893b1744ffa6e74d0))
* **deps:** update module github.com/yuin/goldmark to v1.8.1 ([#40](https://github.com/air-gapped/cooked/issues/40)) ([8a81874](https://github.com/air-gapped/cooked/commit/8a8187421f0df35d99089f80a98092012df72a8a))

## [1.2.1](https://github.com/air-gapped/cooked/compare/v1.2.0...v1.2.1) (2026-02-10)


### Bug Fixes

* resolve hostnames against CIDR allowlist, proxy raw content for copy button, silence .well-known noise ([e22c596](https://github.com/air-gapped/cooked/commit/e22c5962b5e0b359fe534b7b477a8498373c66b9))

## [1.2.0](https://github.com/air-gapped/cooked/compare/v1.1.0...v1.2.0) (2026-02-08)


### Features

* add AsciiDoc and Org-mode rendering support ([3d63832](https://github.com/air-gapped/cooked/commit/3d63832e2fe8b7cd5c698e2eca79c99ac9d9ddac))
* add automated releases with release-please ([7066965](https://github.com/air-gapped/cooked/commit/706696543042f32b66368dad8b55178b65b4ed54))


### Bug Fixes

* add container to all release workflow jobs ([248f377](https://github.com/air-gapped/cooked/commit/248f377e65e4cdf0ed9e57fc9434e94b2de73f81))
* add github.com host to kaniko git context URL ([5802f3d](https://github.com/air-gapped/cooked/commit/5802f3daa61eaa517a4a9100fd17fabae3241dee))
* align first line of code blocks with subsequent lines ([d62eee4](https://github.com/air-gapped/cooked/commit/d62eee43667b83920359fd44a81145b71c9aaa07))
* eliminate two-character offset on first line ([6bd172d](https://github.com/air-gapped/cooked/commit/6bd172d0cd7238ca58d3447c5e1a0b75498deccc))
* eliminate two-character offset on first line by disabling chroma line number padding ([c0231a6](https://github.com/air-gapped/cooked/commit/c0231a63598c13b37ec2a3f628a3d90defea1d22)), closes [#11](https://github.com/air-gapped/cooked/issues/11)
* remove left padding from line numbers to eliminate two-character offset ([66763c1](https://github.com/air-gapped/cooked/commit/66763c1bd87e479cbe1e7256795fe9c1fc8981d2))
* remove left padding from line numbers to eliminate two-character offset ([4e212e5](https://github.com/air-gapped/cooked/commit/4e212e58e328de6830c7453e83fc38f57b46e1dc)), closes [#11](https://github.com/air-gapped/cooked/issues/11)
* remove non-existent WithLineNumberPad chroma option ([c09d01c](https://github.com/air-gapped/cooked/commit/c09d01c1df9caf7c382cc3bd7e2bcd4a8975bb92))
* silence logrus output from libasciidoc ([f03fe85](https://github.com/air-gapped/cooked/commit/f03fe85586fdac0742c0ce0a09658c7d6e758233))
* strip leading space-padding from chroma line numbers ([0bf1f05](https://github.com/air-gapped/cooked/commit/0bf1f05377332e4a951635018708ff59f54fb7d2))
* use crane for Docker image builds instead of BuildKit ([575f8e6](https://github.com/air-gapped/cooked/commit/575f8e6b826068ef8a59db80de0d2545951330bd))
* use full ghcr.io registry path for Docker images in README ([453c12c](https://github.com/air-gapped/cooked/commit/453c12c7be3187484e99ef2b665cce38e558faed))
* use git:// protocol for kaniko context URL ([679bc4e](https://github.com/air-gapped/cooked/commit/679bc4e64d6d91b1c44583913b3bd653d765b095))
* use GitHub App token for release-please PR creation ([c85c4b1](https://github.com/air-gapped/cooked/commit/c85c4b1002653e60c2eee7f45e4910eda5ad18e7))
* use kaniko for Docker image builds on ARC runners ([e2f5fbf](https://github.com/air-gapped/cooked/commit/e2f5fbf1be24537e966c5e04d6e84114234bce11))

## [1.1.0](https://github.com/air-gapped/cooked/compare/v1.0.7...v1.1.0) (2026-02-08)


### Features

* add AsciiDoc and Org-mode rendering support ([3d63832](https://github.com/air-gapped/cooked/commit/3d63832e2fe8b7cd5c698e2eca79c99ac9d9ddac))

## [1.0.7](https://github.com/air-gapped/cooked/compare/v1.0.6...v1.0.7) (2026-02-07)


### Bug Fixes

* align first line of code blocks with subsequent lines ([d62eee4](https://github.com/air-gapped/cooked/commit/d62eee43667b83920359fd44a81145b71c9aaa07))
* strip leading space-padding from chroma line numbers ([0bf1f05](https://github.com/air-gapped/cooked/commit/0bf1f05377332e4a951635018708ff59f54fb7d2))

## [1.0.6](https://github.com/air-gapped/cooked/compare/v1.0.5...v1.0.6) (2026-02-07)


### Bug Fixes

* eliminate two-character offset on first line ([6bd172d](https://github.com/air-gapped/cooked/commit/6bd172d0cd7238ca58d3447c5e1a0b75498deccc))
* eliminate two-character offset on first line by disabling chroma line number padding ([c0231a6](https://github.com/air-gapped/cooked/commit/c0231a63598c13b37ec2a3f628a3d90defea1d22)), closes [#11](https://github.com/air-gapped/cooked/issues/11)
* remove non-existent WithLineNumberPad chroma option ([c09d01c](https://github.com/air-gapped/cooked/commit/c09d01c1df9caf7c382cc3bd7e2bcd4a8975bb92))

## [1.0.5](https://github.com/air-gapped/cooked/compare/v1.0.4...v1.0.5) (2026-02-07)


### Bug Fixes

* remove left padding from line numbers to eliminate two-character offset ([66763c1](https://github.com/air-gapped/cooked/commit/66763c1bd87e479cbe1e7256795fe9c1fc8981d2))
* remove left padding from line numbers to eliminate two-character offset ([4e212e5](https://github.com/air-gapped/cooked/commit/4e212e58e328de6830c7453e83fc38f57b46e1dc)), closes [#11](https://github.com/air-gapped/cooked/issues/11)

## [1.0.4](https://github.com/air-gapped/cooked/compare/v1.0.3...v1.0.4) (2026-02-07)


### Bug Fixes

* use full ghcr.io registry path for Docker images in README ([453c12c](https://github.com/air-gapped/cooked/commit/453c12c7be3187484e99ef2b665cce38e558faed))

## [1.0.3](https://github.com/air-gapped/cooked/compare/v1.0.2...v1.0.3) (2026-02-07)


### Bug Fixes

* add github.com host to kaniko git context URL ([5802f3d](https://github.com/air-gapped/cooked/commit/5802f3daa61eaa517a4a9100fd17fabae3241dee))

## [1.0.2](https://github.com/air-gapped/cooked/compare/v1.0.1...v1.0.2) (2026-02-07)


### Bug Fixes

* use git:// protocol for kaniko context URL ([679bc4e](https://github.com/air-gapped/cooked/commit/679bc4e64d6d91b1c44583913b3bd653d765b095))

## [1.0.1](https://github.com/air-gapped/cooked/compare/v1.0.0...v1.0.1) (2026-02-07)


### Bug Fixes

* use crane for Docker image builds instead of BuildKit ([575f8e6](https://github.com/air-gapped/cooked/commit/575f8e6b826068ef8a59db80de0d2545951330bd))
* use kaniko for Docker image builds on ARC runners ([e2f5fbf](https://github.com/air-gapped/cooked/commit/e2f5fbf1be24537e966c5e04d6e84114234bce11))

## 1.0.0 (2026-02-07)


### Features

* add automated releases with release-please ([7066965](https://github.com/air-gapped/cooked/commit/706696543042f32b66368dad8b55178b65b4ed54))


### Bug Fixes

* add container to all release workflow jobs ([248f377](https://github.com/air-gapped/cooked/commit/248f377e65e4cdf0ed9e57fc9434e94b2de73f81))
* use GitHub App token for release-please PR creation ([c85c4b1](https://github.com/air-gapped/cooked/commit/c85c4b1002653e60c2eee7f45e4910eda5ad18e7))
