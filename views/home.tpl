<section class="grid items-center gap-10 pt-6 pb-12 md:grid-cols-[1fr_340px]">
    <div>
        <p
            class="flex items-center gap-3 text-[11px] font-bold tracking-[0.18em] text-brand uppercase"
        >
            <span class="h-0.5 w-6 bg-brand"></span>Less scrolling. More going.
        </p>
        <h1 class="mt-6 text-4xl leading-[1.1] font-bold tracking-tight sm:text-6xl">
            A city of possibilities.<br /><span class="text-brand">Find your next one.</span>
        </h1>
        <p class="mt-6 max-w-lg text-lg leading-relaxed text-muted">
            Discover music and sports in one place.<br />Choose your city. Find something worth
            heading out for.
        </p>
        <ul class="mt-6 flex flex-wrap gap-2 text-xs text-muted">
            <li class="rounded-full border border-line bg-white px-3 py-1.5">Live music</li>
            <li class="rounded-full border border-line bg-white px-3 py-1.5">
                Sports &amp; matchdays
            </li>
            <li class="rounded-full border border-line bg-white px-3 py-1.5">One simple search</li>
        </ul>
    </div>

    <div class="hidden md:block" aria-hidden="true">
        <div
            class="relative aspect-square rotate-2 overflow-hidden rounded-3xl rounded-tr-[56px] bg-brand p-6 text-white shadow-xl shadow-brand/20"
        >
            <div
                class="absolute -right-10 bottom-8 size-56 rounded-full border border-white/15"
            ></div>
            <div
                class="absolute -right-4 bottom-14 size-44 rounded-full border border-white/15"
            ></div>
            <p class="relative text-[10px] font-bold tracking-widest uppercase">
                The city is calling
            </p>
            <p
                class="relative mt-6 text-7xl leading-[0.9] font-extrabold tracking-tighter text-lime"
            >
                GO
            </p>
            <p class="relative text-7xl leading-[0.9] font-extrabold tracking-tighter">OUT.</p>
            <p class="absolute bottom-6 left-6 text-[10px] font-bold tracking-widest uppercase">
                Good plans.<br />Great memories.
            </p>
            <span class="absolute right-6 bottom-6 text-2xl text-accent">&nearr;</span>
        </div>
    </div>
</section>

<section class="rounded-2xl border border-line bg-white p-6 shadow-sm sm:p-8">
    <div class="flex flex-wrap items-baseline gap-x-4 gap-y-1">
        <h2 class="text-2xl font-bold tracking-tight">Where are we going?</h2>
        <p class="text-sm text-muted">Start with a city, then explore what is on.</p>
    </div>

    <form id="search-form" autocomplete="off" class="mt-6">
        <label for="city-input" class="text-xs font-semibold">Choose a city</label>
        <div class="mt-2 flex flex-col gap-3 sm:flex-row">
            <div class="relative flex-1">
                <svg
                    class="pointer-events-none absolute top-1/2 left-4 size-5 -translate-y-1/2 text-brand"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    aria-hidden="true"
                >
                    <circle cx="12" cy="12" r="9"></circle>
                    <circle cx="12" cy="12" r="3"></circle>
                </svg>
                <input
                    id="city-input"
                    type="text"
                    placeholder="Search a city, e.g. Toronto"
                    autocomplete="off"
                    aria-controls="suggestions"
                    class="h-14 w-full rounded-lg border border-line bg-canvas pr-4 pl-12 placeholder:text-muted/70 focus:border-brand focus:ring-2 focus:ring-brand/20 focus:outline-none"
                />
                <ul
                    id="suggestions"
                    hidden
                    class="absolute inset-x-0 top-full z-10 mt-1 overflow-hidden rounded-lg border border-line bg-white py-1 shadow-lg"
                ></ul>
            </div>
            <button
                id="search-button"
                type="submit"
                disabled
                class="h-14 rounded-lg bg-brand px-8 font-semibold text-white hover:bg-brand-dark disabled:cursor-not-allowed disabled:opacity-60"
            >
                Explore events <span class="ml-3" aria-hidden="true">&rarr;</span>
            </button>
        </div>
        <div class="mt-2 flex flex-wrap justify-between gap-x-4 gap-y-1 text-[11px] text-muted">
            <p>Type at least 3 characters and select a suggestion.</p>
            <p>Powered by Google</p>
        </div>
    </form>

    <p
        id="search-message"
        role="status"
        hidden
        class="mt-4 rounded-lg border border-warn-line bg-warn-bg px-4 py-3 text-sm text-warn-text"
    ></p>
</section>

<section class="mt-16">
    <p class="text-[11px] font-bold tracking-[0.18em] text-brand uppercase">How it works</p>
    <h2 class="mt-2 text-3xl font-bold tracking-tight">From a city to a ticket.</h2>
    <ol class="mt-8 grid gap-8 sm:grid-cols-3">
        <li class="border-t border-line pt-6">
            <p class="text-[11px] font-bold text-brand">01</p>
            <h3 class="mt-3 text-lg font-semibold">Pick a place</h3>
            <p class="mt-1 text-sm text-muted">Find a city with autocomplete.</p>
        </li>
        <li class="border-t border-line pt-6">
            <p class="text-[11px] font-bold text-brand">02</p>
            <h3 class="mt-3 text-lg font-semibold">Find your event</h3>
            <p class="mt-1 text-sm text-muted">Browse music and sports together.</p>
        </li>
        <li class="border-t border-line pt-6">
            <p class="text-[11px] font-bold text-brand">03</p>
            <h3 class="mt-3 text-lg font-semibold">View the details</h3>
            <p class="mt-1 text-sm text-muted">Open an event, then continue to tickets.</p>
        </li>
    </ol>
</section>

{{/* home.js copies this for each suggestion; the classes live here so Tailwind finds them. */}}
<template id="suggestion-template">
    <li>
        <button
            type="button"
            class="block w-full px-4 py-3 text-left text-sm hover:bg-brand-soft focus:bg-brand-soft focus:outline-none"
        ></button>
    </li>
</template>
