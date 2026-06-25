# Changelog

All notable changes to this project are documented in this file. The format is
based on [Keep a Changelog](https://keepachangelog.com/), and this project
follows [Semantic Versioning](https://semver.org/).

## [1.0.0] — 2026-06-26

First stable release. Brings the SDK to parity with the modern Python
`medsenger_api` REST client, adds 17 endpoints, and modernizes
record/attachment handling. Contains **breaking changes** relative to the v0.5.x
line — to `AddRecord` and the agent-hooks methods.

### ⚠️ Breaking changes

#### `AddRecord` — new signature, options, and return type

The trailing `recordTime`/`params` parameters became functional options, and the
return type changed from `*int` to `int`.

```go
// Before (v0.5.x)
idPtr, err := client.AddRecord(contractID, "systolic_pressure", "120", time.Now(), nil)
record := *idPtr

// After (v1.0.0)
id, err := client.AddRecord(contractID, "systolic_pressure", 120,
    maigo.WithRecordTime(time.Now()),
)
```

Migration:

| Before                                   | After                                            |
| ---------------------------------------- | ------------------------------------------------ |
| `recordTime time.Time` positional arg    | `maigo.WithRecordTime(t)`                         |
| `params *json.Marshaler` positional arg  | `maigo.WithRecordParams(map[string]any{…})`       |
| `value string`                           | `value any` (string still accepted — see note)   |
| returns `*int`                           | returns `int`                                     |
| — (not previously possible)              | `WithRecordFiles(...)`, `WithReplace()`, `WithSyncFiles()` |

#### Agent hooks renamed and implemented

`AddHooksForCategories` / `RemoveHooksForCategories` were non-functional stubs
(they called `panic("not implemented")`). They are **removed** and replaced by
working methods that take the category names and return an error.

```go
// Before (v0.5.x) — panicked at runtime
client.AddHooksForCategories(contractID)
client.RemoveHooksForCategories(contractID)

// After (v1.0.0)
err := client.AddHooks(contractID, []string{"systolic_pressure"})
err := client.RemoveHooks(contractID, []string{"systolic_pressure"})
```

#### `Record` struct fields

- `Value` widened from `string` to `any` (source-compatible — string literals
  still work).
- `Time` is now optional (`*json.Timestamp`) and three fields were added:
  `Params`, `Files`, `Replace`. Prefer the `NewRecord(category, value, time)`
  constructor (its `value` is now `any`); set the new fields on the returned
  struct as needed.

### Deprecated

- `MessageAttachment` — was an empty, unusable struct; now a type alias of the
  new [`Attachment`](attachment.go) type, so `[]MessageAttachment` and
  `WithAttachments([]MessageAttachment{…})` keep compiling. Use `Attachment`
  going forward.

> **Note on source compatibility:** widening `value` to `any` in `AddRecord` and
> `NewRecord` is not a breaking change at call sites — `string` is assignable to
> `any`, so existing string values compile unchanged. The breaking part of
> `AddRecord` is only the replaced trailing parameters and the `*int` → `int`
> return.

### Added

- **Records:** `DeleteRecord`, `SetClassifier`; `AddRecord`/`AddRecords` now
  support `params`, `files`, `replace`, and `sync_files`; `WithGroup()` option
  for `GetRecords`.
- **Hooks:** `AddHooks`, `RemoveHooks`.
- **Tasks:** `AddTask` (with `WithTaskTargetNumber`, `WithTaskImportant`,
  `WithTaskDate`, `WithTaskActionLink`), `FinishTask`, `DeleteTask`.
- **Payments:** `RequestPayment`, `GetPayments`.
- **Orders:** `SendOrder` (with `WithOrderReceiverID`, `WithOrderParams`).
- **Files & attachments:** `GetFile`, `GetFileLink`, `GetAttachment`, `GetImage`;
  the `Attachment` type and the `PrepareFile` / `PrepareBinary` helpers.
- **Messages:** `GetMessages`.
- **Contract:** `UpdateCache`, `SetInfoMaterials`, `SetContractParam`.
- **Admin:** `NotifyAdmin`, `GetAdminClinicInfo`.
- **Auth:** `ValidateAgentJWT` plus the `ErrWrongTokenType` and `ErrNoRoles`
  sentinel errors.
- **gRPC transport (opt-in):** `Init` now accepts options; pass `WithGRPC(host)`
  to route `GetRecords`, `GetRecord`, `GetCategories`, and
  `GetAvailableCategories` through gRPC with automatic REST fallback. Adds
  `GetMultipleRecords` (batched reads), `CountRecords`, and `Client.Close`.
  Connections use the embedded Medsenger CA and are established lazily;
  `contract_id` is resolved to the gRPC `user_id` and cached. Generated protobuf
  code lives in `internal/grpc/recordspb`.
- `MedicalRecord` gained `Group`, `Params`, `AttachedFiles`, `Time`, and
  `Uploaded` fields (additive; `Time`/`Uploaded` are populated only via gRPC),
  plus the `MedicalRecordFile` type.

### Dependencies

- Added `google.golang.org/grpc` and `google.golang.org/protobuf` (for the gRPC
  transport). The `go` directive was raised to 1.26 by `go mod tidy`.

### Not included

These parts of the Python client were intentionally left out:

- Locale tagging/filtering of categories.
- Debug logging and Sentry integration.

[1.0.0]: https://github.com/tikhonp/maigo/releases/tag/v1.0.0
