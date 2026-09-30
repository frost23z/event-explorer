<section class="mx-auto max-w-lg py-20 text-center">
    <p class="text-[11px] font-bold tracking-[0.18em] text-brand uppercase">Error {{ .Status }}</p>
    <h1 class="mt-3 text-4xl font-bold tracking-tight">We hit a snag.</h1>
    <p class="mt-4 text-muted">{{ .Message }}</p>
    <a
        href="/"
        class="mt-8 inline-block rounded-lg bg-brand px-6 py-3 font-semibold text-white hover:bg-brand-dark"
    >
        Back to search <span class="ml-2" aria-hidden="true">&rarr;</span>
    </a>
</section>
