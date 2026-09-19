# Policy V1

`.khz/policy.json` is deliberately small:

```json
{
  "schema_version": 1,
  "protected_branches": ["main", "master"],
  "receipt_required_for_mutations": true,
  "allow_color": true,
  "language": "en"
}
```

`khz policy init` refuses to overwrite an existing policy. `khz policy check` validates schema and supported values.

`protected_branches` is local KHZ policy metadata. It is not GitHub branch protection and KHZ never claims otherwise.
