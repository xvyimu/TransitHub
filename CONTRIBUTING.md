# Contributing · TransitHub

Thank you for interest in TransitHub (LLM gateway + admin console).

## Ways to contribute

- **Issues:** bugs, security-adjacent reports (no secrets in public issues), feature proposals with a clear user story.
- **Pull requests:** welcome on public GitHub; keep diffs focused. Maintainer may also **merge locally** (private-solo workflow) without a PR.

## Before a PR

1. Read [`docs/PROJECT.md`](docs/PROJECT.md) (form & stack) and [`docs/PRODUCT-LAYERS.md`](docs/PRODUCT-LAYERS.md) (product layers).
2. Do **not** change production cutover (D7) or live data paths without explicit maintainer approval.
3. Run tests for the surface you touched, for example:
   - Go: `go test` on affected packages
   - Admin UI (`web/default`): typecheck / build via package scripts
4. No secrets, `.env`, or production DB files in the commit.

## License

Contributions are under **AGPL-3.0** (see `LICENSE` and `NOTICE`). Network use and source disclosure obligations apply.

## Security

See [`SECURITY.md`](SECURITY.md). Prefer coordinated disclosure over public PoCs with live credentials.
