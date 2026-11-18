---
date: 2026-11-18
type: technical-change
service: api, web, deploy, docs
---

# Cambio técnico: el producto pasa a llamarse OpenParallax

## Requerimiento original

> Renombrar todo el proyecto (codename Telescope / Telescopio) a "OpenParallax" para llevarlo a
> producción y publicarlo como open source.

## Por qué el nombre

El paralaje mide la distancia a una estrella observándola desde varios puntos de vista.
OpenParallax hace lo mismo con las propuestas: las ordena combinando la mirada de muchos
evaluadores. El nombre conserva el origen astronómico del método (Merrifield & Saari, 2009,
asignación de tiempo de telescopio) sin atarlo a ese dominio: sirve para cualquier recurso
escaso.

## Cambios

| Área | Antes | Ahora |
|---|---|---|
| Repositorio | `gravadigital/telescope` | `gravadigital/openparallax` (renombre pendiente en GitHub) |
| Módulo Go | `github.com/gravadigital/telescopio-api` | `github.com/gravadigital/openparallax/api` |
| Paquete web | `telescopio-web` | `openparallax-web` |
| Variables de entorno | `TELESCOPIO_*` | `OPENPARALLAX_*` |
| `localStorage` | `telescopio_token`, `telescopio_user`, `telescopio_locale`, … | `openparallax_*` |
| Base local | `telescope_db` / `telescope` / `telescope_password` | `openparallax_db` / `openparallax` / `openparallax_password` |
| Base de tests | `telescope_test` | `openparallax_test` |
| Bucket y clave S3 locales | `telescope`, `GKtelescopelocal01` | `openparallax`, `GKopenparallaxlocal01` |
| Imágenes Docker | `gravadigital/telescope-{api,web}` | `gravadigital/openparallax-{api,web}` |
| Stack Compose | `telescope-local` | `openparallax-local` |
| Marca | Telescope / TELESCOPIO | OpenParallax / OPENPARALLAX |
| Emisor JWT | `telescopio-api` | `openparallax-api` |

## Qué no cambia, a propósito

- Citas y enlaces externos: el título del paper (*Telescope time without tears*) y los artículos
  sobre ESO.
- Historia: `CHANGELOG.md` (entradas anteriores), `docs/changelog/`, `docs/stories/`,
  `docs/story-plans/`, `docs/requests/` y `docs/.import-state.yaml`.
- El título de los datos de ejemplo ("Distributed Telescope Time Allocation"): la migración 026
  lo usa para encontrarlos y borrarlos.
- Las frases sobre el origen del producto ("nace del problema del tiempo de telescopio").
- El host real `api.telescope.dev.grava.io` citado en el changelog.

## Compatibilidad

Ninguna, a propósito: no había instalaciones en uso, solo desarrollo con datos de prueba.

- Las sesiones abiertas en el navegador se pierden (cambia la clave del token): hay que volver a
  entrar.
- El stack local nuevo crea volúmenes nuevos. El anterior queda huérfano:
  `docker compose -p telescope-local down -v` lo borra.
- Quien tenga un `.env` propio con `TELESCOPIO_*` tiene que renombrar las variables.

## Pendiente fuera del repo

- Renombrar el repositorio en GitHub (Settings → Repository name). GitHub redirige los enlaces y
  remotos viejos.
- Crear los repositorios `openparallax-api` y `openparallax-web` en Docker Hub antes del próximo
  push a `dev` o del próximo tag.
