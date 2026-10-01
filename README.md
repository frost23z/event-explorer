# Event Explorer

Pick a city, browse **Music** and **Sports** events, open an event and continue to the ticket provider.
Built with Beego (Go), server-rendered HTML templates and vanilla JavaScript.

- City search: Google Places API (New): autocomplete and place details
- Events: Ticketmaster Discovery API v2

## Setup

Requirements: Go 1.26+, and API keys for both providers.

```sh
cp .env.example .env     # then fill in the two keys
go mod download
bee run
```

Open <http://localhost:8080>.

If `bee` is not installed:

```bash
go install github.com/beego/bee/v2@latest
```

then `bee run` will work.

Alternatively, since the mod file has `bee` declared as a tool:

```bash
go tool bee run
```

| Variable                | Purpose                                  |
| ----------------------- | ---------------------------------------- |
| `GOOGLE_PLACES_API_KEY` | Google Places API (New), billing-enabled |
| `TICKETMASTER_API_KEY`  | Ticketmaster Discovery API consumer key  |
| `PORT`, `RUN_MODE`      | Optional (defaults: `8080`, `dev`)       |

Both keys are read on the server only and never sent to the browser. `.env` is git-ignored.
The app exits at start-up if a key is missing.

**Live mode only.** There is no mock mode: the running app always calls Google and Ticketmaster
with your keys. The automated tests never touch the network; they use local fake servers.

### Styles

The CSS is built with Tailwind through Vite. The built `static/css/main.css` and `static/js/main.js`
are committed, so `go run .` works without Node. To change styles:

```sh
pnpm install
pnpm dev      # rebuilds on change (pnpm build for a one-off build)
```

### Tests

```sh
go test ./...                              # run everything
go test -race ./...                        # also checks the shared cache for data races
go test ./... -v 2>&1 | tee test-output.txt   # save the passing output
```

No API keys or network are needed: Google and Ticketmaster are replaced by local fake servers.

### What the automated tests cover

- Cache: hit, miss, `X-*-Cache` headers (service flag and a full `/events` request), empty list cached, expiry after 5 minutes, refresh of an expired entry, key per city/country/category, clear, concurrent access (`-race`).
- Event service: success, API failure, failed and cancelled lookups are not cached, invalid events, Music and Sports run concurrently, one failing section keeps the other.
- Ticket redirect: approved hosts pass; missing, `http`, relative, look-alike and user-info URLs are rejected.
- Request validation: autocomplete input and token, place ID, listing city (80 characters) and country, event ID.
- Both API clients: exact request sent (URL, headers, query, body), field mapping, image choice, provider errors, timeouts, cancelled requests, and the Ticketmaster key never appearing in an error.

## Routes

| Method | Route                         | What it does                                               |
| ------ | ----------------------------- | ---------------------------------------------------------- |
| GET    | `/`                           | Home page with city autocomplete (`home.tpl`)              |
| GET    | `/events?city=&countryCode=`  | Music and Sports lists for a city (`listing.tpl`)          |
| GET    | `/events/:eventId`            | Event details; direct links work (`details.tpl`)           |
| GET    | `/redirect/:eventId`          | 302 to the ticket page, after the host check (not a page)  |
| GET    | `/api/locations/autocomplete` | JSON: `input`, `sessionToken` -> suggestions and place IDs |
| GET    | `/api/locations/:placeId`     | JSON: `sessionToken` -> selected `city` and `countryCode`  |
| DELETE | `/api/cache`                  | JSON: empties the event cache, returns `{"cleared": n}`    |

Pages are rendered on the server inside `views/layout.tpl` (shared header and footer).
Errors use `views/error.tpl` with the right HTTP status.

Routes are declared with `@router` comments on the controllers. `routers/commentsRouter.go` is
generated from them by `bee run` and is committed, so run `bee run` once after adding or
changing a route.

### Clearing the cache

```sh
curl -X DELETE http://localhost:8080/api/cache
# {"cleared":2}
```

`DELETE` is used because the request removes the cached lists and repeating it is harmless.

## Code layout (MVC)

```text
routers/        route table
controllers/    read the request, call a service, render a view (no business logic)
services/       EventService (+ cache), LocationService
  googleplacesapi/   Google client: autocomplete, place details
  ticketmasterapi/   Ticketmaster client: event list, event details
models/         plain types, error type, user-facing messages
utils/          request validation, ticket URL check
views/          layout.tpl, home.tpl, listing.tpl, details.tpl, error.tpl
static/js/      home.js (autocomplete)
```

## The Go concepts

**Goroutines and channels** (`services/event.go`). `MusicAndSports` starts one goroutine per
category _before_ reading either result. Each goroutine sends a `CategoryResult` (events or an
error) on its own buffered channel, then the caller receives from both. If one category fails
the other is still shown, with a clear message in the failed section.

**Shared cache** (`services/cache.go`). One map, protected by one `sync.Mutex`, shared by every
request and by both goroutines. Key: `city|country|category`. Entries last 5 minutes; an expired
entry is dropped and fetched again. Only successful lists are cached (an empty list is a success),
so a failure is retried on the next request. Each cache hit is logged as `cache hit: <key>`.
`DELETE /api/cache` empties the map under the same mutex. `/events` also reports each section as
`X-Music-Cache: HIT|MISS` and `X-Sports-Cache: HIT|MISS`, so reuse can be seen with
`curl -i "http://localhost:8080/events?city=Toronto&countryCode=CA"`. Google responses are never cached.

**Ticket redirect** (`controllers/redirect.go`, `utils/ticketurl.go`). The destination never
comes from the visitor: the server asks Ticketmaster for the event, then `ValidateTicketURL`
requires an absolute `https` URL on an approved host (exact match or a real subdomain, no
user-info) before issuing the 302. Anything else shows "Tickets are not available" with status 400.

**Timeouts and errors.** Both clients use a 5-second timeout. Provider failures become a
`502` with a friendly message; the technical cause is only logged. The Ticketmaster key travels
in the query string, so the client strips the URL from transport errors before they can be logged.

## Google rules

- "Powered by Google" is shown next to the search.
- One session token per search: used for the autocomplete calls and the place lookup, then a new one is generated.
- Google responses are not cached.
- Only a selected suggestion can start a search.
