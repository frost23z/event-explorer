// Home page city search: Google autocomplete through /api/locations.
// Only the suggestions list is updated while typing; the page never reloads.
;(function () {
    var DEBOUNCE_MS = 300
    var MIN_CHARS = 3

    var form = document.getElementById('search-form')
    var input = document.getElementById('city-input')
    var list = document.getElementById('suggestions')
    var button = document.getElementById('search-button')
    var message = document.getElementById('search-message')
    var template = document.getElementById('suggestion-template')

    var sessionToken = newToken()
    var selected = null // the suggestion the visitor picked: { placeId, text }
    var timer = null
    var latestRequest = 0 // lets us ignore answers that arrive late

    // One token per search: used for typing and for the selected-city lookup.
    function newToken() {
        if (window.crypto && crypto.randomUUID) {
            return crypto.randomUUID()
        }
        return (Math.random().toString(36).slice(2) + Date.now().toString(36)).slice(0, 36)
    }

    function showMessage(text) {
        message.textContent = text
        message.hidden = false
    }

    function clearMessage() {
        message.hidden = true
    }

    function hideSuggestions() {
        list.hidden = true
        list.innerHTML = ''
    }

    function showSuggestions(suggestions) {
        list.innerHTML = ''

        suggestions.forEach(function (suggestion) {
            var item = template.content.cloneNode(true)
            var choice = item.querySelector('button')

            choice.textContent = suggestion.text
            choice.addEventListener('click', function () {
                choose(suggestion)
            })

            list.appendChild(item)
        })

        list.hidden = false
    }

    function choose(suggestion) {
        latestRequest++ // drop any answer still on its way
        clearTimeout(timer)

        selected = suggestion
        input.value = suggestion.text
        button.disabled = false
        clearMessage()
        hideSuggestions()
    }

    function fetchSuggestions(text) {
        var request = ++latestRequest
        var url =
            '/api/locations/autocomplete?input=' +
            encodeURIComponent(text) +
            '&sessionToken=' +
            encodeURIComponent(sessionToken)

        fetch(url)
            .then(function (response) {
                return response.json().then(function (data) {
                    return { ok: response.ok, data: data }
                })
            })
            .then(function (result) {
                if (request !== latestRequest) {
                    return
                }
                if (!result.ok) {
                    hideSuggestions()
                    showMessage(
                        result.data.error || 'Could not load suggestions. Please try again.'
                    )
                    return
                }
                if (result.data.suggestions.length === 0) {
                    hideSuggestions()
                    showMessage('No cities found. Try a different search.')
                    return
                }
                clearMessage()
                showSuggestions(result.data.suggestions)
            })
            .catch(function () {
                if (request === latestRequest) {
                    hideSuggestions()
                    showMessage('Could not load suggestions. Please try again.')
                }
            })
    }

    input.addEventListener('input', function () {
        // Typing again means the earlier selection no longer applies.
        selected = null
        button.disabled = true
        clearMessage()
        clearTimeout(timer)

        var text = input.value.trim()
        if (text.length < MIN_CHARS) {
            latestRequest++
            hideSuggestions()
            return
        }

        timer = setTimeout(function () {
            fetchSuggestions(text)
        }, DEBOUNCE_MS)
    })

    input.addEventListener('keydown', function (event) {
        if (event.key === 'Escape') {
            hideSuggestions()
        }
    })

    form.addEventListener('submit', function (event) {
        event.preventDefault()

        if (!selected) {
            showMessage('Please choose a city from the suggestions.')
            return
        }

        var url =
            '/api/locations/' +
            encodeURIComponent(selected.placeId) +
            '?sessionToken=' +
            encodeURIComponent(sessionToken)

        button.disabled = true

        fetch(url)
            .then(function (response) {
                return response.json().then(function (data) {
                    return { ok: response.ok, data: data }
                })
            })
            .then(function (result) {
                // The place lookup ends this search session, so the next search gets a new token.
                sessionToken = newToken()

                if (!result.ok) {
                    showMessage(result.data.error || 'Could not load that city. Please try again.')
                    button.disabled = false
                    return
                }

                window.location.href =
                    '/events?city=' +
                    encodeURIComponent(result.data.city) +
                    '&countryCode=' +
                    encodeURIComponent(result.data.countryCode)
            })
            .catch(function () {
                sessionToken = newToken()
                showMessage('Could not load that city. Please try again.')
                button.disabled = false
            })
    })
})()
