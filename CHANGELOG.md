# Changelog

All notable changes to this project will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project attempts to adhere to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 1.0.0-alpha.0 (2025-09-24)


### 🚀 Features

* add optional basic auth for metrics endpoint ([c5ccdce](https://github.com/engineervix/pseudoc/commit/c5ccdce1e67ad3106ed966cd6ba6f70bf80de11d))
* add quiet mode, dry-run, enhanced version info and better error handling ([e8ac63c](https://github.com/engineervix/pseudoc/commit/e8ac63ce7ad84975e094f79ff583120e7f83fbe3))
* **api:** enhance request validation and error handling ([a4b94ee](https://github.com/engineervix/pseudoc/commit/a4b94ee33367d33681412fe4803b575bcff36dfa))
* **api:** implement document generation endpoints ([55760bb](https://github.com/engineervix/pseudoc/commit/55760bb53ef6cd0f52272b257832fc9ef2687c64))
* **cli:** improve server startup experience ([f4641c7](https://github.com/engineervix/pseudoc/commit/f4641c7953ccb8fa87802f132193264a7df9fa5a))
* create document generator interface with mock implementations ([f4baee6](https://github.com/engineervix/pseudoc/commit/f4baee6963698ee42ae9e8897d437a78f015f084))
* implement configuration package with validation and file handling ([a9c3336](https://github.com/engineervix/pseudoc/commit/a9c3336bc16e8f703f869b54222d26e3251d814a))
* implement DOCX document generation with gomutex/godocx library ([cea6f98](https://github.com/engineervix/pseudoc/commit/cea6f9832dafe3a570c179c1396547ecbc5f454b))
* implement HTTP server infrastructure with Echo ([d72347c](https://github.com/engineervix/pseudoc/commit/d72347c878fcd221546a2989f1373490e3e5ad82))
* implement PDF generation with maroto library and Lorem Ipsum content ([c5bc64c](https://github.com/engineervix/pseudoc/commit/c5bc64c7d397e415d675c4a0a34ea434b7035854))
* implement performance metrics tracking ([0108696](https://github.com/engineervix/pseudoc/commit/0108696409503c2e85ba2c06eda89eebfc291eac))
* implement random document type selection with proper seed handling ([592e2c0](https://github.com/engineervix/pseudoc/commit/592e2c085f26e032c59f22ee3aae3b2a1552298e))
* implement XLSX spreadsheet generation with excelize library ([2f34963](https://github.com/engineervix/pseudoc/commit/2f34963864922db4ddf0790879326ccb058cdab6))
* initial CLI structure with command parsing and flag handling ([1cee42a](https://github.com/engineervix/pseudoc/commit/1cee42a67e104bfd5e04b29a14b33ccf9650f779))
* integrate config and generator packages into CLI ([e22d7b5](https://github.com/engineervix/pseudoc/commit/e22d7b522d112165958daf284822e007e671c034))
* integrate HTTP server with CLI command routing ([b645094](https://github.com/engineervix/pseudoc/commit/b645094963e8f7d94ebda6f3a54621bbd5bda562))
* my first go project, let's go!🚀 ([9b020ec](https://github.com/engineervix/pseudoc/commit/9b020ec39de44cbff0f4a552cbef16bfbd634a16))
* **server:** add custom IP extractor for reverse proxies ([b65721d](https://github.com/engineervix/pseudoc/commit/b65721def82889c27799c14c90a79f61f54befc3))
* **server:** metrics and monitoring ([d1ca072](https://github.com/engineervix/pseudoc/commit/d1ca072c6c973208de71c3eecfaccd01fe848b29))
* udpate docs ([3e1714b](https://github.com/engineervix/pseudoc/commit/3e1714b9efc6daa1af82eb6818ade141ba548ec7))


### 🐛 Bug Fixes

* disable CORS by default for security ([65e0911](https://github.com/engineervix/pseudoc/commit/65e09116ddd8f13ca4a65bb00b8557ba0df95782))
* improve error handling and memory management in document generation ([b91cca3](https://github.com/engineervix/pseudoc/commit/b91cca33d23f018a2935a62fc76ad33e91b0cf8b))
* improve rate limiting for random endpoint workflow ([d79d71b](https://github.com/engineervix/pseudoc/commit/d79d71b4bc796ed7e2599cec4324e78ace97d4c5))
* standardize API response patterns for consistency ([eebaf42](https://github.com/engineervix/pseudoc/commit/eebaf426699b0f390078e2578d07c07323aedaa4))
* use config constants instead of hardcoded values in info handler ([c0706ae](https://github.com/engineervix/pseudoc/commit/c0706ae312d8a3d5cb8a744ff7be2da0efabd039))


### 📝 Docs

* add API docs, powered by swaggo/swag ([7e98891](https://github.com/engineervix/pseudoc/commit/7e9889129d7c66b221a1b77b6031d068ff8a74a5))
* add README ([eb243b0](https://github.com/engineervix/pseudoc/commit/eb243b089f693d5957a748d58fe3d9f1d2f2d6f0))
* update README to reflect current API implementation ([52bad6b](https://github.com/engineervix/pseudoc/commit/52bad6bdc593954dd360435d7737491ff467017a))


### ♻️ Code Refactoring

* extract hardcoded constants to centralized config ([96cdd6b](https://github.com/engineervix/pseudoc/commit/96cdd6b03e82e9882bd412f057708e485a3f2042))
* improve request ID generation with crypto/rand ([6db6d1d](https://github.com/engineervix/pseudoc/commit/6db6d1d583464143b6b8805d8eaae1230ee720eb))
* remove size flag and add CLI behaviour flags to config ([93e867b](https://github.com/engineervix/pseudoc/commit/93e867b4f10ee6d8edc65cde9c80583f7bc2af4e))
* separate CLI and core document generation configuration ([eadfb61](https://github.com/engineervix/pseudoc/commit/eadfb6181dcb34a9528c090fc4648f15993437e5))
* **server:** enhance health check with uptime and system metrics ([8ab6b9b](https://github.com/engineervix/pseudoc/commit/8ab6b9bd203d0c44e30dde4bdec20231b98e7935))
* **server:** make security and CORS settings configurable ([5c33b15](https://github.com/engineervix/pseudoc/commit/5c33b15b2a29f09d75e38c18c97ed1611cf9b9f7))
* **server:** update random endpoint, add env config, fix rate limiting ([e655bde](https://github.com/engineervix/pseudoc/commit/e655bde4e584932944482012ca435d1f5dd06290))
* simplify PDF generation with predictable page-based content structure ([ae104f9](https://github.com/engineervix/pseudoc/commit/ae104f95f166cbf57da7f56086c0b29bc3f7735e))
* structured JSON logging with dynamic levels ([289b454](https://github.com/engineervix/pseudoc/commit/289b454459fd7095ab030d0fad8319c725bdf30c))


### ⚡️ Performance Improvements

* improve metrics storage with efficient ring buffer ([23b0abc](https://github.com/engineervix/pseudoc/commit/23b0abcc61f38e4f5fdad152ea4ef138ef443cf6))


### ✅ Tests

* improve CLI test coverage with better Go testing practices ([2942338](https://github.com/engineervix/pseudoc/commit/294233808d908fcc3712a4298ba4d49d6a675d02))
* update tests with output capture and reproducibility validation ([1b5fc81](https://github.com/engineervix/pseudoc/commit/1b5fc8158ac75c407692d6acd2c362debe8d1198))


### ⚙️ Build System

* add version injection and enhanced build tooling ([c849955](https://github.com/engineervix/pseudoc/commit/c8499559b2a65a6c55a1051182f9d35d694b0f6f))
* **deps:** add dependencies for HTTP server ([80e6da2](https://github.com/engineervix/pseudoc/commit/80e6da2fdceaf951631ee3327a8d82e47dcdd941))
* improve ldflags for version injection ([1b50376](https://github.com/engineervix/pseudoc/commit/1b50376bf1fedbd4f616cf5f111db4543b4d7e19))


### 👷 CI/CD

* add renovate config ([0f7cc7b](https://github.com/engineervix/pseudoc/commit/0f7cc7bd72b7c27753268c514f87f7552944c17d))
* **config:** migrate config .github/renovate.json ([#2](https://github.com/engineervix/pseudoc/issues/2)) ([50e9742](https://github.com/engineervix/pseudoc/commit/50e9742903b8828cb5db69c1b7943eeaf35193cc))
* do not forget to create a windows build ([2b1e845](https://github.com/engineervix/pseudoc/commit/2b1e845bed3c2d8ba7c1cfa7af099eb5868498ce))
* windows runners seem to be tricky ([839c842](https://github.com/engineervix/pseudoc/commit/839c84243dd3438a3255b28dc3d57413f8f5dbbd))
