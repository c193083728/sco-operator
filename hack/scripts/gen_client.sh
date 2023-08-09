#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"
TMP_DIR=$( mktemp -d -t sco-client-gen-XXXXXXXX )

mkdir -p "${TMP_DIR}/client"
mkdir -p "${PROJECT_ROOT}/pkg/client/sco"

"${PROJECT_ROOT}"/bin/applyconfiguration-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/sco1237896/sco-operator/api/sco/v1alpha1 \
  --output-package=github.com/sco1237896/sco-operator/pkg/client/sco/applyconfiguration

"${PROJECT_ROOT}"/bin/client-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input=sco/v1alpha1 \
  --clientset-name "versioned"  \
  --input-base=github.com/sco1237896/sco-operator/api \
  --apply-configuration-package=github.com/sco1237896/sco-operator/pkg/client/sco/applyconfiguration \
  --output-package=github.com/sco1237896/sco-operator/pkg/client/sco/clientset

"${PROJECT_ROOT}"/bin/lister-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/sco1237896/sco-operator/api/sco/v1alpha1 \
  --output-package=github.com/sco1237896/sco-operator/pkg/client/sco/listers

"${PROJECT_ROOT}"/bin/informer-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/sco1237896/sco-operator/api/sco/v1alpha1 \
  --versioned-clientset-package=github.com/sco1237896/sco-operator/pkg/client/sco/clientset/versioned \
  --listers-package=github.com/sco1237896/sco-operator/pkg/client/sco/listers \
  --output-package=github.com/sco1237896/sco-operator/pkg/client/sco/informers


cp -R "${TMP_DIR}"/client/github.com/sco1237896/sco-operator/pkg/client/sco/* "${PROJECT_ROOT}"/pkg/client/sco