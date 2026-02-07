# Changelog

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
