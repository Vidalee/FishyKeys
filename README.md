# FishyKeys [![Tests](https://github.com/Vidalee/FishyKeys/actions/workflows/test.yaml/badge.svg)](https://github.com/Vidalee/FishyKeys/actions/workflows/test.yaml) [![Operator Tests](https://github.com/Vidalee/FishyKeys/actions/workflows/operator-test.yml/badge.svg)](https://github.com/Vidalee/FishyKeys/actions/workflows/operator-test.yml) [![Operator E2E Tests](https://github.com/Vidalee/FishyKeys/actions/workflows/operator-test-e2e.yml/badge.svg)](https://github.com/Vidalee/FishyKeys/actions/workflows/operator-test-e2e.yml)

⚠ This project is still a work in progress, which is why there is no installation instructions ⚠

FishyKeys is a secret management system with role-based access control and Shamir’s Secret Sharing for master key management. It provides a web UI and HTTP/gRPC APIs.

Is it quite similar to HashiCorp's Vault but with less features as it's mainly a learning project.

## Features

- Master key management using Shamir’s Secret Sharing
- User and role management
- Fully unit-tested
- Role-based access control for secrets
- Secrets encrypted at rest
- Web user interface for management
- HTTP API for most actions
- gRPC API for secret access
- Kubernetes operator for managing secrets
- Authentication using JWT tokens
- Passwords stored using bcrypt hashing
- Hierarchical secret paths (secrets can be organized in folders)
- Using PostgreSQL for data storage

## Next steps

- Finish the Kubernetes operator tests
- Do some trivial CRUD endpoints (mainly for roles)
- Optimize DB requests, since we don't use an ORM let's do a few more join tables :)
- Make FishyKeys distributed using Raft protocol

## Screenshots

Keep in mind frontend was not the main focus of this project, so it is focused on functionality rather than "design".

![create maser key](./ui/public/demo/create_master_key.png)

*Create a master key, shown on first startup.*

---

![master key shares](./ui/public/demo/master_key_shares.png)

*Save the shares securely and distributes them to your team.*

---

![master key unlocking](./ui/public/demo/master_key_unlocking.png)

*Unlock the master key using the shares, this is a collaborative page : you can see others adding their shares live.*

---

![secrets dashboard](./ui/public/demo/secrets_dashboard.png)

*Dashboard showing all the secrets you have access to, with the secret `/cat/key_1` selected.*

---

![create secret](./ui/public/demo/create_secret.png)

*Create a secret, you can add metadata and set the access roles.*

