# foxful-cli (in-tree)

This directory holds the TUI framework [foxful-cli](https://github.com/anhoder/foxful-cli)
by anhoder, maintained in-tree for this project. The root `go.mod` points
`github.com/anhoder/foxful-cli` here with a `replace` directive, so import paths stay
the same as upstream go-musicfox.

Base: v1.1.4
Origin: https://github.com/anhoder/foxful-cli @ 676bd0acd848141d857f6c1f18f0c45bd6f698f1 (refs/tags/v1.1.4)
Module sum: h1:eXrqVpjlZU99963JNH/2n2Py35te/xCC6ts/NTgeOrk=

The commit "chore(foxful-cli): import pristine v1.1.4" contains the module exactly as
published; `git diff <that commit> -- third_party/foxful-cli` shows every local change.

## License

The upstream repository has no LICENSE file at this version. The framework is
redistributed here as part of go-musicfox (GPL-3.0), whose upstream repository also
vendors it. Do not add a file here that claims a license on behalf of the upstream
author.

## Updating

1. Diff the old base against the new upstream version and merge that diff into this
   directory (three-way).
2. Update `Base` above and the `require` version in the root `go.mod` together;
   `internal/forkcheck` fails while they disagree.
3. Run `make vendor` from the repository root. `vendor/github.com/anhoder/foxful-cli/`
   is generated from this directory and must never be edited by hand.
