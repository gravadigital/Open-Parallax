# OpenParallax

Evaluación distribuida entre pares para convocatorias con más propuestas que evaluadores: los
participantes ordenan propuestas de otros, un Modified Borda Count (Merrifield & Saari, 2009) arma
el ranking y el ranking oficial premia o castiga la propuesta de cada evaluador según qué tan cerca
del consenso evaluó.

## Actors

| ID | Actor | Access | Access recovery | Origin |
|---|---|---|---|---|
| U-01 | Visitante | Sin cuenta: ve los eventos públicos y sus resultados, y se inscribe en un evento con nombre y email | — | initial |
| U-02 | Participante | Cuenta con email y contraseña, o con Google; inscrito en un evento, sube su propuesta y evalúa las que le tocan | Recuperar contraseña por email (enlace de 1 hora); una cuenta creada al inscribirse define así su primera contraseña | initial |
| U-03 | Organizador | Cualquier cuenta: es el autor del evento que crea y lo gestiona | Igual que U-02 | initial |
| U-04 | Administrador | Cuenta con el rol global admin, asignado escribiendo en la base; actúa como autor en cualquier evento (solo por la api) | Igual que U-02 | initial |
| U-05 | Organizador global | Cuenta con el rol global organizer, asignado escribiendo en la base; configura la votación, genera asignaciones y recalcula resultados de cualquier evento (solo por la api) | Igual que U-02 | initial |

## Surfaces

| Surface | Platform | Actors | Viewports | Service | Origin |
|---|---|---|---|---|---|
| Web | web | U-01, U-02, U-03 | Teléfono y escritorio (diseño adaptable) | web | initial |

## Documentation

Paths are relative to `_jflow/prd/`. The context of a task is assembled with `python3 .claude/scripts/jflow.py load` (see the Product Definition spec).

| Document | Contains | Load when |
|---|---|---|
| blocks/index.md | One block per line: prefix, name, surfaces, what it covers | You need to find which block covers a feature |
| blocks/{PREFIX}.md | Requirements, failure conditions and examples of the block | The task cites an ID `{PREFIX}-NN` or `{PREFIX}-EN`, or touches that block |
| decisions.md | Product and technical decisions, with what was rejected and why | Any block, module or service is loaded (filtered by Affects) |
| parameters.md | Value and variable of each parameter | A loaded file names a parameter in bold |
| nfr.md | Non-functional requirements | A task on a surface or a service (filtered by Scope) |
| out-of-scope.md | What the product does not do | Something new is defined |
| domain/index.md | One entity per line, with its tables | You need to find an entity |
| domain/glossary.md | Domain terms | Any block is loaded |
| domain/{entity}.md | Attributes, relationships, removal, states and transitions | A loaded block names it in entities |
| domain/invariants.md | Rules that always hold | One of its entities is touched |
| architecture/services.md | Services, modules and interactions | You need to find where the code goes |
| architecture/integration.md | Persistence, communication and external integrations | The task uses the database, a channel or an external system |
| architecture/auth.md | Credentials, their use and the roles | The task authenticates or authorizes |
| architecture/guarantees.md | How each rule that always holds is kept | A rule, an invariant or a module is touched |
| architecture/testing.md | How each entry point is tested end to end: test identities and replaced external systems | Something is implemented (always when a requirement closes) |
| ../apis/ | Contracts of each service (OpenAPI, AsyncAPI) | The task exposes or consumes an operation |
| ../db-schemas/ | Database schema of each service | The task touches tables |
| ../flows/index.md | Flows between modules and services | The task crosses modules or depends on a guarantee |
