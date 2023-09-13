## [1.0.4](https://github.com/Randsw/k8s-operator-CascadeDeployment/compare/1.0.3...1.0.4) (2023-09-13)


### 🛠 Fixes

* Use new field in manager Options ([2a165f7](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/2a165f77a9120f7e4ad088c68cfc18fad8b2d9f7))


### Other

* **deps:** bump actions/checkout from 3 to 4 ([906cb6e](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/906cb6e9d7e25d8531043afd90ffae0353fed16e))
* **deps:** bump docker/build-push-action from 4 to 5 ([a264c17](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/a264c17c8aa9b4a18755ea9b29a6c79278c31c31))
* **deps:** bump docker/login-action from 2 to 3 ([fb08a1c](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/fb08a1c563aa655d665e9f3275737706d8553a9a))
* **deps:** bump docker/setup-buildx-action from 2 to 3 ([54c3027](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/54c3027300a7bc929f81b66ce3599f6cfa26a9cb))
* **deps:** bump golang from 1.17 to 1.21 ([c76bd17](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/c76bd178c12f47d8b263964816f647f496f006af))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.11.0 to 0.16.2 ([37813d3](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/37813d3aca8480ea7f82d19a1a7deacec25e2b4c))
* **deps:** bump tj-actions/changed-files from 35 to 39 ([1010101](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/10101018a33e25d735e866ddf1bc1197af13b0cf))
* **helm:** Bump app and chart versionin helm chart ([82a372e](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/82a372e233dcc11f88d7eee23d06cd9b91aacaa1))

## [1.0.3](https://github.com/Randsw/k8s-operator-CascadeDeployment/compare/1.0.2...1.0.3) (2023-08-18)


### 🛠 Fixes

* **app:** Add more comments to tests ([29d0497](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/29d04977f27f5cb2d7ea5965e95117a1c5f4d941))

## [1.0.2](https://github.com/Randsw/k8s-operator-CascadeDeployment/compare/1.0.1...1.0.2) (2023-08-18)


### 🛠 Fixes

* **app:** Add more metrics help ([b0623d8](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/b0623d8cb9392a68627873a7e4fd8aac43b84a6f))

## [1.0.1](https://github.com/Randsw/k8s-operator-CascadeDeployment/compare/1.0.0...1.0.1) (2023-08-18)


### 🛠 Fixes

* **app:** Fix errors in metric definition ([b2775e1](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/b2775e1791dbd3f3dcc5beb4b541be0b34947560))

## [1.0.0](https://github.com/Randsw/k8s-operator-CascadeDeployment/compare/...1.0.0) (2023-08-18)


### 🚀 Features

* **gh-actions:** Add more brnaches name for execution pipeline ([c8f4cb5](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/c8f4cb58304fbe4afaf81af8d4d5d0cedfd1643a))
* **gh-actions:** Upgrade makefile to enable testing. Add gh-action files ([a80002e](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/a80002eb75fd35b6ba8366ec7f13bacf3e428c86))
* **monitoring:** Add instance count metric ([2d98391](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/2d98391735511b84991cc62d147bccc3f1be46c2))


### 🛠 Fixes

* **app:** Remove logger in struct. Add kind config for test. Metrics bind to 0.0.0.0 ([912bd28](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/912bd28798b9b2fed16100906b97d9ee6589ea9e))
* **gh-actions:** Add maintainer ([daa814c](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/daa814c94898ee94e7d823d8275850566b316134))
* **gh-actions:** Change metrics bind address in config ([120b40d](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/120b40d93d137fcb90acca7ec55ed750106da540))
* **gh-actions:** Fix helm chart dir in gh-actions files ([0f2306c](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/0f2306cfd082cb2209ca7a6f3414e1076333cf1b))
* **gh-actions:** Fix helm chart name ([9640c1a](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/9640c1a74ec7cdf78bec0973cd1ec99c89600848))
* **gh-actions:** Fix helm lint errors ([8a91f2e](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/8a91f2eca4eab814fb1825971e5689e02e3d1fda))
* **gh-actions:** Fix helm lint errors ([a0b9b33](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/a0b9b33d3d8e1b86cd2cc53ee729398b7009c76b))
* **gh-actions:** Fix helm test action config file ([744601d](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/744601d96108ac1c63d241a99117089d6eec0520))
* **gh-actions:** Fix helm test action file glob ([a9c8288](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/a9c8288604568b7923273d679b0f793b86ec2be6))
* **gh-actions:** Fix lint errors ([3b57315](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/3b57315b56e51d35b60824372b52fff9fc081ded))
* **gh-actions:** Remove test connection pod ([3438fc3](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/3438fc3a0d2cec7be6ba9aad296f2f1c730da506))
* **gh-actions:** Rename helm chart dir ([aae3677](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/aae3677debed7463eea93b9068423abfc770c983))
* **kustomize:** Update config syntax ([74fda29](https://github.com/Randsw/k8s-operator-CascadeDeployment/commit/74fda29f00554a8efbb61e8153f2ab5e4a4b9db6))
