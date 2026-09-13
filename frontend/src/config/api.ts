// Client for the Spinwheel HTTP API (contract v1, SPI-7 step 4).
//
// Identity: the server owns the sw_sid cookie, but browsers do not send
// SameSite=Lax cookies on cross-site fetch (Pages x Cloud Run), so the FE
// keeps a localStorage echo and sends it as session_id in spin/event
// bodies. The server adopts a valid echo (and sets the cookie too), which
// keeps funnel continuity with or without cookies.

export interface ServerItem {
    id: string;
    option: string;
    color: string;
    weight: number;
}

export interface ServerWheel {
    id: string;
    name: string;
    items: ServerItem[];
}

export interface SpinDecision {
    spin_id: string;
    winner_index: number;
    winner_item: ServerItem;
    nonce: string;
    sig: string;
    expires_at: string;
}

export type EventType = 'PAGEVIEW' | 'SPIN_START' | 'SPIN_END';

export interface AnalyticsEvent {
    type: EventType;
    wheel_id?: string;
    spin_id?: string;
    path?: string;
    sig?: string;
    client_ts?: string;
}

export interface SyncItem {
    id?: string;
    option: string;
}

const API_BASE =
    import.meta.env.VITE_API_URL?.replace(/\/+$/, '') || 'http://localhost:8080/v1';

const SESSION_KEY = 'sw_sid_echo';

// In-memory fallback when localStorage is unavailable (private mode,
// non-browser). Keeps the session stable for the page lifetime.
const memoryStore = new Map<string, string>();

function storageGet(key: string): string | null {
    try {
        const value = localStorage.getItem(key);
        if (value) return value;
    } catch {
        // fall through to memory
    }
    return memoryStore.get(key) ?? null;
}

function storageSet(key: string, value: string): void {
    memoryStore.set(key, value);
    try {
        localStorage.setItem(key, value);
    } catch {
        // memory-only session
    }
}

/** Stable anon session echo, minted once and reused across visits. */
export function getSessionId(): string {
    let id = storageGet(SESSION_KEY);
    if (!id) {
        id = crypto.randomUUID();
        storageSet(SESSION_KEY, id);
    }
    return id;
}

export class ApiError extends Error {
    readonly status: number;
    readonly code: 'wheel_changed' | 'request_failed';

    constructor(status: number, message: string) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
        this.code = message.startsWith('wheel_changed') ? 'wheel_changed' : 'request_failed';
    }
}

export function isWheelChanged(err: unknown): boolean {
    return err instanceof ApiError && err.code === 'wheel_changed';
}

async function request<T>(path: string, method: 'POST' | 'PUT', body: unknown): Promise<T> {
    let res: Response;
    try {
        res = await fetch(`${API_BASE}${path}`, {
            method,
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(body),
        });
    } catch (err) {
        throw new ApiError(0, `network error: ${err instanceof Error ? err.message : 'unreachable'}`);
    }
    if (res.status === 204) return undefined as T;
    let data: { error?: string } & Record<string, unknown> = {};
    try {
        data = (await res.json()) as typeof data;
    } catch {
        throw new ApiError(res.status, `request failed with status ${res.status}`);
    }
    if (!res.ok) {
        throw new ApiError(res.status, typeof data.error === 'string' ? data.error : `request failed with status ${res.status}`);
    }
    return data as T;
}

/** Create a server wheel from option strings (weight 1.0, color '' by default). */
export function createWheel(name: string, options: string[]): Promise<ServerWheel> {
    return request<ServerWheel>('/wheels', 'POST', {
        name,
        items: options.map((option) => ({ option })),
    });
}

/** Full-list sync: known IDs preserved, unknown minted, dropped deleted. */
export function syncWheel(wheelId: string, name: string, items: SyncItem[]): Promise<ServerWheel> {
    return request<ServerWheel>(`/wheels/${wheelId}`, 'PUT', { name, items });
}

/** Draw a server decision. Throws ApiError('wheel_changed') on stale items_hash. */
export function fetchSpin(wheelId: string, clientSeed: string, itemsHash: string): Promise<SpinDecision> {
    return request<SpinDecision>(`/wheels/${wheelId}/spin`, 'POST', {
        client_seed: clientSeed,
        items_hash: itemsHash,
        session_id: getSessionId(),
    });
}

/**
 * SHA-256 over the canonical item list — must match Go ComputeItemsHash
 * byte-for-byte: sha256(join(sorted(id|option|weight|color), "\n")) with
 * Go 'g' float formatting (JS String() agrees for sane weights).
 */
export async function computeItemsHash(items: ServerItem[]): Promise<string> {
    const lines = items
        .map((it) => `${it.id}|${it.option}|${String(it.weight)}|${it.color}`)
        .sort();
    const digest = await crypto.subtle.digest(
        'SHA-256',
        new TextEncoder().encode(lines.join('\n')),
    );
    return [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, '0')).join('');
}

/**
 * Best-effort analytics delivery: fetch with keepalive, sendBeacon
 * fallback (unload-safe), total silence on failure — analytics must never
 * break the UX. client_ts defaults to now (server uses it for idempotency).
 */
export async function postEvent(evt: AnalyticsEvent): Promise<void> {
    const url = `${API_BASE}/events`;
    const body = JSON.stringify({
        ...evt,
        session_id: getSessionId(),
        client_ts: evt.client_ts ?? new Date().toISOString(),
    });
    try {
        const res = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body,
            keepalive: true,
        });
        if (res.ok) return;
    } catch {
        // offline / blocked — try beacon below
    }
    try {
        if (typeof navigator !== 'undefined' && typeof navigator.sendBeacon === 'function') {
            navigator.sendBeacon(url, new Blob([body], { type: 'application/json' }));
        }
    } catch {
        // analytics is fire-and-forget
    }
}
