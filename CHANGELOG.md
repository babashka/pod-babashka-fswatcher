# Changelog

[pod-babashka-fswatcher](https://github.com/babashka/pod-babashka-fswatcher): babashka filewatcher pod

## 0.0.7

- Bump `fsnotify` to `1.9.0`
- Pod should not send `done` for intermediate async results (fixes compatibility with newer babashka versions that handle async requests according to pod spec)

## 0.0.6

- [#22](https://github.com/babashka/pod-babashka-fswatcher/issues/22): fix macos aarch64 binary
- upgrade to golang 1.24


## 0.0.5

- [#17](https://github.com/babashka/pod-babashka-fswatcher/issues/17): Add macOS aarch64 binary ([@lispyclouds](https://github.com/lispyclouds))
- [#18](https://github.com/babashka/pod-babashka-fswatcher/issues/18): deduplicate by default ([@fjsousa](https://github.com/fjsousa))
