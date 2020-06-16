#!/bin/bash

# Copyright The Stash Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eou pipefail

should_cherry_pick() {
    while IFS=$': \t' read -r -u9 marker v; do
        if [ "$marker" = "/cherry-pick" ]; then
            return 0
        fi
    done 9< <(git show -s --format=%b)
    return 1
}

should_cherry_pick || {
    echo "Skipped cherry picking."
    echo "To automatically cherry pick, add /cherry-pick to commit message body."
    exit 0
}

git remote set-url origin https://${GITHUB_USER}:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}.git

while IFS=/ read -r -u9 repo branch; do
    git checkout $branch
    pr_branch="master-${GITHUB_SHA:0:8}"${branch#"release"}
    git checkout -b $pr_branch
    git cherry-pick --strategy=recursive -X theirs $GITHUB_SHA
    git push -u origin HEAD -f
    hub pull-request \
        --base $branch \
        --labels automerge \
        --message "[cherry-pick] $(git show -s --format=%s)" \
        --message "$(git show -s --format=%b | sed --expression='/\/cherry-pick/d')" || true
    sleep 2
done 9< <(git branch -r | grep release)
