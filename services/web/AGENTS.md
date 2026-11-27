# web

The React single-page app of OpenParallax, the product's only surface (`Web` in
`_jflow/prd/index.md`). It exposes nothing: it consumes the api over HTTP and has no business logic of
its own. Its place in the product is in `_jflow/prd/architecture/services.md`.

Code, comments, identifiers and test names are in English; the texts and error messages the service
shows its users are in the product language (`language` in `_jflow/project.yaml`; see the root
`AGENTS.md`): the Spanish catalog (`src/i18n/catalogs/es.ts`) is the base, in neutral Spanish, and the
English one derives its type from it.

## Stack

| Concern | Choice | Version | Why |
|---|---|---|---|
| Language | TypeScript (strict) on Node.js 22 | 4.9.5 | — |
| UI | React with Create React App (react-scripts), react-router-dom | 19.2 / 5.0.1 / 7.13 | DEC-T05 |
| Styles | One CSS file per component, design tokens in `src/styles/tokens.css` | — | DEC-T06 |
| i18n | Own module over `Intl` (`src/i18n/`) | — | DEC-T09 |
| Google sign-in | @react-oauth/google (implicit flow) | 0.13.4 | DEC-T14 |
| Runtime config | `config.js` written when the container starts (`docker/40-runtime-config.sh`) | — | DEC-T11 |
| Unit tests | Jest (react-scripts) + Testing Library, jest-junit | 27 / 16 / 17.0.0 | — |
| End-to-end tests | Playwright in `mcr.microsoft.com/playwright:v1.65.1-noble` (`e2e/`, its own package) | 1.65.1 | — |
| Serving | nginx (alpine) in the image (`Dockerfile`, `docker/nginx.conf`) | — | — |

Chosen before the adoption; versions from `package.json`. Before adding or upgrading a dependency,
check the official documentation for the stable version and its recommended practices.

## Commands

| Command | Runs |
|---|---|
| `make test S=web` | unit (Jest, one pass, `reports/junit/web-unit.xml`) and end to end (Playwright against the product up, `reports/junit/web-e2e.xml`); there are no integration tests |
| `make lint S=web` | `npm run lint` (ESLint `react-app` with `react/jsx-no-literals` on the listed folders, `--max-warnings=0`) |
| `make typecheck S=web` | `tsc --noEmit` for the app and for `e2e/` |
| `make migrate` | n/a |
| `make dev S=web` | `npm start` with hot reload against `make dev S=api` |
| `make build S=web` | the Docker image (which runs `npm run build`) |

## Structure

- `src/App.tsx`: providers and the route table; `src/index.tsx`: entry.
- `src/pages/{page}/`: one folder per screen, with its CSS and tests.
- `src/components/{ui,layout,auth,events,voting,notifications}/`: the design system and the shared pieces.
- `src/domain/`: pure rules of the front (stages, forms, dates, ranking, notifications).
- `src/services/api.ts`, `src/config/api.ts`: the HTTP client and the mapping of the api's JSON.
- `src/context/`: session (`AuthContext`) and unread notifications.
- `src/i18n/`: catalogs and the provider.
- `e2e/`: Playwright tests (`tests/`), with its own `package.json`.
- `Dockerfile`; `docker/`: the nginx config and the runtime-config script.

## Canonical examples

| Kind | File |
|---|---|
| Screen with data, errors and retry | `src/pages/manage-event/ManageEventPage.tsx` |
| Form with validation | `src/pages/create-event/CreateEventPage.tsx` with `src/domain/eventForm.ts` |
| Component test | `src/pages/event-detail/EventDetailPage.test.tsx` |
| Pure rule with its test | `src/domain/voting.ts`, `src/domain/voting.test.ts` |
| End-to-end test | `e2e/tests/main-path.spec.ts` |

## Recurring errors

One line per mistake that showed up twice (in verification, in CI or in review).

- Every visible text goes through the catalogs (`t(...)`): a literal in JSX fails the lint, and every api error code must have its translation (`src/i18n/catalog.test.ts`).
