<nav aria-label="Breadcrumb" class="text-xs text-muted">
    <a href="/" class="hover:text-brand">Discover</a><span class="mx-2">/</span>
    {{ if .FromListing }}
        <a href="{{ .BackURL }}" class="hover:text-brand">{{ .City }}</a><span class="mx-2">/</span>
    {{ end }}
    <span>Event details</span>
</nav>

<a href="{{ .BackURL }}" class="mt-8 inline-block text-sm font-semibold text-brand hover:underline">
    &larr; {{ if .FromListing }}Back to events{{ else }}Back to search{{ end }}
</a>

<div class="mt-6 grid gap-8 lg:grid-cols-[1fr_340px] lg:items-start">
    <article>
        {{ if .Event.ImageURL }}
            <img
                class="aspect-video w-full rounded-2xl object-cover"
                src="{{ .Event.ImageURL }}"
                alt="{{ .Event.Name }}"
            />
        {{ else }}
            <div
                class="flex aspect-video w-full items-center justify-center rounded-2xl bg-linear-to-br from-[#164842] to-[#2a7a62] text-sm text-lime"
                role="img"
                aria-label="No image available"
            >
                No image available
            </div>
        {{ end }}


        <h1 class="mt-8 text-3xl leading-tight font-bold tracking-tight sm:text-4xl">
            {{ .Event.Name }}
        </h1>

        {{ if .Event.Description }}
            <div class="mt-6 border-t border-line pt-6">
                <h2 class="text-2xl font-bold tracking-tight">About this event</h2>
                <p class="mt-3 leading-7 whitespace-pre-line text-muted">
                    {{ .Event.Description }}
                </p>
            </div>
        {{ end }}
    </article>

    <aside class="rounded-2xl border border-line bg-white p-6 shadow-sm lg:sticky lg:top-6">
        <p class="text-[11px] font-bold tracking-[0.18em] text-brand uppercase">Make a plan</p>
        <h2 class="mt-1 text-2xl font-bold tracking-tight">The details</h2>

        <div class="mt-5 border-t border-line pt-4">
            <p class="text-[10px] font-bold tracking-widest text-muted uppercase">When</p>
            {{ if .Event.Date }}
                <p class="mt-1 font-semibold">{{ eventDay .Event.Date }}</p>
                {{ with eventTime .Event.Time }}
                    <p class="mt-0.5 text-xs text-muted">{{ . }}</p>
                {{ end }}
            {{ else }}
                <p class="mt-1 font-semibold">To be announced</p>
            {{ end }}
        </div>

        <div class="mt-4 border-t border-line pt-4">
            <p class="text-[10px] font-bold tracking-widest text-muted uppercase">Where</p>
            {{ $place := eventPlace .Event }}
            {{ if .Event.Venue }}<p class="mt-1 font-semibold">{{ .Event.Venue }}</p>{{ end }}
            {{ if $place }}<p class="mt-0.5 text-xs text-muted">{{ $place }}</p>{{ end }}
            {{ if not (or .Event.Venue $place) }}
                <p class="mt-1 font-semibold">To be announced</p>
            {{ end }}
        </div>

        <a
            href="/redirect/{{ .Event.ID }}"
            class="mt-6 flex items-center justify-between rounded-lg bg-brand px-5 py-4 font-semibold text-white hover:bg-brand-dark"
        >
            View Tickets <span aria-hidden="true">&nearr;</span>
        </a>
        <p class="mt-3 text-xs text-muted">You will continue to the ticket provider to buy.</p>
    </aside>
</div>
