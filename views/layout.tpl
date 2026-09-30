<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{{ if .Title }}{{ .Title }} |{{ end }}Event Explorer</title>
        <link rel="icon" type="image/svg+xml" href="/static/images/favicon.svg" />
        <link rel="stylesheet" href="/static/css/main.css" />
    </head>
    <body class="flex min-h-screen flex-col bg-canvas text-ink antialiased">
        <header class="border-b border-line bg-white">
            <div class="mx-auto flex h-[72px] max-w-6xl items-center justify-between px-5">
                <a href="/" class="flex items-center gap-2.5" aria-label="Event Explorer home">
                    <img src="/static/images/favicon.svg" alt="" class="size-9" />
                    <span class="text-xl tracking-tight"
                        ><span class="font-semibold">event</span
                        ><span class="font-light">explorer</span></span
                    >
                </a>
                <nav class="text-sm">
                    <a href="/" class="border-b-2 border-brand pb-1 font-semibold text-brand"
                        >Discover</a
                    >
                </nav>
            </div>
        </header>

        <main class="mx-auto w-full max-w-6xl flex-1 px-5 pt-8 pb-16">
            {{ .LayoutContent }}
        </main>

        <footer class="border-t border-line">
            <div
                class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-2 px-5 py-6 text-xs text-muted"
            >
                <p>Event Explorer</p>
                <p>
                    Event data from Ticketmaster <span class="mx-1">/</span> City search powered by
                    Google
                </p>
            </div>
        </footer>

        {{ if .Script }}<script src="{{ .Script }}"></script>{{ end }}
    </body>
</html>
