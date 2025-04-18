# Tunnelier

**Tunnelier** is a WireGuard VPN connection manager that imports `.conf` files into MongoDB and connects to random VPNs using `wg-quick`. It's built for distributed VPN orchestration and automation in modern cloud-native environments.

[![Test Tunnelier](https://github.com/vulnebify/tunnelier/actions/workflows/test.yaml/badge.svg)](https://github.com/vulnebify/tunnelier/actions/workflows/test.yaml)
[![Release Tunnelier](https://github.com/vulnebify/tunnelier/actions/workflows/release.yaml/badge.svg)](https://github.com/vulnebify/tunnelier/actions/workflows/release.yaml)

---

## ✨ Features

- 📥 Import WireGuard `.conf` files into MongoDB  
- 🎯 Connect to a random working VPN via `wg-quick`  
- 🔁 Retry failed configs until success (with `--retries`)  
- 🐳 Docker-ready & production-tested  
- ⚡ Built with Go and MongoDB  

---

## 📦 Installation

### Build locally

```bash
make build
```

### Or use Docker

```bash
docker build -t tunnelier .
```

---

## 🚀 Usage

### Local binary

```bash
./bin/tunnelier up --mongo-url=mongodb://admin:adminpassword@localhost:27017
```

### Docker

```bash
docker run --rm -it --network=host --privileged tunnelier up --mongo-url=mongodb://admin:adminpassword@localhost:27017
```

---

### Flags

| Flag                  | Description                          | Default          |
|-----------------------|--------------------------------------|------------------|
| `--mongo-url`         | MongoDB connection string            | *required*       |
| `--mongo-db`          | MongoDB database name                | `tunnelier`      |
| `--mongo-collection`  | MongoDB collection name              | `configs`        |
| `--folder`            | Folder path with `.conf` files       | `.` (import cmd) |
| `--retries`           | Number of configs to try randomly    | `3`              |

---

## 🔧 Commands

| Command             | Description                                 |
|---------------------|---------------------------------------------|
| `tunnelier import`  | Import WireGuard `.conf` files              |
| `tunnelier up`      | Connect to a random working VPN             |
| `tunnelier down`    | Bring down the current VPN                  |

---

## 🧪 Testing

Start MongoDB locally:

```bash
docker compose up -d
```

Then run tests:

```bash
go test ./cmd/tunnelier
```

---

## 📥 GitHub Release

To create a versioned release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The binary will appear under [Releases](../../releases).

---

## 📝 License

This project is licensed under the [MIT License](./LICENSE).
