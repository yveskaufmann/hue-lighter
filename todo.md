# TODO: First release (v1.0.0)

The first run of the Release workflow will produce `v1.0.0` (no tags yet, the
`feat!` commit is a breaking change anyway).

## 1. Code

- [ ] Run `go fmt ./...`. The CI formatting check fails on 7 files:
      `internal/app/app.go`, `internal/app/bootstrap.go`, and 5 already committed
      files in `internal/hue_client` and `internal/services/light_automation`.
- [ ] Update the README:
  - [ ] macOS install: `brew tap yveskaufmann/tap`, `brew trust yveskaufmann/tap`,
        `brew install --cask hue-lighter`, then start the app once.
  - [ ] macOS file locations: config in `~/Library/Application Support/hue-lighter/`,
        logs in `~/Library/Logs/hue-lighter.log`, API key in the Keychain.
  - [ ] Local development: `make mac_install` / `make mac_uninstall`.
  - [ ] Linux: release tarballs (`hue-lighter_<v>_linux_<arch>.tar.gz`) next to the Flatpak.
- [ ] Note in `CONTRIBUTING.md` that commits must follow
      [Conventional Commits](https://www.conventionalcommits.org/). They decide the version and the changelog.
- [ ] Optional: when there is no config yet, the app exits with a fatal error you can't
      see when it's started from Finder. Consider opening the settings page on first start instead.

## 2. Test on a Mac (not done yet)

Run `make mac_install`, then check that:

- [ ] the tray icon appears and there is no Dock icon
- [ ] macOS asks for local network access, and the bridge is found
- [ ] `~/Library/LaunchAgents/com.github.yveskaufmann.hue-lighter.plist` gets written
- [ ] logs end up in `~/Library/Logs/hue-lighter.log` (after the first start via login)
- [ ] the settings page saves to `~/Library/Application Support/hue-lighter/config.yaml`
- [ ] the lights turn off when you log out, and the app starts again after you log in
- [ ] Quit from the tray menu does not restart the app
- [ ] `make mac_uninstall` removes the app and the LaunchAgent

## 3. Commit and merge

- [ ] Commit the changes with conventional commit messages, e.g.
      `feat(macos): app bundle with login autostart`,
      `ci: release workflow with semantic-release and Homebrew tap`.
- [ ] Open a PR. The `main` ruleset requires a PR and status checks.
      Local `main` is also 1 commit ahead of `origin` (`9b3402d`).
- [ ] Make sure CI passes on Linux. The OS-specific tests in `internal/paths` and
      `internal/config` have only run on macOS so far.

## 4. GitHub settings

### hue-lighter repo

- [ ] Secret `HUE_CA_BUNDLE`: PEM content of `configs/certs/cacert_bundle.pem`
      (or change the "Provide Hue CA bundle" steps in `release.yaml`).
- [ ] Environment `homebrew-tap`:
  - [ ] required reviewer: you
  - [ ] secret `TAP_GITHUB_TOKEN`: fine-grained PAT, repository access only to
        `yveskaufmann/homebrew-tap`, permission Contents: read and write, with an expiry date
- [ ] Turn on **immutable releases** (Settings → General → Releases). Do this before
      the first release; it only applies to releases published afterwards.
- [ ] Optional: a tag ruleset for `v*` (no update or deletion). It must allow GitHub
      Actions to create tags, otherwise semantic-release can't tag.
- [ ] Decide about the CA bundle in git history: it's in `ecad1a2` on `main` and in
      `f70291c` on `copilot/vscode-mlarw33k-p26m`. Removing it means
      `git filter-repo` plus a force-push.

### homebrew-tap repo

- [ ] Ruleset on `main`: block force pushes and deletion. **Don't** require PRs, or
      the tap job's direct push fails.

## 5. Run the first release

- [ ] Actions → Release → Run workflow (branch `main`)
- [ ] Approve the `homebrew-tap` environment when the tap job waits
- [ ] Check the release `v1.0.0` has these files:
  - [ ] `hue-lighter_1.0.0_macos_universal.zip`
  - [ ] `hue-lighter_1.0.0_linux_amd64.tar.gz`, `hue-lighter_1.0.0_linux_arm64.tar.gz`
  - [ ] `SHA256SUMS`
  - [ ] `CHANGELOG.md`
- [ ] `gh attestation verify hue-lighter_1.0.0_macos_universal.zip -R yveskaufmann/hue-lighter`
- [ ] `Casks/hue-lighter.rb` in `homebrew-tap` has version `1.0.0` and the checksum from `SHA256SUMS`

## 6. Install on each Mac

- [ ] **Macs with the old pkg install:** remove it first with
      `scripts/uninstall.sh Darwin --legacy-only`. It removes the LaunchDaemon and the
      `hue-lighter` system user, and copies the config to your user folder.
      Otherwise hue-lighter runs twice.
- [ ] `brew tap yveskaufmann/tap && brew trust yveskaufmann/tap`
- [ ] `brew install --cask hue-lighter`
- [ ] Start "Hue Lighter" once from `/Applications`
- [ ] If macOS asks for Keychain access to the API key, click "Always Allow".
      This happens again after each upgrade.

## Later (not needed for v1.0.0)

- [ ] Cask: replace `postflight` with `postflight_steps` (Homebrew 7 lint warning;
      the old syntax still works).
- [ ] Optional: a stable self-signed code-signing certificate, so the Keychain doesn't
      ask again after each upgrade.
- [ ] Optional: attach a `.flatpak` bundle to the release.
- [ ] Optional: include `CHANGELOG.md` in `SHA256SUMS` and the attestation.
