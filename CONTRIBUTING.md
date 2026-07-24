# Contributing to leitwerk-devkit

## Build and test

The toolchain and build tasks are managed by [mise](https://mise.jdx.dev). On a
fresh clone run `mise trust` once, then:

```
mise run build     # build the gate binary -> core/bin/leitwerk
mise run test      # unit + integration tests
mise run vet       # go vet
```

Every change must pass the gate at its blast-radius tier before it lands:

```
leitwerk verify --tier <T0|T1|T2>    # `leitwerk tier <path>` reports the tier
```

## Commit and PR conventions

Commits and pull-request titles follow [Conventional Commits](https://www.conventionalcommits.org).
This is not cosmetic: `release-please` derives the next version and the changelog
from commit history, and because pull requests are **squash-merged**, the PR
**title** becomes the commit on `main` that release-please reads. A pull request
whose title is not a valid Conventional Commit is rejected by the `semantic-pr`
check.

The allowed **types and scopes**, with examples, live in
[`.gitmessage`](.gitmessage) — the single source. Set it as your commit template
so the editor shows them:

```
git config commit.template .gitmessage
```

`feat` bumps the minor and `fix` the patch; the other types do not change the
version. A breaking change (`!` or a `BREAKING CHANGE:` footer) bumps the minor
pre-1.0. The scope is optional. `semantic-pr` enforces the type and format;
scopes are advisory (kept in `.gitmessage`).

## Releases

Releases are automated (see `leitwerk/specs/cli-publish.md`): `release-please`
keeps a Release-PR open reflecting the pending version and changelog; merging it
creates the `core/vX.Y.Z` tag and GitHub Release, and the build job attaches the
prebuilt binaries. Only a human merging the Release-PR cuts a release.
