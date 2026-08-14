# Deployment Center Data Lineage

The deployment center first version is a read-only preflight report over four
ledger families:

- `ops_project_bindings`: project identity, Compose boundary, image/Commit,
  observed project health, deployment evidence, and backup records.
- `ops_vps_bindings`: VPS ownership, provider resource identity, connector
  ownership, and observed VPS identity.
- `ops_connectors`: read-only provider access and scope evidence.
- `ops_domain_bindings`: public domain ownership and observed DNS/proxy state.

## Domain Ownership Migration

Migration `127_bind_ops_domains_to_projects` adds the nullable
`project_binding_id` link. It deliberately backfills only legacy domains that
match the unique `commerce-platform` seed in the same environment. A domain
that matches a gateway alias but has no explicit project owner remains legacy
evidence and is reported as `REVIEW`; it is not silently assigned to an
arbitrary project.

The foreign key uses `ON DELETE SET NULL`. Removing a project therefore keeps
the domain ledger intact and returns the domain to the explicit-unassigned
review path instead of deleting public routing evidence.

The migration is idempotent for the column, index, and constraint creation.
The matching constraint check is scoped to `ops_domain_bindings` so an
unrelated constraint with the same name cannot suppress the required foreign
key.
