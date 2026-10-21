---
name: quasar-hybrid-architecture
description: Istruisce l'agente sulle regole per sviluppare un sito ibrido in Quasar Framework (Pagine pubbliche SSG e Area riservata SPA). Usala quando crei rotte, configurazioni o componenti per distinguere contesti pubblici e privati.
---

# Architettura Ibrida Quasar: Pubblico SSG + Privato SPA

Questo progetto utilizza un'architettura ibrida gestita tramite la modalità **SSG (Static Site Generation)** di Quasar. Le pagine pubbliche devono essere staticizzate per la SEO, mentre l'area riservata deve comportarsi come una SPA tradizionale (Client-Side Rendering).

## 1. Gestione delle Rotte in `quasar.config.js`
Ogni volta che si aggiunge una rotta privata o un'intera sezione protetta, essa **deve** essere registrata nell'array `ssg.clientSideRenderingRoutes` per evitare che il compilatore generi file HTML statici vuoti o basati su stati non autenticati.

Usa sempre i pattern glob (picomatch):
- `/dashboard` per la pagina principale dell'area privata.
- `/dashboard/**` per tutte le sotto-pagine interne.
- `/auth/**` per login, registrazione e recupero password.

## 2. Regole di Sviluppo per i Componenti

### Pagine Pubbliche (SSG)
- **SEO Obligatorio:** Usa sempre il composable `useMeta` di Quasar per definire `title` e `meta description`.
- **No Browser Globals a livello Root:** Non accedere a `window`, `document` o `localStorage` direttamente nel blocco `<script setup>` fuori dai lifecycle hook, altrimenti la build SSG fallirà sul server Node.
- **Inizializzazione:** Se devi usare API del browser, racchiudile tassativamente all'interno di `onMounted()`.

### Pagine Private (SPA / CSR)
- **Sicurezza Client-Side:** Poiché queste pagine sono escluse dall'SSG, puoi accedere liberamente a `localStorage`, cookie e oggetti globali del browser.
- **Guard di Autenticazione:** Implementa i controlli di sessione preferibilmente nei Router Guard (`src/router/index.ts`) o tramite un composable di autenticazione dedicato all'interno di `onMounted()`.

## 3. Comandi di Riferimento
- Sviluppo locale: `quasar dev -m ssg`
- Compilazione produzione: `quasar build -m ssg` (Output in `dist/ssg/`)
