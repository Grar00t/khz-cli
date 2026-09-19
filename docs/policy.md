# Policy v1

`khz policy init` writes `.khz/policy.json` if it does not exist. It refuses to overwrite.

```json
{
  "schema_version": 1,
  "protected_branches": ["main", "master"],
  "receipt_required_for_mutations": true,
  "allow_color": true,
  "language": "en"
}
```

`language` is `en` or `ar`. JSON keys stay English.

`khz policy check` validates the document and reports whether the current Git branch is listed in `protected_branches`.

**This is not GitHub branch protection.** It does not configure the GitHub API, required reviews, or server-side rulesets. It is a local file KHZ reads.

`receipt_required_for_mutations` is recorded and displayed. Phase 1 does not implement a general-purpose mutation gate language.
