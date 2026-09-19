# Receipts

Receipt schema version 1 records operation metadata, redacted argv, timing, exit status, Git snapshots when observable, check evidence, policy events, artifacts, and an SHA-256 integrity hash.

## Hash rule

1. Parse the supported schema strictly.
2. Set `receipt_hash` to the empty string in memory.
3. Serialize the typed Go structure using compact `encoding/json` field order.
4. Compute SHA-256 over those bytes.
5. Store the lowercase hexadecimal digest in `receipt_hash`.

The stored JSON may be indented; verification reconstructs the canonical compact representation before hashing.

Receipt files are written through a temporary file in the destination directory and renamed into place. Existing receipt IDs are refused rather than intentionally overwritten.

## Boundary

For arbitrary commands, `side_effect_observability` is `partial`. The receipt does **not** prove absence of filesystem writes, network use, registry changes, subprocess effects, or sandbox escape. Integrity proves that the stored fields have not changed under the documented hash scheme; it does not make the recorded claims semantically true.

## V1 known-answer test vector

The repository test suite fixes one canonical receipt vector with these key values:

```text
receipt_id   20260101T000000.000000000Z-abcdef123456
khz_version  0.1.0
operation    check
command      example
exit_code    0
status       OK
termination  exit 0
observability partial
```

Its expected SHA-256 receipt hash is:

```text
422d2dc1d1b1516cd486f93af8281866f894f8f8acfe9efa83f5709d987a32d7
```

`internal/receipt/receipt_test.go` recomputes this value from the typed schema and fails if canonical serialization changes without an intentional schema decision.
