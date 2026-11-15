# Server Protection

Server Protection supplies the in-process protection policy, decision and workflow source used by Solovey. It includes management-connectivity guards, typed network intent construction, resource and revision checks, durable operation contracts, recovery planning, and the existing Vue contribution.

The Go module is `github.com/MalenkiySolovey/solovey-component-server-protection`. Its public packages depend on the published Solovey Host contracts. Applications inject narrow durable storage, capability, observation, Planner, artifact, and runtime inspection ports. Missing or unknown authority remains unavailable; a feature policy cannot authorize itself.

Backend adapters retain authentication, SQLite/GORM transactions, schema materialization, backup/restore, installed executable validation and concrete runtime access. The existing Network Planner coordinates network resources and receipts. The privileged Broker retains execution and protected-root authority. The canonical manifest's `brokerplugin` declaration is composed with the application's trusted Broker adapter; this module does not supply an arbitrary command executor.

The feature remains `server-protection`, delivered in-process and disabled by default. Its manifest preserves the existing settings, durable tables and schema compatibility. Disabling feature activity and explicitly removing external resources remain separate operations.

The `frontend` directory contributes the `/server-protection` route, Classic/Nexus navigation, inbound editor/status slots and English/Russian locales. It uses the application's authenticated client and contribution API through `@solovey/component-ui`. It is composed by the existing component source-input pipeline.

Build and test with Go 1.26.11 or newer:

```sh
GOWORK=off go test -race -count=1 ./...
GOWORK=off go vet ./...
GOWORK=off GOFLAGS=-mod=readonly node scripts/check-source.mjs
```

Independent frontend qualification uses the exact first-party UI SDK and lockfile in `HOST_UI_PIN.json`. Check out that commit, run `npm ci` in its `frontend` directory with the pinned Node version, then run:

```sh
node scripts/check-frontend.mjs /path/to/solovey-ui/frontend
```

CI runs source, import, provenance, race, vet and frontend checks on Linux and Windows. These checks qualify software behavior. Concrete firewall, SSH, filesystem and platform execution remain the embedding application's responsibility. Host-local consumers can reuse these packages where they supply the required ports; this repository does not implement a production Node agent.

Source history, original file identities and adaptations are recorded in `SOURCE_PROVENANCE.json`. See `NOTICE` and `LICENSE` for attribution and the GPL-3.0-only license.
