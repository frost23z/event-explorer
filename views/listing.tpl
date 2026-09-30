<nav aria-label="Breadcrumb" class="text-xs text-muted">
    <a href="/" class="hover:text-brand">Discover</a><span class="mx-2">/</span
    ><span>{{ .City }}</span>
</nav>

<div class="mt-8 flex flex-wrap items-start justify-between gap-4">
    <div>
        <p class="text-[11px] font-bold tracking-[0.18em] text-brand uppercase">
            Your city. Your next plan.
        </p>
        <h1 class="mt-4 text-4xl font-bold tracking-tight sm:text-5xl">
            What is on in <span class="text-brand">{{ .City }}</span>.
        </h1>
        <p class="mt-4 text-muted">
            Music and sports, loaded together. Find your next reason to go out.
        </p>
    </div>
    <a
        href="/"
        class="rounded-xl border border-line bg-white px-5 py-3 text-sm font-semibold shadow-sm hover:border-brand"
    >
        Change city <span class="ml-2" aria-hidden="true">&nearr;</span>
    </a>
</div>

<dl class="mt-8 border-y border-line py-5 text-sm">
    <div class="flex items-baseline gap-2">
        <dt class="text-[10px] font-bold tracking-widest text-muted uppercase">Location</dt>
        <dd>{{ .City }}, {{ .CountryCode }}</dd>
    </div>
</dl>

{{ range .Sections }}
    {{ $kind := .Kind }}
    <section class="mt-12">
        <p class="text-[11px] font-bold tracking-[0.18em] text-brand uppercase">{{ .Tagline }}</p>
        <h2 class="mt-1 flex items-center gap-3 text-3xl font-bold tracking-tight">
            {{ .Title }}
            {{ if not .Error }}
                <span class="rounded-md bg-brand-soft px-2 py-0.5 text-sm font-medium text-muted"
                    >{{ len .Events }}</span
                >
            {{ end }}
        </h2>

        {{ if .Error }}
            <div
                class="mt-6 rounded-2xl border border-dashed border-warn-line bg-warn-bg px-6 py-12 text-center"
            >
                <div
                    class="mx-auto flex size-8 items-center justify-center rounded-full bg-warn-line font-bold text-warn-text"
                >
                    !
                </div>
                <h3 class="mt-3 text-xl font-bold tracking-tight">
                    {{ .Title }} could not be loaded
                </h3>
                <p class="mt-2 text-sm text-muted">{{ .Error }}</p>
                <a
                    href="{{ $.ListingURL }}"
                    class="mt-5 inline-block text-sm font-semibold text-brand hover:underline"
                    >Try again <span aria-hidden="true">&rarr;</span></a
                >
            </div>
        {{ else if not .Events }}
            <div
                class="mt-6 rounded-2xl border border-dashed border-line bg-white px-6 py-12 text-center"
            >
                <div
                    class="mx-auto flex size-8 items-center justify-center rounded-full bg-brand-soft text-brand"
                    aria-hidden="true"
                >
                    &#9678;
                </div>
                <h3 class="mt-3 text-xl font-bold tracking-tight">
                    No {{ .Title }} events in
                    {{ $.City }}
                </h3>
                <p class="mt-2 text-sm text-muted">
                    Nothing is listed here at the moment. Try another city.
                </p>
                <a
                    href="/"
                    class="mt-5 inline-block text-sm font-semibold text-brand hover:underline"
                    >Choose another city <span aria-hidden="true">&rarr;</span></a
                >
            </div>
        {{ else }}
            <div class="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                {{ range .Events }}
                    <article
                        class="flex flex-col overflow-hidden rounded-2xl border border-line bg-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
                    >
                        {{ if .ImageURL }}
                            <img
                                class="aspect-[16/10] w-full object-cover"
                                src="{{ .ImageURL }}"
                                alt="{{ .Name }}"
                                loading="lazy"
                            />
                        {{ else if eq $kind "music" }}
                            <div
                                class="flex aspect-[16/10] w-full items-center justify-center bg-linear-to-br from-[#164842] to-[#2a7a62]"
                                role="img"
                                aria-label="No image available"
                            >
                                <svg
                                    viewBox="0 0 120 80"
                                    class="h-24 text-lime"
                                    fill="currentColor"
                                    aria-hidden="true"
                                >
                                    <rect x="20" y="32" width="9" height="16" rx="4.5"></rect>
                                    <rect x="38" y="20" width="9" height="40" rx="4.5"></rect>
                                    <rect x="56" y="8" width="9" height="64" rx="4.5"></rect>
                                    <rect x="74" y="24" width="9" height="32" rx="4.5"></rect>
                                    <rect x="92" y="34" width="9" height="12" rx="4.5"></rect>
                                </svg>
                            </div>
                        {{ else }}
                            <div
                                class="flex aspect-[16/10] w-full items-center justify-center bg-linear-to-br from-[#d9885a] to-[#c56f44]"
                                role="img"
                                aria-label="No image available"
                            >
                                <svg
                                    viewBox="0 0 80 80"
                                    class="h-24 text-[#fdebd0]"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="3"
                                    aria-hidden="true"
                                >
                                    <circle
                                        cx="40"
                                        cy="40"
                                        r="30"
                                        fill="currentColor"
                                        fill-opacity="0.25"
                                    ></circle>
                                    <path
                                        d="M40 10v60M10 40h60M19 19c14 10 28 10 42 0M19 61c14-10 28-10 42 0"
                                    ></path>
                                </svg>
                            </div>
                        {{ end }}
                        <div class="flex flex-1 flex-col px-5 pt-5 pb-4">
                            {{ if .Date }}
                                <p
                                    class="text-[11px] font-bold tracking-wider text-brand uppercase"
                                >
                                    {{ eventDay .Date }}
                                </p>
                            {{ end }}
                            <h3 class="mt-2 text-lg leading-snug font-semibold">{{ .Name }}</h3>
                            {{ if .Venue }}
                                <p class="mt-4 text-sm text-muted">{{ .Venue }}</p>
                            {{ end }}
                            <div class="mt-auto pt-5">
                                <div
                                    class="flex items-center justify-between gap-3 border-t border-line pt-3 text-xs"
                                >
                                    <span class="text-muted"
                                        >{{ with .City }}
                                            {{ . }}
                                        {{ else }}
                                            {{ $.City }}
                                        {{ end }}</span
                                    >
                                    <a
                                        href="/events/{{ .ID }}?city={{ $.City }}&amp;countryCode={{ $.CountryCode }}"
                                        class="font-semibold text-brand hover:underline"
                                        >View Details <span aria-hidden="true">&nearr;</span></a
                                    >
                                </div>
                            </div>
                        </div>
                    </article>
                {{ end }}
            </div>
        {{ end }}
    </section>
{{ end }}
