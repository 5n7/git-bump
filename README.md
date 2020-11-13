# git-bump

Bump semver for Git.

## Installation

Download the binary from [GitHub Releases](https://github.com/skmatz/vin/releases).

Or, if you have Go, you can install `git-bump` with the following command.

```console
go get github.com/skmatz/git-bump/...
```

## Usage

```console
> git-bump

tags:
  - v0.1.0
  - v0.2.0
  - v0.3.0
  - v1.0.0-rc1 (current version)

? select the next version
> patch: v1.0.0
  minor: v1.1.0
  major: v2.0.0
```

## References

- <https://github.com/b4b4r07/git-bump>
