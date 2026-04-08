# Contributing to mh-gdpr-ai.eu S+

Thanks for your interest in contributing! This project is open to everyone.

## Quick Start

```bash
git clone https://github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus.git
cd mh-gdpr-ai.eu-s-plus
cp .env.example .env
# Add your API keys to .env
docker-compose up --build
```

- API: http://localhost:8080
- Frontend: http://localhost:3000

## How to Contribute

1. **Fork** the repo
2. **Create a branch**: `git checkout -b feat/my-feature`
3. **Make your changes**
4. **Test** that everything works locally with Docker
5. **Commit**: `git commit -m "feat: description of your change"`
6. **Push**: `git push origin feat/my-feature`
7. **Open a Pull Request**

## Commit Messages

Use conventional commits:

- `feat:` — new feature
- `fix:` — bug fix
- `docs:` — documentation
- `refactor:` — code refactoring
- `test:` — adding tests
- `security:` — security improvement

## Project Structure

```
apps/api/        ← Go backend (proxy, memory, encryption)
apps/web/        ← Next.js frontend (chat, 3D dashboard)
packages/        ← Plugins and SDKs
```

## Rules

- No API keys, tokens, or secrets in code — use environment variables
- No `any` in TypeScript
- No `panic()` in Go
- SQL queries must use prepared statements (no string concatenation)
- All inputs must be validated

## Good First Issues

Look for issues labeled [`good first issue`](https://github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/labels/good%20first%20issue) — these are beginner-friendly tasks.

## Need Help?

Open an issue or reach out:
- [LinkedIn](https://linkedin.com/in/mahadillah)
- [Email](mailto:mahadillahm4di@proton.me)

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
