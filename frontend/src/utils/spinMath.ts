/**
 * Pure spin-geometry helpers (framework-free so they can be unit-tested in
 * node). The pointer sits at 3 o'clock (canvas angle 0) and slice i spans
 * [rotation + i*slice, rotation + (i+1)*slice), matching Wheel.tsx's draw.
 */

export const TAU = Math.PI * 2;

export function sliceAngle(count: number): number {
    return TAU / count;
}

/**
 * Total forward rotation that lands winnerIndex under the pointer: 6 full
 * turns plus the delta to the slice center, nudged by jitterFrac slices
 * (keep |jitterFrac| <= 0.35 to stay inside the slice).
 */
export function landingRotation(
    start: number,
    winnerIndex: number,
    count: number,
    jitterFrac: number,
): number {
    const target = TAU - (winnerIndex + 0.5 + jitterFrac) * sliceAngle(count);
    const delta = (((target - start) % TAU) + TAU) % TAU;
    return 6 * TAU + delta;
}

/** Inverse of the draw geometry: which slice sits under the pointer. */
export function winnerIndexAt(rotation: number, count: number): number {
    const finalRotation = ((rotation % TAU) + TAU) % TAU;
    const winningAngle = (TAU - finalRotation) % TAU;
    return Math.floor(winningAngle / sliceAngle(count));
}
