# Changesets

This directory holds [changesets](https://github.com/changesets/changesets) for
the `@go-ios/sdk` TypeScript package.

Add a changeset for any user-facing change:

```
npx changeset
```

Pick a bump (patch/minor/major) and describe the change. On release, the publish
workflow (`.github/workflows/publish-typescript.yml`) runs `changeset version`
(applying pending changesets to the version + CHANGELOG) and `changeset publish`
(publishing to npm with provenance).
