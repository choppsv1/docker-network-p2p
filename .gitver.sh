#!/bin/bash

HEAD_COMMIT=$(git rev-list --abbrev-commit --max-count=1 HEAD)
TAG_COMMIT=$(git rev-list --abbrev-commit --tags --max-count=1)
TAG=$(git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)

if [[ -z $TAG ]]; then
    echo git${HEAD_COMMIT}
else
    # Remove an preceding v in vX.X
    TAG=${TAG#v}
    if [[ $HEAD_COMMIT == $TAG_COMMIT ]]; then
        echo $TAG
    else
        echo $TAG-dev
    fi
fi
