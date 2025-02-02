# Changelog

## 0.1.0 (2025-02-02)


### ⚠ BREAKING CHANGES

* Existing build and deployment processes will require updates to use uv instead of Rye.
* Dropped support for Node 6 and older Python versions.
* The CI workflow now relies on Makefile targets, which may require adjustments in local development setups.

### ✨ Features

* add CI workflow for automated builds and deployments ([e530491](https://github.com/liblaf/route-rules/commit/e530491ea40717fbd146b986de2619908c148191))
* add new domain suffixes for Emby service ([13ad85a](https://github.com/liblaf/route-rules/commit/13ad85a7d788bbb6734b9f42d2e2ccbcf4c80a6b))


### ⬆️ Dependencies

* **deps:** update astral-sh/setup-uv action to v4 ([#19](https://github.com/liblaf/route-rules/issues/19)) ([ff427e1](https://github.com/liblaf/route-rules/commit/ff427e1341c69f92773bbbc8f8cab3677208e61b))
* **deps:** update eifinger/setup-rye action to v4 ([01a06e8](https://github.com/liblaf/route-rules/commit/01a06e8514d9177b3e0357d998e2e5654cac4cc1))


### 🐛 Bug Fixes

* correct domain suffix parsing in rule optimization ([edaf1ed](https://github.com/liblaf/route-rules/commit/edaf1ed575c023bcc3d086760b0782e01a990152))
* correct GeoSite and GeoIP references and add missing rule sets ([525d396](https://github.com/liblaf/route-rules/commit/525d3966167e40a9b55ccff6df45fa5316ee4a83))
* correct rule set exclusion and import logic ([cfc1fa6](https://github.com/liblaf/route-rules/commit/cfc1fa674d248c7e4eadd4b9993ca6c2cd59f039))
* correct rule set exclusion and import logic ([5161a0a](https://github.com/liblaf/route-rules/commit/5161a0aab514b448a2c6b2a64a9d244673aa0638))
* correct rule set exclusion logic and update documentation ([3e0e2e9](https://github.com/liblaf/route-rules/commit/3e0e2e94da5cee05ee2653bbc559436b080299b8))
* correct rule set order in build configuration ([3dcea85](https://github.com/liblaf/route-rules/commit/3dcea85a16eb16e477f62a1eee03a6ef16771b4f))
* correct rule set version and integrate new compilation script ([7087d39](https://github.com/liblaf/route-rules/commit/7087d39479957715924cd3332727aedea08d9b8d))
* correct source and preset configuration for new rule set ([a8796b9](https://github.com/liblaf/route-rules/commit/a8796b9b2b5825adde7fde78b9527654e84f5641))
* correct summary script and update README links ([9a01e56](https://github.com/liblaf/route-rules/commit/9a01e56b4fb33b5e99f5681178f32175b843bf54))
* exclude CN ruleset from proxy and direct rulesets ([36a373e](https://github.com/liblaf/route-rules/commit/36a373e90d1169a556c7aae2a3349a16aceca853))
* exclude YouTube from AI rule set to prevent data leakage ([66b6028](https://github.com/liblaf/route-rules/commit/66b602834c8879c302d87bdd09fef4766e7e5d2d))
* generalize rule compilation and update summary timestamp ([58eacf2](https://github.com/liblaf/route-rules/commit/58eacf22777533e14c6654b69e2f96020ba1b562))
* refactor rule set export and save logic ([3a64632](https://github.com/liblaf/route-rules/commit/3a6463200a3a6c803e5cab4a47bc39efde58745f))
* refactor rule set generation and optimization ([ae1b1f2](https://github.com/liblaf/route-rules/commit/ae1b1f20c09e4431db495650c123975d7b3dedd9))
* remove `PROCESS-NAME` rules ([c978f31](https://github.com/liblaf/route-rules/commit/c978f31833a8a36d9dce38c0775c1302a7ee6bf6))
* remove redundant developer category from source lists ([15e1804](https://github.com/liblaf/route-rules/commit/15e18040aede763e892225366c6077051b9401e9))
* **scripts:** correct the save method in the build script ([323211f](https://github.com/liblaf/route-rules/commit/323211fefdf68046723a229670657f3717b44784))
* streamline CI workflow and update project structure ([6a0afa7](https://github.com/liblaf/route-rules/commit/6a0afa7e2135c5c75e3c1f47cc628d079e32423c))
* update documentation and add rule set statistics ([72e417a](https://github.com/liblaf/route-rules/commit/72e417a435f811755f045e6a9d61ac4ec48cbe70))
* update domain suffixes to include additional domains ([f8e0758](https://github.com/liblaf/route-rules/commit/f8e0758805460a1f15809b8270991babd9b037e0))
* update installation method for sing-box due to upstream issue ([e305025](https://github.com/liblaf/route-rules/commit/e3050252e2f176101c0874676a7add3acd11a0d2))


### 📝 Documentation

* add Cloudflare download links for rule sets ([e6400b8](https://github.com/liblaf/route-rules/commit/e6400b859f51b761cce77eb8df41a0727405fd49))
* add GitHub Actions status badge and last commit badge to README ([8116439](https://github.com/liblaf/route-rules/commit/8116439bbbaa8d310dc689d8e2ca7d78ae8014bb))
* add GitHub Actions workflow status badge to README ([de5931f](https://github.com/liblaf/route-rules/commit/de5931f08109eca11d03018d7b94470f1c980412))
* add instruction to remove unresolvable domains ([bd09087](https://github.com/liblaf/route-rules/commit/bd090879d186f504c407f2238ede55827f69aff0))
* add repository size and star count badges ([cdd9fd9](https://github.com/liblaf/route-rules/commit/cdd9fd98fd708a006f4bd68845f060be075cdfe8))
* enable navigation sections in documentation ([37e99ed](https://github.com/liblaf/route-rules/commit/37e99ede8064ba42bc0c2a52f83480b99028cd4e))
* enhance documentation build and site configuration ([c58b69b](https://github.com/liblaf/route-rules/commit/c58b69bca19677f77e1fef19a635d52caaafc34e))
* enhance documentation with new configuration examples and improved formatting ([87fd223](https://github.com/liblaf/route-rules/commit/87fd2236f5d13c8bffa2fcf7049318daf7766a6f))
* enhance readability and consistency in README and build script ([a03f86b](https://github.com/liblaf/route-rules/commit/a03f86bba12122ffe20cffcaa7c97225c64aa7a9))
* remove outdated site URL from configuration ([a095a08](https://github.com/liblaf/route-rules/commit/a095a08f481b45bae845bb3a2f67c14477b430d5))
* reorganize documentation for clarity and accessibility ([53e1e9c](https://github.com/liblaf/route-rules/commit/53e1e9c723c02e3f7a6290a09d7eed4d1473742f))
* standardize anchor links in documentation ([562e9d4](https://github.com/liblaf/route-rules/commit/562e9d46a56e93ecd89de95baa2d9b97e7a9f074))
* update DNS configuration and documentation for improved clarity and consistency ([874e1e8](https://github.com/liblaf/route-rules/commit/874e1e88951d26bc72d5c4b732021d824fd0a530))
* update documentation configuration and remove deprecated plugin ([b13c501](https://github.com/liblaf/route-rules/commit/b13c501dcc30dc4beb478081dacef5c8aaaacf3a))
* update documentation configuration for improved navigation and source display ([2818d18](https://github.com/liblaf/route-rules/commit/2818d186354e611a1b977006e2dbd5509925632c))
* update documentation for clarity and consistency ([0005523](https://github.com/liblaf/route-rules/commit/00055231811c88d1ac7b3dfa8666fc927c5945a2))
* update documentation links and reorganize structure ([368ed1b](https://github.com/liblaf/route-rules/commit/368ed1b69130c59581c5cfa9e2e13867a4eeaac5))


### ♻ Code Refactoring

* migrate from Rye to uv for dependency management ([7b5b14d](https://github.com/liblaf/route-rules/commit/7b5b14d081482ee1b60c1b9049ba70bacb263df0))
* overhaul build and CI workflows ([aceb22b](https://github.com/liblaf/route-rules/commit/aceb22b976c1e1eaccc5fbcf7722ea1e59b4f814))
* rename package from 'sbr' to 'route_rules' ([cd8bbda](https://github.com/liblaf/route-rules/commit/cd8bbda8357b1af4925b95921dd7f0a8b125f4c5))
* update dependency management and lock files ([2427e52](https://github.com/liblaf/route-rules/commit/2427e52bc31dd14d5def2150ca174ea1a2da20c9))


### 🔧 Continuous Integration

* add GH_TOKEN environment variable to CI workflow ([8c0a51a](https://github.com/liblaf/route-rules/commit/8c0a51aec206a962066c1b6eafa69e9c0cc99c9e))
* add Prettier installation to CI workflow and format timestamp in summary script ([82ff20d](https://github.com/liblaf/route-rules/commit/82ff20d2a0604d2934c161203b84eb39a7fdf75c))
* add scheduled documentation build and deployment workflow ([6627365](https://github.com/liblaf/route-rules/commit/66273659b7c4e8a803f2cb263c3c1ed8c42814c6))
* **pre-commit:** auto fixes from pre-commit hooks ([bb5cabf](https://github.com/liblaf/route-rules/commit/bb5cabf5539214c1a12c00e3c9cf7148888f8587))
* **repo:** sync with repo template ([28e700f](https://github.com/liblaf/route-rules/commit/28e700f434aab66696160a6df84d4e55215daf75))
* **repo:** sync with repo template ([fb03a54](https://github.com/liblaf/route-rules/commit/fb03a5448fb840362e3c3821027fd65c264d1add))
* **repo:** sync with repo template ([cbccf22](https://github.com/liblaf/route-rules/commit/cbccf221e503298066de3be34ffe21f470c2f6d4))
* **repo:** sync with repo template ([1e3ad1b](https://github.com/liblaf/route-rules/commit/1e3ad1b9406820a773989b478728e7215f12b3c2))
* **repo:** sync with repo template ([0673ed6](https://github.com/liblaf/route-rules/commit/0673ed6e96ccd32df982d780ca77ad92b349bd74))
* **repo:** sync with repo template ([bce241f](https://github.com/liblaf/route-rules/commit/bce241fe725ea130954c8fe72a593e51238f3d1b))
* **repo:** sync with repo template ([bbd2b0b](https://github.com/liblaf/route-rules/commit/bbd2b0b430c3e1940ed32fa7c8953ac48dde143e))
* **repo:** sync with repo template ([8be6efe](https://github.com/liblaf/route-rules/commit/8be6efed09a9aaf433c493651e3370c99392c706))
* **repo:** sync with repo template ([b07d9d7](https://github.com/liblaf/route-rules/commit/b07d9d74c5c03075cd053092ebd9de4b1d4f4f8f))
* **repo:** sync with repo template ([e3bae26](https://github.com/liblaf/route-rules/commit/e3bae262743c21898dc704ec7cc55d4539e82aa9))
* update build script path in CI configuration ([df7d7e3](https://github.com/liblaf/route-rules/commit/df7d7e30f1c87d4fe333e56bb4cfe0032ba7ad66))
* update GitHub Actions deploy action for streamlined deployment ([b17c190](https://github.com/liblaf/route-rules/commit/b17c190c6b769ec3a4af6601641304918c0e7d39))
* update GitHub Actions deployment to use peaceiris/actions-gh-pages ([04719e4](https://github.com/liblaf/route-rules/commit/04719e4dbd23dffc9a686261d44c8356a21cc5a5))
* update publish directory for CI workflow ([52b74a9](https://github.com/liblaf/route-rules/commit/52b74a9778c647b6582d1c83ad9a740055d9ed96))
