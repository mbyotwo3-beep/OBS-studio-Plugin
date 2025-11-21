Release v1.0.0 (Demo/Staging build)

This release packages the OBS QR Donations plugin in demo mode.

Highlights
- Demo mode enabled by default (no real funds transferred).
- QR rotation between Liquid → Lightning → Bitcoin.
- Manage Wallet dialog available (sends are simulated in demo mode).
- CI will produce downloadable artifacts for Linux and Windows when you publish this release.

How to get binaries
1. Create and push a git tag (example below).
2. Create or publish a GitHub Release from that tag.
3. The release workflow will build the plugin and attach OS-specific artifacts to the release.

Notes
- To enable real Breez/Spark functionality you must provide the Breez SDK and build with `-DBREEZ_USE_STUB=OFF`. See `CONTRIBUTING.md` for details.

Changelog
- Initial demo release (stub simulation)
