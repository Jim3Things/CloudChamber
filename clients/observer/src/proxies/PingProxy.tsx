
const minDelay = 5 * 1000 * 60  // Should be 5m in milliseconds

export class PingProxy {
    handle: number = 0
    running: boolean = false

    abortController: AbortController | undefined = undefined

    start() {
        this.abortController = new AbortController()
        this.running = true
        this.timerExpired()
    }

    cancel() {
        this.running = false
        this.issueAbort()

        if (this.handle !== 0) {
            window.clearTimeout(this.handle)
            this.handle = 0
        }
    }

    timerExpired() {
        const path = "/api/ping"
        const request = new Request(path, { method: "GET" })
        fetch(request, { signal: this.getSignal() })
            .then((resp) => {
                if (resp.ok) {
                    const expiry = resp.headers.get("Expires")
                    if (expiry !== null) {
                        const expTime = Date.parse(expiry)

                        // Always wait at least a second, just in case
                        // something went wrong with the parsing
                        const fullDelay = Math.max(expTime - Date.now() - minDelay, 1000)

                        // Preferably we refresh every 5 min, but can do so
                        // more quickly if we're near the end of the inactivity
                        // timer
                        const delay = Math.min(minDelay, fullDelay)

                        this.issueTimer(delay)
                        return
                    }
                }

                this.issueTimer(minDelay)
            })
            .catch(() => {
                this.issueTimer(minDelay)
            })
    }

    issueTimer(delay: number) {
        if (this.running) {
            this.handle = window.setTimeout(() => this.timerExpired(), delay)
        }
    }

    // Issue the abort for any outstanding operation, assuming that aborts are
    // enabled (which they should be)
    issueAbort() {
        if (this.abortController !== undefined) {
            this.abortController.abort()
        }
    }

    // Get the listener to use to sign up for notification of an abort demand.
    getSignal() : AbortSignal | undefined {
        if (this.abortController === undefined) {
            return undefined
        }

        return this.abortController.signal
    }
}